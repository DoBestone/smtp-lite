package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"smtp-lite/internal/config"
	"smtp-lite/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	config.Load()
	config.Get().Auth.Username = "admin"
	hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash test password: %v", err)
	}
	config.Get().Auth.Password = string(hashed)
	config.Get().JWT.Secret = "test-secret-key-for-testing-purpose-32b!"
	os.Exit(m.Run())
}

func TestAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 获取有效 token
	authSvc := service.NewAuthService()
	validToken, _ := authSvc.Login("admin", "admin123")

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{"无Token", "", http.StatusUnauthorized},
		{"有效Token", validToken, http.StatusOK},
		{"无效Token", "invalid.token.here", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(AuthRequired())
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("AuthRequired() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
