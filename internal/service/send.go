package service

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"smtp-lite/internal/model"

	"github.com/google/uuid"
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
	To       string   `json:"to" binding:"required,email"`
	CC       []string `json:"cc"`
	BCC      []string `json:"bcc"`
	Subject  string   `json:"subject" binding:"required"`
	Body     string   `json:"body" binding:"required"`
	FromName string   `json:"from_name"`
	IsHTML   bool     `json:"is_html"`
}

type SendResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	UsedSMTP string `json:"used_smtp,omitempty"`
}

func (s *SendService) Send(req *SendRequest) (*SendResponse, error) {
	const maxRetries = 3
	var triedIDs []uuid.UUID

	for attempt := 0; attempt < maxRetries; attempt++ {
		// 获取可用的 SMTP 账号（排除已失败的）
		account, err := s.smtpService.GetAvailableAccountExcluding(triedIDs)
		if err != nil {
			break
		}

		// 解密密码
		password, err := s.smtpService.DecryptAccountPassword(account)
		if err != nil {
			triedIDs = append(triedIDs, account.ID)
			continue
		}

		// 构建邮件
		from := account.Email
		if req.FromName != "" {
			from = fmt.Sprintf("%s <%s>", req.FromName, account.Email)
		}

		msg := s.buildMessage(from, req)

		// 收件人列表（To + CC + BCC 都需要投递）
		recipients := []string{req.To}
		recipients = append(recipients, req.CC...)
		recipients = append(recipients, req.BCC...)

		// 发送
		smtpAuth := smtp.PlainAuth("", account.Email, password, account.SmtpHost)
		sendErr := sendMailAuto(account.SmtpHost, account.SmtpPort, smtpAuth, account.Email, recipients, []byte(msg))

		// 记录日志
		logEntry := &model.SendLog{
			SmtpAccountID: &account.ID,
			ToEmail:       req.To,
			Subject:       req.Subject,
			Status:        "success",
		}

		if sendErr != nil {
			logEntry.Status = "failed"
			logEntry.ErrorMessage = sendErr.Error()
			s.smtpService.UpdateError(account.ID, sendErr.Error())
			s.db.Create(logEntry)
			triedIDs = append(triedIDs, account.ID)
			continue
		}

		s.db.Create(logEntry)
		s.smtpService.IncrementUsed(account.ID)
		return &SendResponse{
			Success:  true,
			Message:  "Email sent successfully",
			UsedSMTP: maskEmail(account.Email),
		}, nil
	}

	return &SendResponse{Success: false, Message: "Failed to send email after all attempts"}, nil
}

func (s *SendService) buildMessage(from string, req *SendRequest) string {
	contentType := "text/plain"
	if req.IsHTML {
		contentType = "text/html"
	}

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\n", from, req.To)
	if len(req.CC) > 0 {
		headers += fmt.Sprintf("Cc: %s\r\n", strings.Join(req.CC, ", "))
	}
	headers += fmt.Sprintf("Subject: %s\r\nContent-Type: %s; charset=UTF-8\r\n\r\n", req.Subject, contentType)
	return headers + req.Body
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
