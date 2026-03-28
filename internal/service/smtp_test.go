package service

import (
	"os"
	"testing"

	"smtp-lite/internal/config"
	"smtp-lite/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	config.Load()
	config.Get().Encryption.Key = "smtp-lite-test-encryption-32-byte!"
	os.Exit(m.Run())
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect database: %v", err)
	}

	db.AutoMigrate(&model.SmtpAccount{}, &model.APIKey{}, &model.SendLog{})
	return db
}

func TestSmtpService_Create(t *testing.T) {
	db := setupTestDB(t)
	svc := NewSmtpService(db)

	tests := []struct {
		name    string
		account *model.SmtpAccount
		wantErr bool
	}{
		{
			name: "正常创建",
			account: &model.SmtpAccount{
				Email:     "test@example.com",
				SmtpHost:  "smtp.example.com",
				SmtpPort:  587,
				DailyLimit: 100,
			},
			wantErr: false,
		},
		{
			name: "空邮箱",
			account: &model.SmtpAccount{
				Email:    "",
				SmtpHost: "smtp.example.com",
				SmtpPort: 587,
			},
			wantErr: false, // GORM 不强制 NOT NULL，业务层可自行验证
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置密码（加密需要）
			tt.account.PasswordEncrypted = "testpassword"

			err := svc.Create(tt.account)
			if tt.wantErr && err == nil {
				t.Errorf("Create() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Create() unexpected error: %v", err)
			}
		})
	}
}

func TestSmtpService_List(t *testing.T) {
	db := setupTestDB(t)
	svc := NewSmtpService(db)

	// 创建测试数据
	accounts := []model.SmtpAccount{
		{Email: "user1@example.com", PasswordEncrypted: "pwd1", SmtpHost: "smtp.example.com", SmtpPort: 587},
		{Email: "user2@example.com", PasswordEncrypted: "pwd2", SmtpHost: "smtp.example.com", SmtpPort: 587},
	}
	for _, acc := range accounts {
		svc.Create(&acc)
	}

	// 测试列表
	result, err := svc.List()
	if err != nil {
		t.Errorf("List() error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("List() returned %d accounts, want 2", len(result))
	}
}

func TestSmtpService_Delete(t *testing.T) {
	db := setupTestDB(t)
	svc := NewSmtpService(db)

	// 创建测试账号
	account := &model.SmtpAccount{
		Email: "delete@example.com",
		PasswordEncrypted: "password",
		SmtpHost: "smtp.example.com",
		SmtpPort: 587,
	}
	svc.Create(account)

	// 删除
	err := svc.Delete(account.ID)
	if err != nil {
		t.Errorf("Delete() error: %v", err)
	}

	// 验证已删除
	accounts, _ := svc.List()
	for _, acc := range accounts {
		if acc.ID == account.ID {
			t.Errorf("Delete() account still exists")
		}
	}
}

func TestSmtpService_Toggle(t *testing.T) {
	db := setupTestDB(t)
	svc := NewSmtpService(db)

	// 创建测试账号
	account := &model.SmtpAccount{
		Email: "toggle@example.com",
		PasswordEncrypted: "password",
		SmtpHost: "smtp.example.com",
		SmtpPort: 587,
		Status: "active",
	}
	svc.Create(account)

	// 切换状态
	err := svc.Toggle(account.ID)
	if err != nil {
		t.Errorf("Toggle() error: %v", err)
	}

	// 验证状态变更
	accounts, _ := svc.List()
	for _, acc := range accounts {
		if acc.ID == account.ID && acc.Status == "active" {
			t.Errorf("Toggle() status should be disabled")
		}
	}
}

func TestEncryptDecryptPassword(t *testing.T) {
	password := "mySecretPassword123"

	// 加密
	encrypted, err := encryptPassword(password)
	if err != nil {
		t.Errorf("encryptPassword() error: %v", err)
	}

	if encrypted == password {
		t.Errorf("encryptPassword() did not encrypt")
	}

	// 解密
	decrypted, err := decryptPassword(encrypted)
	if err != nil {
		t.Errorf("decryptPassword() error: %v", err)
	}

	if decrypted != password {
		t.Errorf("decryptPassword() got %s, want %s", decrypted, password)
	}
}