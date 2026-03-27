package service

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"smtp-lite/internal/model"

	"gorm.io/gorm"
)

type SendService struct {
	db          *gorm.DB
	smtpService *SmtpService
}

func NewSendService(db *gorm.DB, smtpService *SmtpService) *SendService {
	return &SendService{db: db, smtpService: smtpService}
}

type SendRequest struct {
	To       string `json:"to" binding:"required,email"`
	Subject  string `json:"subject" binding:"required"`
	Body     string `json:"body" binding:"required"`
	FromName string `json:"from_name"`
	IsHTML   bool   `json:"is_html"`
}

type SendResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	UsedSMTP string `json:"used_smtp,omitempty"`
}

func (s *SendService) Send(req *SendRequest) (*SendResponse, error) {
	// 获取可用的 SMTP 账号
	account, err := s.smtpService.GetAvailableAccount()
	if err != nil {
		return &SendResponse{Success: false, Message: "No available SMTP account"}, nil
	}

	// 解密密码
	password, err := s.smtpService.DecryptAccountPassword(account)
	if err != nil {
		return &SendResponse{Success: false, Message: "Failed to decrypt password"}, nil
	}

	// 构建邮件
	from := account.Email
	if req.FromName != "" {
		from = fmt.Sprintf("%s <%s>", req.FromName, account.Email)
	}

	msg := s.buildMessage(from, req)

	// 发送
	auth := smtp.PlainAuth("", account.Email, password, account.SmtpHost)
	addr := fmt.Sprintf("%s:%d", account.SmtpHost, account.SmtpPort)

	err = smtp.SendMail(addr, auth, account.Email, []string{req.To}, []byte(msg))

	// 记录日志
	logEntry := &model.SendLog{
		SmtpAccountID: &account.ID,
		ToEmail:       req.To,
		Subject:       req.Subject,
		Status:        "success",
	}

	if err != nil {
		logEntry.Status = "failed"
		logEntry.ErrorMessage = err.Error()
		s.smtpService.UpdateError(account.ID, err.Error())
	}

	s.db.Create(logEntry)

	if err == nil {
		s.smtpService.IncrementUsed(account.ID)
		return &SendResponse{
			Success:  true,
			Message:  "Email sent successfully",
			UsedSMTP: maskEmail(account.Email),
		}, nil
	}

	return &SendResponse{Success: false, Message: fmt.Sprintf("Failed to send: %v", err)}, nil
}

func (s *SendService) buildMessage(from string, req *SendRequest) string {
	contentType := "text/plain"
	if req.IsHTML {
		contentType = "text/html"
	}

	return fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: %s; charset=UTF-8\r\n\r\n%s",
		from, req.To, req.Subject, contentType, req.Body)
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	name := parts[0]
	if len(name) <= 2 {
		return name[0:1] + "***@" + parts[1]
	}
	return name[0:2] + "***@" + parts[1]
}

func (s *SendService) Logs(page, pageSize int) ([]model.SendLog, int64, error) {
	var logs []model.SendLog
	var total int64

	s.db.Model(&model.SendLog{}).Count(&total)

	offset := (page - 1) * pageSize
	err := s.db.Preload("SmtpAccount").Order("created_at desc").
		Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}

func (s *SendService) Stats() (map[string]interface{}, error) {
	var totalSent, successCount, failedCount int64
	var todaySent int64

	s.db.Model(&model.SendLog{}).Count(&totalSent)
	s.db.Model(&model.SendLog{}).Where("status = ?", "success").Count(&successCount)
	s.db.Model(&model.SendLog{}).Where("status = ?", "failed").Count(&failedCount)

	today := time.Now().Format("2006-01-02")
	s.db.Model(&model.SendLog{}).Where("DATE(created_at) = ?", today).Count(&todaySent)

	var successRate float64
	if totalSent > 0 {
		successRate = float64(successCount) / float64(totalSent) * 100
	}

	return map[string]interface{}{
		"total_sent":   totalSent,
		"success":      successCount,
		"failed":       failedCount,
		"today_sent":   todaySent,
		"success_rate": successRate,
	}, nil
}

func (s *SendService) CreateLog(logEntry *model.SendLog) error {
	return s.db.Create(logEntry).Error
}