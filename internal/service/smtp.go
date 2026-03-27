package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
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
	var client *smtp.Client
	var err error
	if port == 465 {
		conn, e := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
		if e != nil {
			return e
		}
		client, err = smtp.NewClient(conn, host)
	} else {
		client, err = smtp.Dial(addr)
		if err == nil {
			if ok, _ := client.Extension("STARTTLS"); ok {
				err = client.StartTLS(&tls.Config{ServerName: host})
			}
		}
	}
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Auth(auth)
}

func (s *SmtpService) GetAvailableAccount() (*model.SmtpAccount, error) {
	return s.GetAvailableAccountExcluding(nil)
}

// GetAvailableAccountExcluding 获取可用账号，排除指定 ID 列表（用于自动故障切换）
func (s *SmtpService) GetAvailableAccountExcluding(excludeIDs []uuid.UUID) (*model.SmtpAccount, error) {
	var account model.SmtpAccount
	today := time.Now().Format("2006-01-02")

	err := s.db.Transaction(func(tx *gorm.DB) error {
		query := tx.Where("status = ?", "active")
		if len(excludeIDs) > 0 {
			query = query.Where("id NOT IN ?", excludeIDs)
		}

		// 查找所有活跃账号，包括昨天未重置的（重置后即可用）
		var accounts []model.SmtpAccount
		if err := query.Find(&accounts).Error; err != nil {
			return err
		}

		// 筛选可用账号：未超限或需要重置的
		for i := range accounts {
			a := &accounts[i]
			needsReset := a.LastResetDate.Format("2006-01-02") != today
			if needsReset {
				// 新的一天，重置计数
				if err := tx.Model(a).Updates(map[string]interface{}{
					"daily_used":      0,
					"last_reset_date": today,
				}).Error; err != nil {
					return err
				}
				a.DailyUsed = 0
			}
			if a.DailyLimit == 0 || a.DailyUsed < a.DailyLimit {
				account = *a
				return nil
			}
		}

		return errors.New("no available SMTP account")
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

// TestSend 通过指定账号发送测试邮件到指定地址
func (s *SmtpService) TestSend(account *model.SmtpAccount, to string) error {
	password, err := s.DecryptAccountPassword(account)
	if err != nil {
		return fmt.Errorf("解密密码失败: %v", err)
	}
	subject := "SMTP Lite - 连通性测试邮件"
	body := fmt.Sprintf(
		"这是来自 SMTP Lite 的测试邮件。\n\n发件账号: %s\nSMTP 服务器: %s:%d\n\n如收到此邮件，说明您的 SMTP 配置正确。",
		account.Email, account.SmtpHost, account.SmtpPort,
	)
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		account.Email, to, subject, body,
	)
	auth := smtp.PlainAuth("", account.Email, password, account.SmtpHost)
	return sendMailAuto(account.SmtpHost, account.SmtpPort, auth, account.Email, []string{to}, []byte(msg))
}

// sendMailAuto 同时支持隐式 TLS（端口 465）和 STARTTLS（端口 587 等）
func sendMailAuto(host string, port int, auth smtp.Auth, from string, to []string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	var client *smtp.Client
	var err error
	if port == 465 {
		conn, e := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
		if e != nil {
			return e
		}
		client, err = smtp.NewClient(conn, host)
	} else {
		client, err = smtp.Dial(addr)
		if err == nil {
			if ok, _ := client.Extension("STARTTLS"); ok {
				err = client.StartTLS(&tls.Config{ServerName: host})
			}
		}
	}
	if err != nil {
		return err
	}
	defer client.Close()
	if err = client.Auth(auth); err != nil {
		return err
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
