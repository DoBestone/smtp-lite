package service

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"smtp-lite/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/sha3"
	"gorm.io/gorm"
)

type APIKeyService struct {
	db *gorm.DB
}

func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

func (s *APIKeyService) List() ([]model.APIKey, error) {
	var keys []model.APIKey
	err := s.db.Order("created_at desc").Find(&keys).Error
	return keys, err
}

func (s *APIKeyService) Create(name string) (*model.APIKey, string, error) {
	// 生成 API Key: sk_xxxxx
	keyBytes := make([]byte, 24)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", err
	}
	fullKey := "sk_" + hex.EncodeToString(keyBytes)

	// 计算哈希
	hash := sha3.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hash[:])

	key := &model.APIKey{
		Name:      name,
		KeyHash:   keyHash,
		KeyPrefix: fullKey[:8], // sk_xxxxx
	}

	if err := s.db.Create(key).Error; err != nil {
		return nil, "", err
	}

	return key, fullKey, nil
}

func (s *APIKeyService) Delete(id uuid.UUID) error {
	return s.db.Delete(&model.APIKey{}, id).Error
}

func (s *APIKeyService) Reset(id uuid.UUID) (*model.APIKey, string, error) {
	var key model.APIKey
	if err := s.db.First(&key, id).Error; err != nil {
		return nil, "", err
	}

	// 生成新 Key
	keyBytes := make([]byte, 24)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", err
	}
	fullKey := "sk_" + hex.EncodeToString(keyBytes)

	hash := sha3.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hash[:])

	if err := s.db.Model(&key).Updates(map[string]interface{}{
		"key_hash":   keyHash,
		"key_prefix": fullKey[:8],
	}).Error; err != nil {
		return nil, "", err
	}

	return &key, fullKey, nil
}

func (s *APIKeyService) Validate(keyString string) bool {
	hash := sha3.Sum256([]byte(keyString))
	keyHash := hex.EncodeToString(hash[:])

	var count int64
	s.db.Model(&model.APIKey{}).Where("key_hash = ?", keyHash).Count(&count)
	return count > 0
}

func (s *APIKeyService) UpdateLastUsed(id uuid.UUID) {
	now := time.Now()
	s.db.Model(&model.APIKey{}).Where("id = ?", id).Update("last_used_at", now)
}

func (s *APIKeyService) FindByKey(keyString string) *model.APIKey {
	hash := sha3.Sum256([]byte(keyString))
	keyHash := hex.EncodeToString(hash[:])

	var key model.APIKey
	if err := s.db.Where("key_hash = ?", keyHash).First(&key).Error; err != nil {
		return nil
	}
	return &key
}
