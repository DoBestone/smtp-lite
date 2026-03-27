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
	To          string       `json:"to" binding:"required,email"`
	CC          []string     `json:"cc"`
	BCC         []string     `json:"bcc"`
	Subject     string       `json:"subject" binding:"required"`
	Body        string       `json:"body" binding:"required"`
	FromName    string       `json:"from_name"`
	IsHTML      bool         `json:"is_html"`
	Attachments []Attachment `json:"attachments"`
	TrackEnabled bool        `json:"track_enabled"` // 是否启用追踪
}

type SendResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	UsedSMTP string `json:"used_smtp,omitempty"`
	LogID    string `json:"log_id,omitempty"`
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

	const maxRetries = 3
	var triedIDs []uuid.UUID

	for attempt := 0; attempt < maxRetries; attempt++ {
		// 检查账号限流并获取可用账号
		account, err := s.smtpService.GetAvailableAccountExcluding(triedIDs)
		if err != nil {
			break
		}

		// 检查账号限流
		if s.rateLimitSvc != nil && !s.rateLimitSvc.CheckAccountLimit(account.ID) {
			triedIDs = append(triedIDs, account.ID)
			continue
		}

		// 解密密码
		password, err := s.smtpService.DecryptAccountPassword(account)
		if err != nil {
			triedIDs = append(triedIDs, account.ID)
			continue
		}

		// 生成追踪ID
		trackID := ""
		if req.TrackEnabled {
			trackID = uuid.New().String()[:8]
		}

		// 构建邮件
		from := account.Email
		if req.FromName != "" {
			from = fmt.Sprintf("%s <%s>", req.FromName, account.Email)
		}

		msg := s.buildMessage(from, req, trackID)

		// 收件人列表
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
			TrackID:       trackID,
		}

		if sendErr != nil {
			logEntry.Status = "failed"
			logEntry.ErrorMessage = sendErr.Error()
			s.smtpService.UpdateError(account.ID, sendErr.Error())
			s.db.Create(logEntry)

			// 触发失败Webhook
			if s.webhookSvc != nil {
				s.webhookSvc.TriggerSendFailed(logEntry)
			}

			triedIDs = append(triedIDs, account.ID)
			continue
		}

		s.db.Create(logEntry)
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

	return &SendResponse{Success: false, Message: "Failed to send email after all attempts"}, nil
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

	if hasAttachment {
		boundary := fmt.Sprintf("boundary_%s", uuid.New().String()[:8])
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
		boundary := fmt.Sprintf("boundary_%s", uuid.New().String()[:8])

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
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n", att.Type))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", att.Filename))
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