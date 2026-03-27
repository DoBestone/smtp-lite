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
	PasswordEncrypted string  `gorm:"type:varchar(512);not null" json:"-"`
	SmtpHost         string    `gorm:"type:varchar(255);not null" json:"smtp_host"`
	SmtpPort         int       `gorm:"default:587" json:"smtp_port"`
	DailyLimit       int       `gorm:"default:500" json:"daily_limit"`
	DailyUsed        int       `gorm:"default:0" json:"daily_used"`
	LastResetDate    time.Time `gorm:"type:date" json:"last_reset_date"`
	Status           string    `gorm:"type:varchar(50);default:'active'" json:"status"`
	LastError        string    `gorm:"type:text" json:"last_error,omitempty"`
	Priority         int       `gorm:"default:0" json:"priority"` // 优先级，数字越大优先级越高
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
	KeyPrefix  string     `gorm:"type:varchar(8)" json:"key_prefix"`
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
	Status        string     `gorm:"type:varchar(20);not null;index" json:"status"`
	ErrorMessage  string     `gorm:"type:text" json:"error_message,omitempty"`
	Opened        bool       `gorm:"default:false" json:"opened"`           // 邮件是否被打开
	OpenedAt      *time.Time `json:"opened_at"`                            // 打开时间
	Clicked       bool       `gorm:"default:false" json:"clicked"`         // 链接是否被点击
	ClickedAt     *time.Time `json:"clicked_at"`                           // 点击时间
	TrackID       string     `gorm:"type:varchar(32);index" json:"track_id"` // 追踪ID
	BatchID       *uuid.UUID `gorm:"type:char(36);index" json:"batch_id"`   // 批次ID
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

// EmailTemplate 邮件模板
type EmailTemplate struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Subject     string    `gorm:"type:varchar(500)" json:"subject"`
	Body        string    `gorm:"type:text;not null" json:"body"`
	IsHTML      bool      `gorm:"default:false" json:"is_html"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (t *EmailTemplate) TableName() string { return "email_templates" }

func (t *EmailTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// RecipientGroup 收件人分组
type RecipientGroup struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	Count       int       `gorm:"default:0" json:"count"` // 收件人数量
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Recipients []Recipient `gorm:"foreignKey:GroupID" json:"recipients,omitempty"`
}

func (g *RecipientGroup) TableName() string { return "recipient_groups" }

func (g *RecipientGroup) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

// Recipient 收件人
type Recipient struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	GroupID   uuid.UUID `gorm:"type:char(36);not null;index" json:"group_id"`
	Email     string    `gorm:"type:varchar(255);not null" json:"email"`
	Name      string    `gorm:"type:varchar(100)" json:"name"`      // 收件人名称
	Status    string    `gorm:"default:'active'" json:"status"`     // active/blacklisted
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	Group *RecipientGroup `gorm:"foreignKey:GroupID" json:"-"`
}

func (r *Recipient) TableName() string { return "recipients" }

func (r *Recipient) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// Blacklist 黑名单
type Blacklist struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Reason    string    `gorm:"type:varchar(255)" json:"reason"` // 黑名单原因
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (b *Blacklist) TableName() string { return "blacklist" }

func (b *Blacklist) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// Webhook 回调配置
type Webhook struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	URL       string    `gorm:"type:varchar(500);not null" json:"url"`
	Secret    string    `gorm:"type:varchar(100)" json:"-"`
	Events    string    `gorm:"type:text" json:"events"` // JSON数组: ["send_success","send_failed","opened","clicked"]
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (w *Webhook) TableName() string { return "webhooks" }

func (w *Webhook) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// SendQueue 发送队列
type SendQueue struct {
	ID             uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	ToEmail        string     `gorm:"type:varchar(255);not null;index" json:"to_email"`
	Subject        string     `gorm:"type:varchar(500)" json:"subject"`
	Body           string     `gorm:"type:text;not null" json:"body"`
	IsHTML         bool       `gorm:"default:false" json:"is_html"`
	FromName       string     `gorm:"type:varchar(100)" json:"from_name"`
	CC             string     `gorm:"type:text" json:"cc"`          // JSON数组
	BCC            string     `gorm:"type:text" json:"bcc"`         // JSON数组
	Attachments    string     `gorm:"type:text" json:"attachments"` // JSON数组
	TrackEnabled   bool       `gorm:"default:false" json:"track_enabled"`
	TrackID        string     `gorm:"type:varchar(32)" json:"track_id"`
	Status         string     `gorm:"default:'pending';index" json:"status"` // pending/processing/sent/failed
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	RetryCount     int        `gorm:"default:0" json:"retry_count"`
	ScheduledAt    *time.Time `gorm:"index" json:"scheduled_at"` // 定时发送时间
	SentAt         *time.Time `json:"sent_at"`
	BatchID        *uuid.UUID `gorm:"type:char(36);index" json:"batch_id"`
	SmtpAccountID  *uuid.UUID `gorm:"type:char(36)" json:"smtp_account_id"`
	CreatedAt      time.Time  `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (q *SendQueue) TableName() string { return "send_queue" }

func (q *SendQueue) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return nil
}

// BatchSend 批次发送记录
type BatchSend struct {
	ID          uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string     `gorm:"type:varchar(100)" json:"name"`
	Total       int        `gorm:"default:0" json:"total"`       // 总数量
	Sent        int        `gorm:"default:0" json:"sent"`        // 已发送
	Success     int        `gorm:"default:0" json:"success"`     // 成功数
	Failed      int        `gorm:"default:0" json:"failed"`      // 失败数
	Status      string     `gorm:"default:'pending'" json:"status"` // pending/processing/completed/failed
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (b *BatchSend) TableName() string { return "batch_sends" }

func (b *BatchSend) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// RateLimit 发送限流记录
type RateLimit struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Key       string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"` // global 或 smtp_account_id
	Count     int       `gorm:"default:0" json:"count"`
	ResetAt   time.Time `gorm:"not null" json:"reset_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (r *RateLimit) TableName() string { return "rate_limits" }

func (r *RateLimit) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}