package service

import (
	"encoding/json"
	"smtp-lite/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QueueService struct {
	db           *gorm.DB
	sendService  *SendService
	smtpService  *SmtpService
	webhookSvc   *WebhookService
	rateLimitSvc *RateLimitService
	blacklistSvc *BlacklistService
	running      bool
	stopChan     chan struct{}
}

func NewQueueService(db *gorm.DB, sendService *SendService, smtpService *SmtpService, webhookSvc *WebhookService, rateLimitSvc *RateLimitService, blacklistSvc *BlacklistService) *QueueService {
	return &QueueService{
		db:           db,
		sendService:  sendService,
		smtpService:  smtpService,
		webhookSvc:   webhookSvc,
		rateLimitSvc: rateLimitSvc,
		blacklistSvc: blacklistSvc,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动队列处理
func (s *QueueService) Start() {
	if s.running {
		return
	}
	s.running = true
	go s.processLoop()
}

// Stop 停止队列处理
func (s *QueueService) Stop() {
	if !s.running {
		return
	}
	s.running = false
	close(s.stopChan)
}

func (s *QueueService) processLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.processQueue()
		}
	}
}

func (s *QueueService) processQueue() {
	// 检查限流
	if s.rateLimitSvc != nil && !s.rateLimitSvc.CheckGlobalLimit() {
		return
	}

	// 获取待处理任务（包含定时发送）
	var tasks []model.SendQueue
	now := time.Now()

	err := s.db.Where("status = ? AND (scheduled_at IS NULL OR scheduled_at <= ?)", "pending", now).
		Order("created_at asc").
		Limit(10).
		Find(&tasks).Error

	if err != nil || len(tasks) == 0 {
		return
	}

	for _, task := range tasks {
		s.processTask(&task)
	}
}

func (s *QueueService) processTask(task *model.SendQueue) {
	// 标记为处理中
	s.db.Model(task).Update("status", "processing")

	// 检查黑名单
	if s.blacklistSvc != nil && s.blacklistSvc.IsBlacklisted(task.ToEmail) {
		s.db.Model(task).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": "Email is blacklisted",
		})
		return
	}

	// 解析CC/BCC
	var cc, bcc []string
	if task.CC != "" {
		json.Unmarshal([]byte(task.CC), &cc)
	}
	if task.BCC != "" {
		json.Unmarshal([]byte(task.BCC), &bcc)
	}

	// 解析附件
	var attachments []Attachment
	if task.Attachments != "" {
		json.Unmarshal([]byte(task.Attachments), &attachments)
	}

	// 构建发送请求
	req := &SendRequest{
		To:          task.ToEmail,
		Subject:     task.Subject,
		Body:        task.Body,
		IsHTML:      task.IsHTML,
		FromName:    task.FromName,
		CC:          cc,
		BCC:         bcc,
		Attachments: attachments,
	}

	// 发送
	resp, _ := s.sendService.Send(req)

	if resp.Success {
		now := time.Now()
		s.db.Model(task).Updates(map[string]interface{}{
			"status":          "sent",
			"sent_at":         now,
			"smtp_account_id": s.getLastUsedAccount(),
		})

		// 更新批次统计
		if task.BatchID != nil {
			s.updateBatchStats(*task.BatchID, true)
		}
	} else {
		task.RetryCount++
		if task.RetryCount >= 3 {
			s.db.Model(task).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": resp.Message,
			})

			if task.BatchID != nil {
				s.updateBatchStats(*task.BatchID, false)
			}
		} else {
			s.db.Model(task).Updates(map[string]interface{}{
				"status":      "pending",
				"retry_count": task.RetryCount,
			})
		}
	}
}

func (s *QueueService) getLastUsedAccount() *uuid.UUID {
	// 从最后的发送记录获取
	var log model.SendLog
	if err := s.db.Order("created_at desc").First(&log).Error; err == nil {
		return log.SmtpAccountID
	}
	return nil
}

// AddToQueue 添加到发送队列
func (s *QueueService) AddToQueue(req *SendRequest, scheduledAt *time.Time, batchID *uuid.UUID) (*model.SendQueue, error) {
	ccJSON, _ := json.Marshal(req.CC)
	bccJSON, _ := json.Marshal(req.BCC)
	attachmentsJSON, _ := json.Marshal(req.Attachments)

	trackID := ""
	if req.TrackEnabled {
		trackID = uuid.New().String()[:8]
	}

	task := &model.SendQueue{
		ToEmail:      req.To,
		Subject:      req.Subject,
		Body:         req.Body,
		IsHTML:       req.IsHTML,
		FromName:     req.FromName,
		CC:           string(ccJSON),
		BCC:          string(bccJSON),
		Attachments:  string(attachmentsJSON),
		TrackEnabled: req.TrackEnabled,
		TrackID:      trackID,
		Status:       "pending",
		ScheduledAt:  scheduledAt,
		BatchID:      batchID,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

// BatchSend 批量发送
func (s *QueueService) BatchSend(name string, emails []string, req *SendRequest) (*model.BatchSend, error) {
	// 创建批次
	batch := &model.BatchSend{
		Name:  name,
		Total: len(emails),
	}

	if err := s.db.Create(batch).Error; err != nil {
		return nil, err
	}

	// 过滤黑名单
	validEmails := []string{}
	for _, email := range emails {
		if s.blacklistSvc != nil && !s.blacklistSvc.IsBlacklisted(email) {
			validEmails = append(validEmails, email)
		}
	}

	// 添加到队列
	for _, email := range validEmails {
		singleReq := &SendRequest{
			To:          email,
			Subject:     req.Subject,
			Body:        req.Body,
			IsHTML:      req.IsHTML,
			FromName:    req.FromName,
			CC:          []string{},
			BCC:         []string{},
			Attachments: req.Attachments,
		}
		s.AddToQueue(singleReq, nil, &batch.ID)
	}

	// 更新批次总数
	s.db.Model(batch).Update("total", len(validEmails))

	return batch, nil
}

func (s *QueueService) updateBatchStats(batchID uuid.UUID, success bool) {
	s.db.Model(&model.BatchSend{}).Where("id = ?", batchID).
		Update("sent", gorm.Expr("sent + 1"))

	if success {
		s.db.Model(&model.BatchSend{}).Where("id = ?", batchID).
			Update("success", gorm.Expr("success + 1"))
	} else {
		s.db.Model(&model.BatchSend{}).Where("id = ?", batchID).
			Update("failed", gorm.Expr("failed + 1"))
	}

	// 检查是否完成
	var batch model.BatchSend
	if err := s.db.First(&batch, batchID).Error; err == nil {
		if batch.Sent >= batch.Total {
			now := time.Now()
			s.db.Model(&batch).Updates(map[string]interface{}{
				"status":      "completed",
				"completed_at": now,
			})
		}
	}
}

// GetBatchStatus 获取批次状态
func (s *QueueService) GetBatchStatus(id uuid.UUID) (*model.BatchSend, error) {
	var batch model.BatchSend
	err := s.db.First(&batch, id).Error
	return &batch, err
}

// GetQueueStats 获取队列统计
func (s *QueueService) GetQueueStats() (map[string]interface{}, error) {
	var pending, processing, sent, failed int64

	s.db.Model(&model.SendQueue{}).Where("status = ?", "pending").Count(&pending)
	s.db.Model(&model.SendQueue{}).Where("status = ?", "processing").Count(&processing)
	s.db.Model(&model.SendQueue{}).Where("status = ?", "sent").Count(&sent)
	s.db.Model(&model.SendQueue{}).Where("status = ?", "failed").Count(&failed)

	return map[string]interface{}{
		"pending":    pending,
		"processing": processing,
		"sent":       sent,
		"failed":     failed,
	}, nil
}

// CancelQueue 取消队列任务
func (s *QueueService) CancelQueue(id uuid.UUID) error {
	return s.db.Model(&model.SendQueue{}).Where("id = ? AND status = ?", id, "pending").
		Update("status", "cancelled").Error
}