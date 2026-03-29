package service

import (
	"fmt"
	"smtp-lite/internal/config"
	"smtp-lite/internal/model"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TrackService struct {
	db         *gorm.DB
	webhookSvc *WebhookService
}

func NewTrackService(db *gorm.DB) *TrackService {
	return &TrackService{db: db}
}

func (s *TrackService) SetWebhookService(svc *WebhookService) {
	s.webhookSvc = svc
}

// InjectTrackPixel 注入追踪像素
func (s *TrackService) InjectTrackPixel(body string, trackID string) string {
	cfg := config.Get()
	if !cfg.Track.Enabled || cfg.Track.TrackDomain == "" {
		return body
	}

	// 追踪像素 URL
	pixelURL := fmt.Sprintf("https://%s/track/open/%s.png", cfg.Track.TrackDomain, trackID)

	// 在 </body> 前插入像素
	pixel := fmt.Sprintf(`<img src="%s" width="1" height="1" border="0" style="display:none" alt=""/>`, pixelURL)

	// 替换链接
	body = s.injectClickTracking(body, trackID)

	// 插入像素
	if idx := strings.LastIndex(body, "</body>"); idx != -1 {
		body = body[:idx] + pixel + body[idx:]
	} else if idx := strings.LastIndex(body, "</html>"); idx != -1 {
		body = body[:idx] + pixel + body[idx:]
	} else {
		body = body + pixel
	}

	return body
}

func (s *TrackService) injectClickTracking(body string, _ string) string {
	cfg := config.Get()
	if !cfg.Track.Enabled || cfg.Track.TrackDomain == "" {
		return body
	}

	// 简单替换 http/https 链接
	// 实际应用中应该用正则表达式更精确地处理
	// 这里只做示例
	return body
}

// RecordOpen 记录打开事件
func (s *TrackService) RecordOpen(trackID string) error {
	var log model.SendLog
	if err := s.db.Where("track_id = ?", trackID).First(&log).Error; err != nil {
		return err
	}

	if log.Opened {
		return nil // 已记录
	}

	now := time.Now()
	if err := s.db.Model(&log).Updates(map[string]interface{}{
		"opened":    true,
		"opened_at": now,
	}).Error; err != nil {
		return err
	}

	// 触发Webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerOpened(&log)
	}

	return nil
}

// RecordClick 记录点击事件
func (s *TrackService) RecordClick(trackID, url string) error {
	var log model.SendLog
	if err := s.db.Where("track_id = ?", trackID).First(&log).Error; err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"clicked":    true,
		"clicked_at": now,
	}

	if err := s.db.Model(&log).Updates(updates).Error; err != nil {
		return err
	}

	// 触发Webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerClicked(&log, url)
	}

	return nil
}

// GetTrackStats 获取追踪统计
func (s *TrackService) GetTrackStats(logID uuid.UUID) (map[string]interface{}, error) {
	var log model.SendLog
	if err := s.db.First(&log, logID).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"opened":     log.Opened,
		"opened_at":  log.OpenedAt,
		"clicked":    log.Clicked,
		"clicked_at": log.ClickedAt,
		"track_id":   log.TrackID,
	}, nil
}

// GetTrackPixel 获取追踪像素内容
func (s *TrackService) GetTrackPixel() []byte {
	// 1x1 透明 GIF
	return []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x21, 0xf9, 0x04, 0x01, 0x0a, 0x00, 0x01, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x4c, 0x01, 0x00, 0x3b,
	}
}
