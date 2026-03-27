package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"smtp-lite/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebhookService struct {
	db *gorm.DB
}

func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{db: db}
}

func (s *WebhookService) List() ([]model.Webhook, error) {
	var webhooks []model.Webhook
	err := s.db.Order("created_at desc").Find(&webhooks).Error
	return webhooks, err
}

func (s *WebhookService) GetByID(id uuid.UUID) (*model.Webhook, error) {
	var webhook model.Webhook
	err := s.db.First(&webhook, id).Error
	return &webhook, err
}

func (s *WebhookService) Create(webhook *model.Webhook) error {
	return s.db.Create(webhook).Error
}

func (s *WebhookService) Update(id uuid.UUID, updates map[string]interface{}) error {
	return s.db.Model(&model.Webhook{}).Where("id = ?", id).Updates(updates).Error
}

func (s *WebhookService) Delete(id uuid.UUID) error {
	return s.db.Delete(&model.Webhook{}, id).Error
}

func (s *WebhookService) Toggle(id uuid.UUID) error {
	var webhook model.Webhook
	if err := s.db.First(&webhook, id).Error; err != nil {
		return err
	}
	return s.db.Model(&webhook).Update("enabled", !webhook.Enabled).Error
}

// Trigger 触发Webhook
func (s *WebhookService) Trigger(event string, data interface{}) {
	var webhooks []model.Webhook
	s.db.Where("enabled = ?", true).Find(&webhooks)

	for _, webhook := range webhooks {
		// 检查事件是否在订阅列表中
		var events []string
		if webhook.Events != "" {
			json.Unmarshal([]byte(webhook.Events), &events)
		}

		shouldTrigger := false
		for _, e := range events {
			if e == event || e == "*" {
				shouldTrigger = true
				break
			}
		}

		if !shouldTrigger {
			continue
		}

		// 异步发送
		go s.sendWebhook(&webhook, event, data)
	}
}

type WebhookPayload struct {
	Event     string      `json:"event"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

func (s *WebhookService) sendWebhook(webhook *model.Webhook, event string, data interface{}) {
	payload := WebhookPayload{
		Event:     event,
		Data:      data,
		Timestamp: time.Now(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", webhook.URL, bytes.NewReader(body))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// 添加签名
	if webhook.Secret != "" {
		mac := hmac.New(sha256.New, []byte(webhook.Secret))
		mac.Write(body)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Signature", signature)
	}

	req.Header.Set("X-Event", event)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// TriggerSendSuccess 发送成功事件
func (s *WebhookService) TriggerSendSuccess(log *model.SendLog) {
	s.Trigger("send_success", map[string]interface{}{
		"to_email": log.ToEmail,
		"subject":  log.Subject,
		"status":   log.Status,
		"log_id":   log.ID,
	})
}

// TriggerSendFailed 发送失败事件
func (s *WebhookService) TriggerSendFailed(log *model.SendLog) {
	s.Trigger("send_failed", map[string]interface{}{
		"to_email":      log.ToEmail,
		"subject":       log.Subject,
		"status":        log.Status,
		"error_message": log.ErrorMessage,
		"log_id":        log.ID,
	})
}

// TriggerOpened 邮件打开事件
func (s *WebhookService) TriggerOpened(log *model.SendLog) {
	s.Trigger("opened", map[string]interface{}{
		"to_email": log.ToEmail,
		"subject":  log.Subject,
		"log_id":   log.ID,
	})
}

// TriggerClicked 链接点击事件
func (s *WebhookService) TriggerClicked(log *model.SendLog, url string) {
	s.Trigger("clicked", map[string]interface{}{
		"to_email": log.ToEmail,
		"subject":  log.Subject,
		"log_id":   log.ID,
		"url":      url,
	})
}