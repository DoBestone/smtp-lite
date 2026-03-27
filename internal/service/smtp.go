package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/smtp"
	"time"

	"smtp-lite/internal/config"
	"smtp-lite/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SmtpService struct {
	db *gorm.DB
}

func NewSmtpService(db *gorm.DB) *SmtpService {
	return &SmtpService{db: db}
}

func (s *SmtpService) List() ([]model.SmtpAccount, error) {
	var accounts []model.SmtpAccount
	err := s.db.Order("created_at desc").Find(&accounts).Error
	return accounts, err
}

func (s *SmtpService) Create(account *model.SmtpAccount) error {
	// 加密密码
	encrypted, err := encryptPassword(account.PasswordEncrypted)
	if err != nil {
		return err
	}
	account.PasswordEncrypted = encrypted
	account.LastResetDate = time.Now()
	return s.db.Create(account).Error
}

func (s *SmtpService) Update(id uuid.UUID, updates map[string]interface{}) error {
	// 如果更新密码，需要加密
	if pwd, ok := updates["password"].(string); ok && pwd != "" {
		encrypted, err := encryptPassword(pwd)
		if err != nil {
			return err
		}
		updates["password_encrypted"] = encrypted
		delete(updates, "password")
	}
	return s.db.Model(&model.SmtpAccount{}).Where("id = ?", id).Updates(updates).Error
}

func (s *SmtpService) Delete(id uuid.UUID) error {
	return s.db.Delete(&model.SmtpAccount{}, id).Error
}

func (s *SmtpService) GetByID(id uuid.UUID) (*model.SmtpAccount, error) {
	var account model.SmtpAccount
	err := s.db.First(&account, id).Error
	return &account, err
}

func (s *SmtpService) Toggle(id uuid.UUID) error {
	var account model.SmtpAccount
	if err := s.db.First(&account, id).Error; err != nil {
		return err
	}
	newStatus := "active"
	if account.Status == "active" {
		newStatus = "disabled"
	}
	return s.db.Model(&account).Update("status", newStatus).Error
}

func (s *SmtpService) TestConnection(host string, port int, email, password string) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", email, password, host)
	return smtp.SendMail(addr, auth, email, []string{email}, []byte("Test connection"))
}

func (s *SmtpService) GetAvailableAccount() (*model.SmtpAccount, error) {
	var account model.SmtpAccount
	today := time.Now().Format("2006-01-02")

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 查找可用的、未超限的账号
		err := tx.Where("status = ? AND (daily_limit = 0 OR daily_used < daily_limit) AND (date(last_reset_date) = ? OR last_reset_date IS NULL)", "active", today).
			First(&account).Error
		if err != nil {
			return err
		}

		// 如果是新的一天，重置计数
		if account.LastResetDate.Format("2006-01-02") != today {
			err = tx.Model(&account).Updates(map[string]interface{}{
				"daily_used":      0,
				"last_reset_date": today,
			}).Error
			if err != nil {
				return err
			}
			account.DailyUsed = 0
		}

		return nil
	})

	return &account, err
}

func (s *SmtpService) IncrementUsed(id uuid.UUID) error {
	today := time.Now().Format("2006-01-02")
	return s.db.Model(&model.SmtpAccount{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"daily_used":      gorm.Expr("daily_used + 1"),
			"last_reset_date": today,
		}).Error
}

func (s *SmtpService) UpdateError(id uuid.UUID, errMsg string) {
	s.db.Model(&model.SmtpAccount{}).Where("id = ?", id).Update("last_error", errMsg)
}

// 加密密码
func encryptPassword(password string) (string, error) {
	key := config.Get().Encryption.Key
	if len(key) < 32 {
		key = key + "                                "[:32-len(key)]
	} else if len(key) > 32 {
		key = key[:32]
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	plaintext := []byte(password)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 解密密码
func decryptPassword(encrypted string) (string, error) {
	key := config.Get().Encryption.Key
	if len(key) < 32 {
		key = key + "                                "[:32-len(key)]
	} else if len(key) > 32 {
		key = key[:32]
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func (s *SmtpService) DecryptAccountPassword(account *model.SmtpAccount) (string, error) {
	return decryptPassword(account.PasswordEncrypted)
}