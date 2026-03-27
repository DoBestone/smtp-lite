package service

import (
	"smtp-lite/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TemplateService struct {
	db *gorm.DB
}

func NewTemplateService(db *gorm.DB) *TemplateService {
	return &TemplateService{db: db}
}

func (s *TemplateService) List() ([]model.EmailTemplate, error) {
	var templates []model.EmailTemplate
	err := s.db.Order("created_at desc").Find(&templates).Error
	return templates, err
}

func (s *TemplateService) GetByID(id uuid.UUID) (*model.EmailTemplate, error) {
	var template model.EmailTemplate
	err := s.db.First(&template, id).Error
	return &template, err
}

func (s *TemplateService) Create(template *model.EmailTemplate) error {
	return s.db.Create(template).Error
}

func (s *TemplateService) Update(id uuid.UUID, updates map[string]interface{}) error {
	return s.db.Model(&model.EmailTemplate{}).Where("id = ?", id).Updates(updates).Error
}

func (s *TemplateService) Delete(id uuid.UUID) error {
	return s.db.Delete(&model.EmailTemplate{}, id).Error
}

func (s *TemplateService) Duplicate(id uuid.UUID) (*model.EmailTemplate, error) {
	var template model.EmailTemplate
	if err := s.db.First(&template, id).Error; err != nil {
		return nil, err
	}

	newTemplate := &model.EmailTemplate{
		Name:        template.Name + " (副本)",
		Subject:     template.Subject,
		Body:        template.Body,
		IsHTML:      template.IsHTML,
		Description: template.Description,
	}

	if err := s.db.Create(newTemplate).Error; err != nil {
		return nil, err
	}

	return newTemplate, nil
}