package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/smtp"
	"net/textproto"
	"smtp-lite/internal/model"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Attachment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"` // base64编码
	Type     string `json:"type"`    // MIME类型
}

type SendService struct {
	db           *gorm.DB
	smtpService  *SmtpService
	webhookSvc   *WebhookService
	blacklistSvc *BlacklistService
	rateLimitSvc *RateLimitService
	trackSvc     *TrackService
}

func NewSendService(db *gorm.DB, smtpService *SmtpService) *SendService {
	return &SendService{db: db, smtpService: smtpService}
}

func (s *SendService) SetWebhookService(svc *WebhookService) {
	s.webhookSvc = svc
}

func (s *SendService) SetBlacklistService(svc *BlacklistService) {
	s.blacklistSvc = svc
}

func (s *SendService) SetRateLimitService(svc *RateLimitService) {
	s.rateLimitSvc = svc
}

func (s *SendService) SetTrackService(svc *TrackService) {
	s.trackSvc = svc
}

type SendRequest struct {
	To           string       `json:"to" binding:"required,email"`
	CC           []string     `json:"cc"`
	BCC          []string     `json:"bcc"`
	Subject      string       `json:"subject" binding:"required"`
	Body         string       `json:"body" binding:"required"`
	FromName     string       `json:"from_name"`
	IsHTML       bool         `json:"is_html"`
	Attachments  []Attachment `json:"attachments"`
	TrackEnabled bool         `json:"track_enabled"` // 是否启用追踪
}

type SendResponse struct {
	Success  bool     `json:"success"`
	Message  string   `json:"message"`
	UsedSMTP string   `json:"used_smtp,omitempty"`
	LogID    string   `json:"log_id,omitempty"`
	Details  []string `json:"details,omitempty"`
}

func (s *SendService) Send(req *SendRequest) (*SendResponse, error) {
	// 检查黑名单
	if s.blacklistSvc != nil && s.blacklistSvc.IsBlacklisted(req.To) {
		return &SendResponse{Success: false, Message: "Email is blacklisted"}, nil
	}

	// 检查限流
	if s.rateLimitSvc != nil && !s.rateLimitSvc.CheckGlobalLimit() {
		return &SendResponse{Success: false, Message: "Rate limit exceeded"}, nil
	}

	accounts, err := s.smtpService.ListActiveAccountsExcluding(nil)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		message := "No active SMTP account configured"
		logEntry := &model.SendLog{ToEmail: req.To, Subject: req.Subject, Status: "failed", ErrorMessage: message}
		s.db.Create(logEntry)
		return &SendResponse{Success: false, Message: message, LogID: logEntry.ID.String()}, nil
	}

	trackID := ""
	if req.TrackEnabled {
		trackID = uuid.New().String()[:8]
	}

	recipients := []string{req.To}
	recipients = append(recipients, req.CC...)
	recipients = append(recipients, req.BCC...)

	var attempted int
	var details []string
	var rateLimited []string

	for i := range accounts {
		account := &accounts[i]
		if account.DailyLimit > 0 && account.DailyUsed >= account.DailyLimit {
			rateLimited = append(rateLimited, maskEmail(account.Email))
			continue
		}
		if s.rateLimitSvc != nil && !s.rateLimitSvc.CheckAccountLimit(account.ID) {
			rateLimited = append(rateLimited, maskEmail(account.Email))
			continue
		}

		attempted++

		password, err := s.smtpService.DecryptAccountPassword(account)
		if err != nil {
			detail := fmt.Sprintf("%s: password decrypt failed", maskEmail(account.Email))
			details = append(details, detail)
			s.smtpService.UpdateError(account.ID, detail)
			logEntry := &model.SendLog{SmtpAccountID: &account.ID, ToEmail: req.To, Subject: req.Subject, Status: "failed", TrackID: trackID, ErrorMessage: detail}
			s.db.Create(logEntry)
			if s.webhookSvc != nil {
				s.webhookSvc.TriggerSendFailed(logEntry)
			}
			continue
		}

		from := account.Email
		if req.FromName != "" {
			from = fmt.Sprintf("%s <%s>", req.FromName, account.Email)
		}

		msg := s.buildMessage(from, req, trackID)
		smtpAuth := smtp.PlainAuth("", account.Email, password, account.SmtpHost)
		sendErr := sendMailAuto(account.SmtpHost, account.SmtpPort, smtpAuth, account.Email, recipients, []byte(msg))

		logEntry := &model.SendLog{
			SmtpAccountID: &account.ID,
			ToEmail:       req.To,
			Subject:       req.Subject,
			Status:        "success",
			TrackID:       trackID,
		}

		if sendErr != nil {
			detail := fmt.Sprintf("%s: %s", maskEmail(account.Email), sendErr.Error())
			logEntry.Status = "failed"
			logEntry.ErrorMessage = detail
			details = append(details, detail)
			s.smtpService.UpdateError(account.ID, detail)
			s.db.Create(logEntry)
			if s.webhookSvc != nil {
				s.webhookSvc.TriggerSendFailed(logEntry)
			}
			continue
		}

		s.db.Create(logEntry)
		s.smtpService.ClearError(account.ID)
		s.smtpService.IncrementUsed(account.ID)

		// 更新限流计数
		if s.rateLimitSvc != nil {
			s.rateLimitSvc.IncrementGlobal()
			s.rateLimitSvc.IncrementAccount(account.ID)
		}

		// 触发成功Webhook
		if s.webhookSvc != nil {
			s.webhookSvc.TriggerSendSuccess(logEntry)
		}

		return &SendResponse{
			Success:  true,
			Message:  "Email sent successfully",
			UsedSMTP: maskEmail(account.Email),
			LogID:    logEntry.ID.String(),
		}, nil
	}

	message := "Failed to send email"
	if attempted == 0 {
		if len(rateLimited) > 0 {
			message = "No available SMTP account: all active accounts reached daily limits"
			details = append([]string{fmt.Sprintf("rate limited accounts: %s", strings.Join(rateLimited, ", "))}, details...)
		} else {
			message = "No available SMTP account"
		}
	} else {
		message = fmt.Sprintf("All %d SMTP account attempts failed", attempted)
	}

	logEntry := &model.SendLog{ToEmail: req.To, Subject: req.Subject, Status: "failed", TrackID: trackID, ErrorMessage: message}
	if err := s.db.Create(logEntry).Error; err != nil {
		return nil, err
	}

	return &SendResponse{Success: false, Message: message, LogID: logEntry.ID.String(), Details: details}, nil
}

// sanitizeFilename 清理附件文件名，防止头部注入和路径遍历
func sanitizeFilename(name string) string {
	// 移除路径分隔符
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	// 移除可能导致 MIME 头部注入的字符
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\n", "")
	// 限制长度
	if len(name) > 200 {
		name = name[:200]
	}
	if name == "" {
		name = "attachment"
	}
	return name
}

func (s *SendService) buildMessage(from string, req *SendRequest, trackID string) string {
	var buf bytes.Buffer
	headers := make(textproto.MIMEHeader)

	headers.Set("From", from)
	headers.Set("To", req.To)
	if len(req.CC) > 0 {
		headers.Set("Cc", strings.Join(req.CC, ", "))
	}
	headers.Set("Subject", req.Subject)

	hasAttachment := len(req.Attachments) > 0

	// 使用统一的 boundary 变量，避免 header 与 body 不一致
	var boundary string
	if hasAttachment {
		boundary = fmt.Sprintf("boundary_%s", uuid.New().String()[:8])
		headers.Set("MIME-Version", "1.0")
		headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=\"%s\"", boundary))
	} else {
		contentType := "text/plain; charset=UTF-8"
		if req.IsHTML {
			contentType = "text/html; charset=UTF-8"
		}
		headers.Set("Content-Type", contentType)
	}

	// 写入头部
	for k, v := range headers {
		for _, vv := range v {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, vv))
		}
	}
	buf.WriteString("\r\n")

	body := req.Body

	// 添加追踪像素
	if req.TrackEnabled && s.trackSvc != nil && req.IsHTML {
		body = s.trackSvc.InjectTrackPixel(body, trackID)
	}

	if hasAttachment {
		// 正文部分
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		contentType := "text/plain; charset=UTF-8"
		if req.IsHTML {
			contentType = "text/html; charset=UTF-8"
		}
		buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n\r\n", contentType))
		buf.WriteString(body)
		buf.WriteString("\r\n")

		// 附件部分
		for _, att := range req.Attachments {
			safeName := sanitizeFilename(att.Filename)
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n", att.Type))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", safeName))
			buf.WriteString(att.Content)
			buf.WriteString("\r\n")
		}

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		buf.WriteString(body)
	}

	return buf.String()
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
	var openedCount, clickedCount int64

	s.db.Model(&model.SendLog{}).Count(&totalSent)
	s.db.Model(&model.SendLog{}).Where("status = ?", "success").Count(&successCount)
	s.db.Model(&model.SendLog{}).Where("status = ?", "failed").Count(&failedCount)

	today := time.Now().Format("2006-01-02")
	s.db.Model(&model.SendLog{}).Where("DATE(created_at) = ?", today).Count(&todaySent)

	s.db.Model(&model.SendLog{}).Where("opened = ?", true).Count(&openedCount)
	s.db.Model(&model.SendLog{}).Where("clicked = ?", true).Count(&clickedCount)

	var successRate, openRate, clickRate float64
	if totalSent > 0 {
		successRate = float64(successCount) / float64(totalSent) * 100
	}
	if successCount > 0 {
		openRate = float64(openedCount) / float64(successCount) * 100
		clickRate = float64(clickedCount) / float64(successCount) * 100
	}

	return map[string]interface{}{
		"total_sent":   totalSent,
		"success":      successCount,
		"failed":       failedCount,
		"today_sent":   todaySent,
		"success_rate": successRate,
		"opened":       openedCount,
		"clicked":      clickedCount,
		"open_rate":    openRate,
		"click_rate":   clickRate,
	}, nil
}

func (s *SendService) CreateLog(logEntry *model.SendLog) error {
	return s.db.Create(logEntry).Error
}

// ExportLogs 导出日志为CSV
func (s *SendService) ExportLogs() (string, error) {
	var logs []model.SendLog
	s.db.Order("created_at desc").Limit(10000).Find(&logs)

	var buf bytes.Buffer
	buf.WriteString("ID,To,Subject,Status,Error,Opened,Clicked,CreatedAt\n")

	for _, log := range logs {
		buf.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%v,%v,%s\n",
			log.ID.String(),
			log.ToEmail,
			strings.ReplaceAll(log.Subject, ",", " "),
			log.Status,
			strings.ReplaceAll(log.ErrorMessage, ",", " "),
			log.Opened,
			log.Clicked,
			log.CreatedAt.Format("2006-01-02 15:04:05"),
		))
	}

	return buf.String(), nil
}

// ExportLogsJSON 导出日志为JSON
func (s *SendService) ExportLogsJSON() ([]byte, error) {
	var logs []model.SendLog
	s.db.Order("created_at desc").Limit(10000).Find(&logs)
	return json.Marshal(logs)
}
