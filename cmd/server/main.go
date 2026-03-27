package main

import (
	"fmt"
	"log"
	"smtp-lite/internal/config"
	"smtp-lite/internal/handler"
	"smtp-lite/internal/middleware"
	"smtp-lite/internal/model"
	"smtp-lite/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	db, err := gorm.Open(sqlite.Open("smtp-lite.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// 自动迁移
	db.AutoMigrate(&model.SmtpAccount{}, &model.APIKey{}, &model.SendLog{})

	// 初始化服务
	authService := service.NewAuthService()
	smtpService := service.NewSmtpService(db)
	apiKeyService := service.NewAPIKeyService(db)
	sendService := service.NewSendService(db, smtpService)

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(authService)
	smtpHandler := handler.NewSmtpHandler(smtpService)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService)
	sendHandler := handler.NewSendHandler(sendService, smtpService)

	// 路由
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-API-Key")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 静态文件
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/", "./frontend/dist/index.html")
	r.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	// API 路由
	api := r.Group("/api/v1")
	{
		// 认证
		api.POST("/auth/login", authHandler.Login)

		// 需要 Token 认证的路由
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// 修改密码
			protected.POST("/auth/change-password", authHandler.ChangePassword)

			// SMTP 账号
			protected.GET("/smtp-accounts", smtpHandler.List)
			protected.POST("/smtp-accounts", smtpHandler.Create)
			protected.PUT("/smtp-accounts/:id", smtpHandler.Update)
			protected.DELETE("/smtp-accounts/:id", smtpHandler.Delete)
			protected.POST("/smtp-accounts/:id/test", smtpHandler.Test)
			protected.POST("/smtp-accounts/:id/test-send", smtpHandler.TestSend)
			protected.POST("/smtp-accounts/:id/toggle", smtpHandler.Toggle)

			// API Key
			protected.GET("/api-keys", apiKeyHandler.List)
			protected.POST("/api-keys", apiKeyHandler.Create)
			protected.DELETE("/api-keys/:id", apiKeyHandler.Delete)
			protected.POST("/api-keys/:id/reset", apiKeyHandler.Reset)
			protected.GET("/stats", sendHandler.Stats)
			protected.GET("/send/logs", sendHandler.Logs)
		}

		// 发送邮件（支持 Token 或 API Key 认证）
		api.POST("/send", middleware.TokenOrAPIKey(apiKeyService), sendHandler.Send)
	}

	// SPA 路由支持
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("SMTP Lite starting on %s", addr)
	log.Fatal(r.Run(addr))
}
