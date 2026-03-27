package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SmtpAccount SMTP 账号
type SmtpAccount struct {
	ID               uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Email            string    `gorm:"type:varchar(255);not null" json:"email"`
	PasswordEncrypted string  `gorm:"type:varchar(512);not null" json:"-"` // AES 加密存储
	SmtpHost         string    `gorm:"type:varchar(255);not null" json:"smtp_host"`
	SmtpPort         int       `gorm:"default:587" json:"smtp_port"`
	DailyLimit       int       `gorm:"default:500" json:"daily_limit"`
	DailyUsed        int       `gorm:"default:0" json:"daily_used"`
	LastResetDate    time.Time `gorm:"type:date" json:"last_reset_date"`
	Status           string    `gorm:"type:varchar(50);default:'active'" json:"status"` // active/disabled
	LastError        string    `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (s *SmtpAccount) TableName() string { return "smtp_accounts" }

func (s *SmtpAccount) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// APIKey API 密钥
type APIKey struct {
	ID         uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	Name       string     `gorm:"type:varchar(100)" json:"name"`
	KeyHash    string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`
	KeyPrefix  string     `gorm:"type:varchar(8)" json:"key_prefix"` // 显示前8位
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (a *APIKey) TableName() string { return "api_keys" }

func (a *APIKey) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// SendLog 发送日志
type SendLog struct {
	ID            uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	SmtpAccountID *uuid.UUID `gorm:"type:char(36)" json:"smtp_account_id"`
	ToEmail       string     `gorm:"type:varchar(255);not null;index" json:"to_email"`
	Subject       string     `gorm:"type:varchar(500)" json:"subject"`
	Status        string     `gorm:"type:varchar(20);not null;index" json:"status"` // success/failed
	ErrorMessage  string     `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;index" json:"created_at"`

	SmtpAccount *SmtpAccount `gorm:"foreignKey:SmtpAccountID" json:"-"`
}

func (l *SendLog) TableName() string { return "send_logs" }

func (l *SendLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}