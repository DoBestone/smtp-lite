package service

import (
	"log"
	"smtp-lite/internal/model"

	"gorm.io/gorm"
)

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// Log 记录审计日志
func (s *AuditService) Log(action, username, ip, detail string) {
	entry := &model.AuditLog{
		Action:   action,
		Username: username,
		IP:       ip,
		Detail:   detail,
	}
	if err := s.db.Create(entry).Error; err != nil {
		log.Printf("[audit] failed to write audit log: %v", err)
	}
}

// List 分页查询审计日志
func (s *AuditService) List(page, pageSize int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	s.db.Model(&model.AuditLog{}).Count(&total)
	offset := (page - 1) * pageSize
	err := s.db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

// CleanupOld 清理 90 天前的审计日志
func (s *AuditService) CleanupOld() {
	result := s.db.Where("created_at < datetime('now', '-90 days')").Delete(&model.AuditLog{})
	if result.RowsAffected > 0 {
		log.Printf("[audit] cleaned up %d old audit logs", result.RowsAffected)
	}
}
