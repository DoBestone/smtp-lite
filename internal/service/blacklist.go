package service

import (
	"smtp-lite/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BlacklistService struct {
	db *gorm.DB
}

func NewBlacklistService(db *gorm.DB) *BlacklistService {
	return &BlacklistService{db: db}
}

func (s *BlacklistService) List() ([]model.Blacklist, error) {
	var list []model.Blacklist
	err := s.db.Order("created_at desc").Find(&list).Error
	return list, err
}

func (s *BlacklistService) Add(email, reason string) error {
	// 检查是否已存在
	var existing model.Blacklist
	if err := s.db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil // 已存在
	}

	blacklist := &model.Blacklist{
		Email:  email,
		Reason: reason,
	}

	if err := s.db.Create(blacklist).Error; err != nil {
		return err
	}

	// 更新收件人状态
	s.db.Model(&model.Recipient{}).Where("email = ?", email).
		Update("status", "blacklisted")

	return nil
}

func (s *BlacklistService) BatchAdd(emails []string, reason string) (int, error) {
	count := 0
	for _, email := range emails {
		if err := s.Add(email, reason); err == nil {
			count++
		}
	}
	return count, nil
}

func (s *BlacklistService) Remove(id uuid.UUID) error {
	var blacklist model.Blacklist
	if err := s.db.First(&blacklist, id).Error; err != nil {
		return err
	}

	email := blacklist.Email

	if err := s.db.Delete(&blacklist).Error; err != nil {
		return err
	}

	// 更新收件人状态
	s.db.Model(&model.Recipient{}).Where("email = ?", email).
		Update("status", "active")

	return nil
}

func (s *BlacklistService) RemoveByEmail(email string) error {
	result := s.db.Where("email = ?", email).Delete(&model.Blacklist{})
	if result.RowsAffected > 0 {
		s.db.Model(&model.Recipient{}).Where("email = ?", email).
			Update("status", "active")
	}
	return result.Error
}

func (s *BlacklistService) IsBlacklisted(email string) bool {
	var count int64
	s.db.Model(&model.Blacklist{}).Where("email = ?", email).Count(&count)
	return count > 0
}