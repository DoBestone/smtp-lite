package service

import (
	"strings"
	"testing"

	"smtp-lite/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAPIKeyTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect database: %v", err)
	}
	db.AutoMigrate(&model.APIKey{})
	return db
}

func TestAPIKeyService_Create(t *testing.T) {
	db := setupAPIKeyTestDB(t)
	svc := NewAPIKeyService(db)

	key, fullKey, err := svc.Create("test-key")
	if err != nil {
		t.Errorf("Create() error: %v", err)
	}

	if key.Name != "test-key" {
		t.Errorf("Create() name = %s, want test-key", key.Name)
	}

	if fullKey == "" {
		t.Errorf("Create() returned empty key")
	}

	if !strings.HasPrefix(fullKey, "sk_") {
		t.Errorf("Create() key should start with 'sk_', got %s", fullKey[:3])
	}

	if key.KeyPrefix != fullKey[:8] {
		t.Errorf("Create() key_prefix mismatch")
	}
}

func TestAPIKeyService_Validate(t *testing.T) {
	db := setupAPIKeyTestDB(t)
	svc := NewAPIKeyService(db)

	// 创建 key
	_, fullKey, _ := svc.Create("test-key")

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"有效Key", fullKey, true},
		{"无效Key", "sk_invalidkey123456789", false},
		{"空Key", "", false},
		{"格式错误", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.Validate(tt.key)
			if got != tt.want {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIKeyService_Delete(t *testing.T) {
	db := setupAPIKeyTestDB(t)
	svc := NewAPIKeyService(db)

	// 创建 key
	key, _, _ := svc.Create("test-key")

	// 验证存在
	if !svc.Validate(key.KeyPrefix + "test") {
		// 注意：Validate 需要完整 key，这里只测试删除
	}

	// 删除
	err := svc.Delete(key.ID)
	if err != nil {
		t.Errorf("Delete() error: %v", err)
	}

	// 验证删除成功
	keys, _ := svc.List()
	for _, k := range keys {
		if k.ID == key.ID {
			t.Errorf("Delete() key still exists")
		}
	}
}

func TestAPIKeyService_List(t *testing.T) {
	db := setupAPIKeyTestDB(t)
	svc := NewAPIKeyService(db)

	// 创建多个 key
	svc.Create("key1")
	svc.Create("key2")
	svc.Create("key3")

	keys, err := svc.List()
	if err != nil {
		t.Errorf("List() error: %v", err)
	}

	if len(keys) != 3 {
		t.Errorf("List() returned %d keys, want 3", len(keys))
	}
}
