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
	os.Exit(m.Run())
}

func setupSendTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect database: %v", err)
	}
	db.AutoMigrate(&model.SmtpAccount{}, &model.SendLog{}, &model.Blacklist{})
	return db
}

func TestSendService_Stats(t *testing.T) {
	db := setupSendTestDB(t)
	smtpSvc := NewSmtpService(db)
	sendSvc := NewSendService(db, smtpSvc)

	stats, err := sendSvc.Stats()
	if err != nil {
		t.Errorf("Stats() error: %v", err)
	}

	// 空数据库应该返回零值
	if stats["total_sent"].(int64) != 0 {
		t.Errorf("Stats() total_sent = %v, want 0", stats["total_sent"])
	}
}

func TestSendService_Logs(t *testing.T) {
	db := setupSendTestDB(t)
	smtpSvc := NewSmtpService(db)
	sendSvc := NewSendService(db, smtpSvc)

	// 创建测试日志
	log1 := &model.SendLog{
		ToEmail: "user1@example.com",
		Subject: "Test 1",
		Status:  "success",
	}
	log2 := &model.SendLog{
		ToEmail: "user2@example.com",
		Subject: "Test 2",
		Status:  "failed",
	}
	db.Create(log1)
	db.Create(log2)

	// 测试查询
	logs, total, err := sendSvc.Logs(1, 10)
	if err != nil {
		t.Errorf("Logs() error: %v", err)
	}

	if total != 2 {
		t.Errorf("Logs() total = %d, want 2", total)
	}

	if len(logs) != 2 {
		t.Errorf("Logs() returned %d logs, want 2", len(logs))
	}
}

func TestBlacklistService_Check(t *testing.T) {
	db := setupSendTestDB(t)
	svc := NewBlacklistService(db)

	// 添加黑名单
	svc.Add("spam@example.com", "测试黑名单")

	tests := []struct {
		email     string
		blacklisted bool
	}{
		{"spam@example.com", true},
		{"normal@example.com", false},
		{"SPAM@example.com", false}, // 大小写敏感
	}

	for _, tt := range tests {
		got := svc.IsBlacklisted(tt.email)
		if got != tt.blacklisted {
			t.Errorf("IsBlacklisted(%s) = %v, want %v", tt.email, got, tt.blacklisted)
		}
	}
}

func TestBlacklistService_AddRemove(t *testing.T) {
	db := setupSendTestDB(t)
	svc := NewBlacklistService(db)

	// 添加
	err := svc.Add("test@example.com", "测试")
	if err != nil {
		t.Errorf("Add() error: %v", err)
	}

	// 验证存在
	if !svc.IsBlacklisted("test@example.com") {
		t.Errorf("Add() failed to add to blacklist")
	}

	// 获取列表
	list, _ := svc.List()
	if len(list) != 1 {
		t.Errorf("List() returned %d items, want 1", len(list))
	}

	// 移除
	err = svc.Remove(list[0].ID)
	if err != nil {
		t.Errorf("Remove() error: %v", err)
	}

	// 验证已移除
	if svc.IsBlacklisted("test@example.com") {
		t.Errorf("Remove() failed to remove from blacklist")
	}
}