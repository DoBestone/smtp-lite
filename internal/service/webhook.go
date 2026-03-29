package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"smtp-lite/internal/model"
	"strings"
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

// ValidateWebhookURL 验证 webhook URL 安全性，防止 SSRF
func ValidateWebhookURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return errors.New("invalid URL")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return errors.New("only http/https URLs are allowed")
	}
	host := parsed.Hostname()
	if host == "" {
		return errors.New("URL must have a host")
	}
	// 解析 IP 地址检查私有网段
	ips, err := net.LookupIP(host)
	if err != nil {
		// 无法解析当前不阻止，运行时还有 http client 限制
		return nil
	}
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return errors.New("webhook URL must not point to private/loopback addresses")
		}
	}
	return nil
}

// isPrivateIP 检查 IP 是否为私有 / 回环 / 链路本地地址
func isPrivateIP(ip net.IP) bool {
	privateRanges := []struct {
		network string
	}{
		{"127.0.0.0/8"},
		{"10.0.0.0/8"},
		{"172.16.0.0/12"},
		{"192.168.0.0/16"},
		{"169.254.0.0/16"},
		{"::1/128"},
		{"fc00::/7"},
		{"fe80::/10"},
	}
	for _, r := range privateRanges {
		_, cidr, _ := net.ParseCIDR(r.network)
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
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
	// SSRF 防护：运行时再次检查
	if err := ValidateWebhookURL(webhook.URL); err != nil {
		return
	}

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
