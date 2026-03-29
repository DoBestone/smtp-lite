package service

import (
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

// CheckGlobalLimit 检查全局限流
func (s *RateLimitService) CheckGlobalLimit() bool {
	cfg := config.Get()
	if !cfg.RateLimit.Enabled {
		return true
	}

	key := "global"
	return s.checkLimit(key, cfg.RateLimit.GlobalLimit)
}

// CheckAccountLimit 检查账号限流
func (s *RateLimitService) CheckAccountLimit(accountID uuid.UUID) bool {
	cfg := config.Get()
	if !cfg.RateLimit.Enabled {
		return true
	}

	key := accountID.String()
	return s.checkLimit(key, cfg.RateLimit.AccountLimit)
}

// checkLimit 检查限流（使用原子操作避免竞态条件）
func (s *RateLimitService) checkLimit(key string, limit int) bool {
	if limit <= 0 {
		return true
	}

	now := time.Now()
	resetAt := now.Add(1 * time.Minute).Truncate(time.Minute)

	var rateLimit model.RateLimit
	err := s.db.Where("key = ?", key).First(&rateLimit).Error

	if err != nil {
		// 不存在，创建新记录
		rateLimit = model.RateLimit{
			Key:     key,
			Count:   1,
			ResetAt: resetAt,
		}
		s.db.Create(&rateLimit)
		return true
	}

	// 检查是否需要重置
	if now.After(rateLimit.ResetAt) {
		s.db.Model(&rateLimit).Updates(map[string]interface{}{
			"count":    1,
			"reset_at": resetAt,
		})
		return true
	}

	// 原子操作：仅当 count < limit 时递增，避免 TOCTOU 竞态
	result := s.db.Model(&model.RateLimit{}).
		Where("key = ? AND count < ?", key, limit).
		Update("count", gorm.Expr("count + 1"))

	return result.RowsAffected > 0
}

// IncrementGlobal 增加全局计数
func (s *RateLimitService) IncrementGlobal() {
	s.increment("global")
}

// IncrementAccount 增加账号计数
func (s *RateLimitService) IncrementAccount(accountID uuid.UUID) {
	s.increment(accountID.String())
}

func (s *RateLimitService) increment(key string) {
	now := time.Now()
	resetAt := now.Add(1 * time.Minute).Truncate(time.Minute)

	var rateLimit model.RateLimit
	err := s.db.Where("key = ?", key).First(&rateLimit).Error

	if err != nil {
		rateLimit = model.RateLimit{
			Key:     key,
			Count:   1,
			ResetAt: resetAt,
		}
		s.db.Create(&rateLimit)
		return
	}

	if now.After(rateLimit.ResetAt) {
		rateLimit.Count = 1
		rateLimit.ResetAt = resetAt
		s.db.Save(&rateLimit)
		return
	}

	s.db.Model(&rateLimit).Update("count", gorm.Expr("count + 1"))
}

// GetLimitStatus 获取限流状态
func (s *RateLimitService) GetLimitStatus() map[string]interface{} {
	cfg := config.Get()

	globalRemaining := 0
	accountRemaining := 0

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
		"account_remaining": accountRemaining,
	}
}
