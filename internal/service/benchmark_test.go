package service

import (
	"os"
	"testing"

	"smtp-lite/internal/config"
)

func TestMain(m *testing.M) {
	config.Load()
	config.Get().Encryption.Key = "smtp-lite-test-encryption-32-byte!"
	os.Exit(m.Run())
}

func BenchmarkAuthService_Login(b *testing.B) {
	svc := NewAuthService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Login("admin", "admin123")
	}
}

func BenchmarkAuthService_ValidateToken(b *testing.B) {
	svc := NewAuthService()
	token, _ := svc.Login("admin", "admin123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.ValidateToken(token)
	}
}

func BenchmarkEncryptPassword(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encryptPassword("testpassword123")
	}
}

func BenchmarkDecryptPassword(b *testing.B) {
	encrypted, _ := encryptPassword("testpassword123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decryptPassword(encrypted)
	}
}