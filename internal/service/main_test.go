package service

import (
	"log"
	"os"
	"testing"

	"smtp-lite/internal/config"

	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	config.Load()
	config.Get().Auth.Username = "admin"
	// 测试密码需要 bcrypt hash，与生产行为一致
	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("TestMain: bcrypt hash failed: %v", err)
	}
	config.Get().Auth.Password = string(hashed)
	config.Get().JWT.Secret = "test-secret-key-for-testing-purpose-32b!"
	config.Get().Encryption.Key = "smtp-lite-test-encryption-32-byte!"
	os.Exit(m.Run())
}
