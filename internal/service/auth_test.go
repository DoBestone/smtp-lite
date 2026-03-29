package service

import (
	"testing"
)

func TestAuthService_Login(t *testing.T) {

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{"正确密码", "admin", "admin123", false},
		{"错误密码", "admin", "wrongpassword", true},
		{"空密码", "admin", "", true},
		{"错误用户名", "wronguser", "admin123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAuthService()
			token, err := svc.Login(tt.username, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Login() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Login() unexpected error: %v", err)
				}
				if token == "" {
					t.Errorf("Login() returned empty token")
				}
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	svc := NewAuthService()

	// 先登录获取有效 token
	validToken, _ := svc.Login("admin", "admin123")

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{"有效Token", validToken, false},
		{"空Token", "", true},
		{"无效Token", "invalid.token.here", true},
		{"过期Token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNjAwMDAwMDAwfQ.invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ValidateToken(tt.token)

			if tt.wantErr && err == nil {
				t.Errorf("ValidateToken() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateToken() unexpected error: %v", err)
			}
		})
	}
}
