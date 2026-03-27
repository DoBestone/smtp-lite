package middleware

import (
	"smtp-lite/internal/service"

	"github.com/gin-gonic/gin"
)

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

// APIKeyRequired API Key 认证中间件
func APIKeyRequired(apiKeyService *service.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(401, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		if !apiKeyService.Validate(apiKey) {
			c.JSON(401, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// 更新最后使用时间
		key := apiKeyService.FindByKey(apiKey)
		if key != nil {
			apiKeyService.UpdateLastUsed(key.ID)
		}

		c.Set("auth_type", "api_key")
		c.Next()
	}
}

// TokenOrAPIKey 支持 Token 或 API Key 认证
func TokenOrAPIKey(apiKeyService *service.APIKeyService) gin.HandlerFunc {
	authService := service.NewAuthService()

	return func(c *gin.Context) {
		// 先尝试 API Key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			if apiKeyService.Validate(apiKey) {
				key := apiKeyService.FindByKey(apiKey)
				if key != nil {
					apiKeyService.UpdateLastUsed(key.ID)
				}
				c.Set("auth_type", "api_key")
				c.Next()
				return
			}
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