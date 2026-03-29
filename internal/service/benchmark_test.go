package service

import (
	"testing"
)

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
