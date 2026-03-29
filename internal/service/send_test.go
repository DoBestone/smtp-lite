package service

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"smtp-lite/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestSendService(t *testing.T) (*SendService, *SmtpService, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.SmtpAccount{}, &model.SendLog{}); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	smtpSvc := NewSmtpService(db)
	sendSvc := NewSendService(db, smtpSvc)
	return sendSvc, smtpSvc, db
}

func TestSendReturnsClearMessageWhenNoActiveAccount(t *testing.T) {
	sendSvc, _, _ := newTestSendService(t)

	resp, err := sendSvc.Send(&SendRequest{To: "user@example.com", Subject: "hello", Body: "world"})
	if err != nil {
		t.Fatalf("send returned error: %v", err)
	}
	if resp.Success {
		t.Fatalf("expected send to fail")
	}
	if resp.Message != "No active SMTP account configured" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestSendReturnsRateLimitMessageWhenAllAccountsExhausted(t *testing.T) {
	sendSvc, _, db := newTestSendService(t)
	account := &model.SmtpAccount{
		Email:             "sender@example.com",
		PasswordEncrypted: "invalid-base64",
		SmtpHost:          "smtp.example.com",
		SmtpPort:          587,
		DailyLimit:        10,
		DailyUsed:         10,
		LastResetDate:     time.Now(),
		Status:            "active",
	}
	if err := db.Create(account).Error; err != nil {
		t.Fatalf("create account: %v", err)
	}

	resp, err := sendSvc.Send(&SendRequest{To: "user@example.com", Subject: "hello", Body: "world"})
	if err != nil {
		t.Fatalf("send returned error: %v", err)
	}
	if resp.Success {
		t.Fatalf("expected send to fail")
	}
	if resp.Message != "No available SMTP account: all active accounts reached daily limits" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
	if len(resp.Details) == 0 || !strings.Contains(resp.Details[0], "rate limited accounts") {
		t.Fatalf("expected rate limit details, got %#v", resp.Details)
	}
}

func TestSendReturnsDetailedDecryptFailures(t *testing.T) {
	sendSvc, _, db := newTestSendService(t)
	accounts := []*model.SmtpAccount{
		{Email: "first@example.com", PasswordEncrypted: "bad-1", SmtpHost: "smtp1.example.com", SmtpPort: 587, Status: "active", LastResetDate: time.Now(), Priority: 10},
		{Email: "second@example.com", PasswordEncrypted: "bad-2", SmtpHost: "smtp2.example.com", SmtpPort: 587, Status: "active", LastResetDate: time.Now(), Priority: 5},
	}
	for _, account := range accounts {
		if err := db.Create(account).Error; err != nil {
			t.Fatalf("create account: %v", err)
		}
	}

	resp, err := sendSvc.Send(&SendRequest{To: "user@example.com", Subject: "hello", Body: "world"})
	if err != nil {
		t.Fatalf("send returned error: %v", err)
	}
	if resp.Success {
		t.Fatalf("expected send to fail")
	}
	if resp.Message != "All 2 SMTP account attempts failed" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
	if len(resp.Details) != 2 {
		t.Fatalf("expected 2 details, got %#v", resp.Details)
	}
	if !strings.Contains(resp.Details[0], "password decrypt failed") {
		t.Fatalf("expected decrypt detail, got %#v", resp.Details)
	}
}

func TestListActiveAccountsExcludingOrdersByPriorityAndUsage(t *testing.T) {
	_, smtpSvc, db := newTestSendService(t)
	accounts := []*model.SmtpAccount{
		{Email: "low@example.com", PasswordEncrypted: "x", SmtpHost: "smtp.example.com", SmtpPort: 587, Status: "active", LastResetDate: time.Now(), Priority: 1, DailyUsed: 1},
		{Email: "high-busy@example.com", PasswordEncrypted: "x", SmtpHost: "smtp.example.com", SmtpPort: 587, Status: "active", LastResetDate: time.Now(), Priority: 10, DailyUsed: 5},
		{Email: "high-idle@example.com", PasswordEncrypted: "x", SmtpHost: "smtp.example.com", SmtpPort: 587, Status: "active", LastResetDate: time.Now(), Priority: 10, DailyUsed: 1},
	}
	for _, account := range accounts {
		if err := db.Create(account).Error; err != nil {
			t.Fatalf("create account: %v", err)
		}
	}

	result, err := smtpSvc.ListActiveAccountsExcluding(nil)
	if err != nil {
		t.Fatalf("list active accounts: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 accounts, got %d", len(result))
	}
	if result[0].Email != "high-idle@example.com" || result[1].Email != "high-busy@example.com" || result[2].Email != "low@example.com" {
		t.Fatalf("unexpected order: %s, %s, %s", result[0].Email, result[1].Email, result[2].Email)
	}
}
