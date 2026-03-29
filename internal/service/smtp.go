package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
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
	if _, ok := updates["email"]; ok {
		updates["last_error"] = ""
	}
	if _, ok := updates["password_encrypted"]; ok {
		updates["last_error"] = ""
	}
	if _, ok := updates["smtp_host"]; ok {
		updates["last_error"] = ""
	}
	if _, ok := updates["smtp_port"]; ok {
		updates["last_error"] = ""
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
	auth := smtp.PlainAuth("", email, password, host)
	client, err := dialSMTPClient(host, port)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}
	return nil
}

func (s *SmtpService) GetAvailableAccount() (*model.SmtpAccount, error) {
	return s.GetAvailableAccountExcluding(nil)
}

func (s *SmtpService) ListActiveAccountsExcluding(excludeIDs []uuid.UUID) ([]model.SmtpAccount, error) {
	today := time.Now().Format("2006-01-02")
	var accounts []model.SmtpAccount

	err := s.db.Transaction(func(tx *gorm.DB) error {
		query := tx.Where("status = ?", "active")
		if len(excludeIDs) > 0 {
			query = query.Where("id NOT IN ?", excludeIDs)
		}

		if err := query.Order("priority desc").Order("daily_used asc").Order("created_at asc").Find(&accounts).Error; err != nil {
			return err
		}

		for i := range accounts {
			a := &accounts[i]
			needsReset := a.LastResetDate.Format("2006-01-02") != today
			if needsReset {
				if err := tx.Model(a).Updates(map[string]interface{}{
					"daily_used":      0,
					"last_reset_date": today,
				}).Error; err != nil {
					return err
				}
				a.DailyUsed = 0
			}
		}

		return nil
	})

	return accounts, err
}

// GetAvailableAccountExcluding 获取可用账号，排除指定 ID 列表（用于自动故障切换）
func (s *SmtpService) GetAvailableAccountExcluding(excludeIDs []uuid.UUID) (*model.SmtpAccount, error) {
	accounts, err := s.ListActiveAccountsExcluding(excludeIDs)
	if err != nil {
		return nil, err
	}
	for i := range accounts {
		if accounts[i].DailyLimit == 0 || accounts[i].DailyUsed < accounts[i].DailyLimit {
			return &accounts[i], nil
		}
	}
	return nil, errors.New("no available SMTP account")
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

func (s *SmtpService) ClearError(id uuid.UUID) {
	s.db.Model(&model.SmtpAccount{}).Where("id = ?", id).Update("last_error", "")
}

// deriveKey 使用 SHA-256 从配置密钥派生固定 32 字节 AES 密钥
func deriveKey() []byte {
	key := config.Get().Encryption.Key
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

// 加密密码
func encryptPassword(password string) (string, error) {
	keyBytes := deriveKey()

	block, err := aes.NewCipher(keyBytes)
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
	keyBytes := deriveKey()

	block, err := aes.NewCipher(keyBytes)
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

func dialSMTPClient(host string, port int) (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	if port == 465 {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
		if err != nil {
			return nil, fmt.Errorf("TLS connect to %s failed: %w", addr, err)
		}
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return nil, fmt.Errorf("create SMTP client failed: %w", err)
		}
		return client, nil
	}

	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("SMTP dial %s failed: %w", addr, err)
	}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			client.Close()
			return nil, fmt.Errorf("STARTTLS failed: %w", err)
		}
	}
	return client, nil
}

// sendMailAuto 同时支持隐式 TLS（端口 465）和 STARTTLS（端口 587 等）
func sendMailAuto(host string, port int, auth smtp.Auth, from string, to []string, msg []byte) error {
	client, err := dialSMTPClient(host, port)
	if err != nil {
		return err
	}
	defer client.Close()
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("RCPT TO %s failed: %w", addr, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("write message failed: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("finalize message failed: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("QUIT failed: %w", err)
	}
	return nil
}
