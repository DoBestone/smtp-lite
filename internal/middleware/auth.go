package middleware

import (
	"smtp-lite/internal/service"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// apiKeyAttempt 记录 API Key 验证失败（暴力破解防护）
type apiKeyAttempt struct {
	count    int
	lockedAt time.Time
}

var (
	apiKeyAttemptsMu sync.Mutex
	apiKeyAttempts   = make(map[string]*apiKeyAttempt)
)

const (
	maxAPIKeyAttempts  = 10               // 最大连续失败次数
	apiKeyLockDuration = 15 * time.Minute // 锁定时长
)

func checkAPIKeyLock(ip string) bool {
	apiKeyAttemptsMu.Lock()
	defer apiKeyAttemptsMu.Unlock()

	attempt, exists := apiKeyAttempts[ip]
	if !exists {
		return false
	}
	if attempt.count >= maxAPIKeyAttempts {
		if time.Since(attempt.lockedAt) < apiKeyLockDuration {
			return true
		}
		delete(apiKeyAttempts, ip)
	}
	return false
}

func recordAPIKeyFailure(ip string) {
	apiKeyAttemptsMu.Lock()
	defer apiKeyAttemptsMu.Unlock()

	attempt, exists := apiKeyAttempts[ip]
	if !exists {
		attempt = &apiKeyAttempt{}
		apiKeyAttempts[ip] = attempt
	}
	attempt.count++
	if attempt.count >= maxAPIKeyAttempts {
		attempt.lockedAt = time.Now()
	}
}

func clearAPIKeyAttempts(ip string) {
	apiKeyAttemptsMu.Lock()
	defer apiKeyAttemptsMu.Unlock()
	delete(apiKeyAttempts, ip)
}

// AuthRequired JWT Token 认证中间件
func AuthRequired() gin.HandlerFunc {
	authService := service.NewAuthService()

	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Bearer token
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Set("auth_type", "token")
		c.Next()
	}
}

// APIKeyRequired API Key 认证中间件（含暴力破解防护）
func APIKeyRequired(apiKeyService *service.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(401, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		ip := c.ClientIP()
		if checkAPIKeyLock(ip) {
			c.JSON(429, gin.H{"error": "Too many failed attempts, please try later"})
			c.Abort()
			return
		}

		key := apiKeyService.FindByKey(apiKey)
		if key == nil {
			recordAPIKeyFailure(ip)
			c.JSON(401, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		clearAPIKeyAttempts(ip)
		apiKeyService.UpdateLastUsed(key.ID)

		c.Set("auth_type", "api_key")
		c.Next()
	}
}

// TokenOrAPIKey 支持 Token 或 API Key 认证（含暴力破解防护）
func TokenOrAPIKey(apiKeyService *service.APIKeyService) gin.HandlerFunc {
	authService := service.NewAuthService()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		// 先尝试 API Key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			if checkAPIKeyLock(ip) {
				c.JSON(429, gin.H{"error": "Too many failed attempts, please try later"})
				c.Abort()
				return
			}

			key := apiKeyService.FindByKey(apiKey)
			if key != nil {
				clearAPIKeyAttempts(ip)
				apiKeyService.UpdateLastUsed(key.ID)
				c.Set("auth_type", "api_key")
				c.Next()
				return
			}
			recordAPIKeyFailure(ip)
		}

		// 再尝试 Token
		token := c.GetHeader("Authorization")
		if token != "" {
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			claims, err := authService.ValidateToken(token)
			if err == nil {
				c.Set("username", claims.Username)
				c.Set("auth_type", "token")
				c.Next()
				return
			}
		}

		c.JSON(401, gin.H{"error": "Authentication required (Token or API Key)"})
		c.Abort()
	}
}