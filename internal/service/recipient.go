package service

import (
	"smtp-lite/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RecipientService struct {
	db *gorm.DB
}

func NewRecipientService(db *gorm.DB) *RecipientService {
	return &RecipientService{db: db}
}

// GroupList 获取分组列表
func (s *RecipientService) GroupList() ([]model.RecipientGroup, error) {
	var groups []model.RecipientGroup
	err := s.db.Order("created_at desc").Find(&groups).Error
	return groups, err
}

// GroupGetByID 获取分组详情
func (s *RecipientService) GroupGetByID(id uuid.UUID) (*model.RecipientGroup, error) {
	var group model.RecipientGroup
	err := s.db.Preload("Recipients").First(&group, id).Error
	return &group, err
}

// GroupCreate 创建分组
func (s *RecipientService) GroupCreate(group *model.RecipientGroup) error {
	return s.db.Create(group).Error
}

// GroupUpdate 更新分组
func (s *RecipientService) GroupUpdate(id uuid.UUID, updates map[string]interface{}) error {
	return s.db.Model(&model.RecipientGroup{}).Where("id = ?", id).Updates(updates).Error
}

// GroupDelete 删除分组（连同收件人）
func (s *RecipientService) GroupDelete(id uuid.UUID) error {
	// 先删除收件人
	s.db.Where("group_id = ?", id).Delete(&model.Recipient{})
	// 再删除分组
	return s.db.Delete(&model.RecipientGroup{}, id).Error
}

// RecipientList 获取分组内的收件人
func (s *RecipientService) RecipientList(groupID uuid.UUID) ([]model.Recipient, error) {
	var recipients []model.Recipient
	err := s.db.Where("group_id = ?", groupID).Order("created_at desc").Find(&recipients).Error
	return recipients, err
}

// RecipientCreate 创建收件人
func (s *RecipientService) RecipientCreate(recipient *model.Recipient) error {
	// 检查是否在黑名单
	var blacklist model.Blacklist
	if err := s.db.Where("email = ?", recipient.Email).First(&blacklist).Error; err == nil {
		recipient.Status = "blacklisted"
	}

	if err := s.db.Create(recipient).Error; err != nil {
		return err
	}

	// 更新分组计数
	s.db.Model(&model.RecipientGroup{}).Where("id = ?", recipient.GroupID).
		Update("count", gorm.Expr("count + 1"))

	return nil
}

// RecipientBatchCreate 批量创建收件人
func (s *RecipientService) RecipientBatchCreate(groupID uuid.UUID, emails []string) (int, int, error) {
	successCount := 0
	blacklistedCount := 0

	for _, email := range emails {
		recipient := &model.Recipient{
			GroupID: groupID,
			Email:   email,
		}

		// 检查黑名单
		var blacklist model.Blacklist
		if err := s.db.Where("email = ?", email).First(&blacklist).Error; err == nil {
			recipient.Status = "blacklisted"
			blacklistedCount++
		}

		if err := s.db.Create(recipient).Error; err != nil {
			continue
		}
		successCount++
	}

	// 更新分组计数
	s.db.Model(&model.RecipientGroup{}).Where("id = ?", groupID).
		Update("count", successCount)

	return successCount, blacklistedCount, nil
}

// RecipientUpdate 更新收件人
func (s *RecipientService) RecipientUpdate(id uuid.UUID, updates map[string]interface{}) error {
	return s.db.Model(&model.Recipient{}).Where("id = ?", id).Updates(updates).Error
}

// RecipientDelete 删除收件人
func (s *RecipientService) RecipientDelete(id uuid.UUID) error {
	var recipient model.Recipient
	if err := s.db.First(&recipient, id).Error; err != nil {
		return err
	}

	if err := s.db.Delete(&recipient).Error; err != nil {
		return err
	}

	// 更新分组计数
	s.db.Model(&model.RecipientGroup{}).Where("id = ?", recipient.GroupID).
		Update("count", gorm.Expr("count - 1"))

	return nil
}

// RecipientDeleteByEmail 根据邮箱删除收件人
func (s *RecipientService) RecipientDeleteByEmail(groupID uuid.UUID, email string) error {
	result := s.db.Where("group_id = ? AND email = ?", groupID, email).Delete(&model.Recipient{})
	if result.RowsAffected > 0 {
		s.db.Model(&model.RecipientGroup{}).Where("id = ?", groupID).
			Update("count", gorm.Expr("count - 1"))
	}
	return result.Error
}

// GetAllActiveRecipients 获取所有活跃收件人（用于群发）
func (s *RecipientService) GetAllActiveRecipients(groupID uuid.UUID) ([]string, error) {
	var emails []string
	err := s.db.Model(&model.Recipient{}).
		Where("group_id = ? AND status = ?", groupID, "active").
		Pluck("email", &emails).Error
	return emails, err
}