package main

import (
	"fmt"
	"log"
	"smtp-lite/internal/config"
	"smtp-lite/internal/handler"
	"smtp-lite/internal/middleware"
	"smtp-lite/internal/model"
	"smtp-lite/internal/service"
	"smtp-lite/internal/version"
	"strings"
	"time"

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
	if err := db.AutoMigrate(
		&model.SmtpAccount{},
		&model.APIKey{},
		&model.SendLog{},
		&model.EmailTemplate{},
		&model.RecipientGroup{},
		&model.Recipient{},
		&model.Blacklist{},
		&model.Webhook{},
		&model.SendQueue{},
		&model.BatchSend{},
		&model.RateLimit{},
	); err != nil {
		log.Fatal("Failed to auto migrate:", err)
	}

	// 初始化服务
	authService := service.NewAuthService()
	smtpService := service.NewSmtpService(db)
	apiKeyService := service.NewAPIKeyService(db)
	sendService := service.NewSendService(db, smtpService)
	templateService := service.NewTemplateService(db)
	recipientService := service.NewRecipientService(db)
	blacklistService := service.NewBlacklistService(db)
	webhookService := service.NewWebhookService(db)
	rateLimitService := service.NewRateLimitService(db)
	trackService := service.NewTrackService(db)
	queueService := service.NewQueueService(db, sendService, smtpService, webhookService, rateLimitService, blacklistService)
	exportService := service.NewExportService(db)
	localeService := service.NewLocaleService()

	// 注入依赖
	sendService.SetWebhookService(webhookService)
	sendService.SetBlacklistService(blacklistService)
	sendService.SetRateLimitService(rateLimitService)
	sendService.SetTrackService(trackService)
	trackService.SetWebhookService(webhookService)

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(authService)
	smtpHandler := handler.NewSmtpHandler(smtpService)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService)
	sendHandler := handler.NewSendHandler(sendService, smtpService)
	templateHandler := handler.NewTemplateHandler(templateService)
	recipientHandler := handler.NewRecipientHandler(recipientService)
	blacklistHandler := handler.NewBlacklistHandler(blacklistService)
	webhookHandler := handler.NewWebhookHandler(webhookService)
	trackHandler := handler.NewTrackHandler(trackService)
	systemHandler := handler.NewSystemHandler()

	// 路由
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowOrigins := cfg.CORS.AllowOrigins
		allowed := false
		if len(allowOrigins) == 0 {
			// 未配置时允许同源请求（不设置 CORS 头）
			allowed = (origin == "")
		} else {
			for _, o := range allowOrigins {
				if o == "*" || strings.EqualFold(o, origin) {
					allowed = true
					break
				}
			}
		}
		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-API-Key")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 静态文件
	r.Static("/assets", "./frontend/dist/assets")

	// 追踪端点（公开，无需认证）
	r.GET("/track/open/:track_id.png", trackHandler.Open)
	r.GET("/track/click/:track_id", trackHandler.Click)

	// API 路由
	api := r.Group("/api/v1")
	{
		// 版本（公开）
		api.GET("/version", func(c *gin.Context) {
			c.JSON(200, gin.H{"version": version.Version})
		})

		// 语言（公开）
		api.GET("/locale", func(c *gin.Context) {
			locale := c.Query("locale")
			if locale == "" {
				locale = localeService.GetLocale()
			}
			c.JSON(200, gin.H{
				"locale":       locale,
				"translations": localeService.GetAllTranslations(locale),
			})
		})

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

			// 发送日志和统计
			protected.GET("/stats", sendHandler.Stats)
			protected.GET("/send/logs", sendHandler.Logs)

			// 邮件模板
			protected.GET("/templates", templateHandler.List)
			protected.GET("/templates/:id", templateHandler.Get)
			protected.POST("/templates", templateHandler.Create)
			protected.PUT("/templates/:id", templateHandler.Update)
			protected.DELETE("/templates/:id", templateHandler.Delete)
			protected.POST("/templates/:id/duplicate", templateHandler.Duplicate)

			// 收件人分组
			protected.GET("/recipient-groups", recipientHandler.GroupList)
			protected.POST("/recipient-groups", recipientHandler.GroupCreate)
			protected.PUT("/recipient-groups/:id", recipientHandler.GroupUpdate)

			// 收件人（通过 query 参数指定分组）
			protected.GET("/recipients", recipientHandler.RecipientListByGroup)
			protected.POST("/recipients", recipientHandler.RecipientCreate)
			protected.POST("/recipients/batch", recipientHandler.RecipientBatchImport)
			protected.DELETE("/recipients/:id", recipientHandler.RecipientDelete)

			// 分组详情（放在 recipients 后面避免冲突）
			protected.GET("/recipient-groups/:id", recipientHandler.GroupGet)
			protected.DELETE("/recipient-groups/:id", recipientHandler.GroupDelete)

			// 黑名单
			protected.GET("/blacklist", blacklistHandler.List)
			protected.POST("/blacklist", blacklistHandler.Add)
			protected.POST("/blacklist/batch", blacklistHandler.BatchAdd)
			protected.DELETE("/blacklist/:id", blacklistHandler.Remove)
			protected.GET("/blacklist/check", blacklistHandler.Check)

			// Webhook
			protected.GET("/webhooks", webhookHandler.List)
			protected.GET("/webhooks/:id", webhookHandler.Get)
			protected.POST("/webhooks", webhookHandler.Create)
			protected.PUT("/webhooks/:id", webhookHandler.Update)
			protected.DELETE("/webhooks/:id", webhookHandler.Delete)
			protected.POST("/webhooks/:id/toggle", webhookHandler.Toggle)
			protected.POST("/webhooks/:id/test", webhookHandler.Test)

			// 队列状态
			protected.GET("/queue/stats", func(c *gin.Context) {
				stats, _ := queueService.GetQueueStats()
				c.JSON(200, stats)
			})

			// 限流状态
			protected.GET("/rate-limit/status", func(c *gin.Context) {
				c.JSON(200, rateLimitService.GetLimitStatus())
			})

			// 导出
			protected.GET("/export/logs", func(c *gin.Context) {
				data, err := exportService.ExportSendLogsCSV(nil, nil)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.Header("Content-Disposition", "attachment; filename=send_logs.csv")
				c.Data(200, "text/csv", data)
			})

			protected.GET("/export/accounts", func(c *gin.Context) {
				data, err := exportService.ExportSmtpAccountsCSV()
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.Header("Content-Disposition", "attachment; filename=smtp_accounts.csv")
				c.Data(200, "text/csv", data)
			})

			protected.GET("/export/recipients", func(c *gin.Context) {
				data, err := exportService.ExportRecipientsCSV(nil)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.Header("Content-Disposition", "attachment; filename=recipients.csv")
				c.Data(200, "text/csv", data)
			})

			// 设置语言
			protected.POST("/settings/locale", func(c *gin.Context) {
				var req struct {
					Locale string `json:"locale" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					return
				}
				localeService.SetLocale(req.Locale)
				config.UpdateLocale(req.Locale)
				c.JSON(200, gin.H{"message": "Locale updated"})
			})

			// 系统
			protected.POST("/system/update-prepare", systemHandler.UpdatePrepare)
			protected.POST("/system/update", systemHandler.Update)
		}

		// 发送邮件（支持 Token 或 API Key 认证）
		api.POST("/send", middleware.TokenOrAPIKey(apiKeyService), sendHandler.Send)

		// 批量发送
		api.POST("/send/batch", middleware.TokenOrAPIKey(apiKeyService), func(c *gin.Context) {
			var req struct {
				Name     string   `json:"name" binding:"required"`
				Emails   []string `json:"emails" binding:"required"`
				Subject  string   `json:"subject" binding:"required"`
				Body     string   `json:"body" binding:"required"`
				IsHTML   bool     `json:"is_html"`
				FromName string   `json:"from_name"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			// 批量发送数量上限
			if len(req.Emails) > 1000 {
				c.JSON(400, gin.H{"error": "Batch size cannot exceed 1000"})
				return
			}

			sendReq := &service.SendRequest{
				Subject:  req.Subject,
				Body:     req.Body,
				IsHTML:   req.IsHTML,
				FromName: req.FromName,
			}

			batch, err := queueService.BatchSend(req.Name, req.Emails, sendReq)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(201, batch)
		})

		// 定时发送
		api.POST("/send/scheduled", middleware.TokenOrAPIKey(apiKeyService), func(c *gin.Context) {
			var req struct {
				To          string    `json:"to" binding:"required,email"`
				Subject     string    `json:"subject" binding:"required"`
				Body        string    `json:"body" binding:"required"`
				IsHTML      bool      `json:"is_html"`
				FromName    string    `json:"from_name"`
				ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			sendReq := &service.SendRequest{
				To:       req.To,
				Subject:  req.Subject,
				Body:     req.Body,
				IsHTML:   req.IsHTML,
				FromName: req.FromName,
			}

			task, err := queueService.AddToQueue(sendReq, &req.ScheduledAt, nil)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(201, task)
		})
	}

	// 启动队列处理
	queueService.Start()
	defer queueService.Stop()

	// SPA 路由支持
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("SMTP Lite %s starting on %s", version.Version, addr)
	log.Fatal(r.Run(addr))
}
