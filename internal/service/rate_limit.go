package service

import (
	"log"
	"smtp-lite/internal/config"
	"smtp-lite/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RateLimitService struct {
	db *gorm.DB
}

func NewRateLimitService(db *gorm.DB) *RateLimitService {
	return &RateLimitService{db: db}
}

// CheckGlobalLimit 检查并递增全局限流（原子操作）
func (s *RateLimitService) CheckGlobalLimit() bool {
	cfg := config.Get()
	if !cfg.RateLimit.Enabled {
		return true
	}
	return s.checkLimit("global", cfg.RateLimit.GlobalLimit)
}

// CheckAccountLimit 检查并递增账号限流（原子操作）
func (s *RateLimitService) CheckAccountLimit(accountID uuid.UUID) bool {
	cfg := config.Get()
	if !cfg.RateLimit.Enabled {
		return true
	}
	return s.checkLimit(accountID.String(), cfg.RateLimit.AccountLimit)
}

// checkLimit 原子检查并递增限流计数，避免 TOCTOU 竞态条件
func (s *RateLimitService) checkLimit(key string, limit int) bool {
	if limit <= 0 {
		return true
	}

	now := time.Now()
	resetAt := now.Add(1 * time.Minute).Truncate(time.Minute)

	// 1. 尝试原子递增：记录未过期且未超限
	result := s.db.Model(&model.RateLimit{}).
		Where("key = ? AND reset_at > ? AND count < ?", key, now, limit).
		Update("count", gorm.Expr("count + 1"))
	if result.RowsAffected > 0 {
		return true
	}

	// 2. 尝试重置已过期的记录
	result = s.db.Model(&model.RateLimit{}).
		Where("key = ? AND reset_at <= ?", key, now).
		Updates(map[string]interface{}{
			"count":    1,
			"reset_at": resetAt,
		})
	if result.RowsAffected > 0 {
		return true
	}

	// 3. 记录不存在，创建新记录
	rl := model.RateLimit{Key: key, Count: 1, ResetAt: resetAt}
	if err := s.db.Create(&rl).Error; err != nil {
		// 并发创建冲突，重试原子递增
		result = s.db.Model(&model.RateLimit{}).
			Where("key = ? AND reset_at > ? AND count < ?", key, now, limit).
			Update("count", gorm.Expr("count + 1"))
		return result.RowsAffected > 0
	}
	return true
}

// IncrementGlobal 增加全局计数（向后兼容，checkLimit 已内含递增）
func (s *RateLimitService) IncrementGlobal() {
	// checkLimit 已在检查时原子递增，此处为空操作保持 API 兼容
}

// IncrementAccount 增加账号计数（向后兼容，checkLimit 已内含递增）
func (s *RateLimitService) IncrementAccount(accountID uuid.UUID) {
	// checkLimit 已在检查时原子递增，此处为空操作保持 API 兼容
}

// GetLimitStatus 获取限流状态
func (s *RateLimitService) GetLimitStatus() map[string]interface{} {
	cfg := config.Get()

	globalRemaining := 0

	now := time.Now()

	var globalLimit model.RateLimit
	if err := s.db.Where("key = ?", "global").First(&globalLimit).Error; err == nil {
		if now.Before(globalLimit.ResetAt) {
			globalRemaining = cfg.RateLimit.GlobalLimit - globalLimit.Count
		} else {
			globalRemaining = cfg.RateLimit.GlobalLimit
		}
	} else {
		globalRemaining = cfg.RateLimit.GlobalLimit
	}

	return map[string]interface{}{
		"enabled":           cfg.RateLimit.Enabled,
		"global_limit":      cfg.RateLimit.GlobalLimit,
		"account_limit":     cfg.RateLimit.AccountLimit,
		"global_remaining":  globalRemaining,
		"account_remaining": 0,
	}
}

// CleanupExpired 清理过期的限流记录
func (s *RateLimitService) CleanupExpired() {
	threshold := time.Now().Add(-24 * time.Hour)
	result := s.db.Where("reset_at < ?", threshold).Delete(&model.RateLimit{})
	if result.RowsAffected > 0 {
		log.Printf("[rate_limit] cleaned up %d expired records", result.RowsAffected)
	}
}
