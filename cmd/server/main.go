package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"smtp-lite/internal/config"
	"smtp-lite/internal/handler"
	"smtp-lite/internal/middleware"
	"smtp-lite/internal/model"
	"smtp-lite/internal/service"
	"smtp-lite/internal/version"
	"smtp-lite/web"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// --version flag
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(version.Version)
		return
	}

	// 加载配置
	cfg := config.Load()

	// 设置 Gin 模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	var db *gorm.DB
	var err error
	switch cfg.Database.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect MySQL database:", err)
		}
		log.Println("[database] Using MySQL:", cfg.Database.Host, cfg.Database.Name)
	default:
		dbPath := cfg.Database.SQLite
		if dbPath == "" {
			dbPath = "smtp-lite.db"
		}
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			log.Fatal("Failed to connect SQLite database:", err)
		}
		log.Println("[database] Using SQLite:", dbPath)
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
		&model.AuditLog{},
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
	auditService := service.NewAuditService(db)

	// 注入依赖
	sendService.SetWebhookService(webhookService)
	sendService.SetBlacklistService(blacklistService)
	sendService.SetRateLimitService(rateLimitService)
	sendService.SetTrackService(trackService)
	trackService.SetWebhookService(webhookService)

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(authService, auditService)
	smtpHandler := handler.NewSmtpHandler(smtpService, auditService)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyService, auditService)
	sendHandler := handler.NewSendHandler(sendService, smtpService)
	templateHandler := handler.NewTemplateHandler(templateService)
	recipientHandler := handler.NewRecipientHandler(recipientService)
	blacklistHandler := handler.NewBlacklistHandler(blacklistService)
	webhookHandler := handler.NewWebhookHandler(webhookService)
	trackHandler := handler.NewTrackHandler(trackService)
	systemHandler := handler.NewSystemHandler(auditService)

	// 路由
	r := gin.Default()

	// 请求体大小限制：10MB
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)
		c.Next()
	})

	// Release 模式下隐藏内部错误细节
	if cfg.Server.Mode == "release" {
		r.Use(func(c *gin.Context) {
			c.Next()
			if c.Writer.Status() >= 500 {
				c.JSON(c.Writer.Status(), gin.H{"error": "internal server error"})
			}
		})
	}

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

	// 内嵌前端：将 web/dist 挂载为前端资源
	frontendDist, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		log.Fatal("failed to access embedded frontend:", err)
	}
	httpFS := http.FS(frontendDist)
	fileServer := http.FileServer(httpFS)

	// 静态资源（/assets/, /favicon.*）
	r.GET("/assets/*filepath", gin.WrapH(http.StripPrefix("/", fileServer)))
	r.GET("/favicon.ico", gin.WrapH(http.StripPrefix("/", fileServer)))
	r.GET("/favicon.svg", gin.WrapH(http.StripPrefix("/", fileServer)))

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
			// 修改账号密码
			protected.POST("/auth/change-password", authHandler.ChangePassword)
			// 修改用户名
			protected.POST("/auth/change-username", authHandler.ChangeUsername)

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
				validLocales := map[string]bool{"zh-CN": true, "en-US": true}
				if !validLocales[req.Locale] {
					c.JSON(400, gin.H{"error": "unsupported locale, valid: zh-CN, en-US"})
					return
				}
				localeService.SetLocale(req.Locale)
				config.UpdateLocale(req.Locale)
				c.JSON(200, gin.H{"message": "Locale updated"})
			})

			// 审计日志
			protected.GET("/audit-logs", func(c *gin.Context) {
				page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
				pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
				if page < 1 {
					page = 1
				}
				if pageSize < 1 || pageSize > 100 {
					pageSize = 50
				}
				logs, total, err := auditService.List(page, pageSize)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, gin.H{
					"logs":       logs,
					"total":      total,
					"page":       page,
					"page_size":  pageSize,
					"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
				})
			})

			// 系统
			protected.GET("/system/update-check", systemHandler.UpdateCheck)
			protected.GET("/system/changelog", systemHandler.Changelog)
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

	// SPA 路由支持 - 从内嵌 FS 读取 index.html
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/track/") {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		data, err := web.Dist.ReadFile("dist/index.html")
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})

	// 启动服务器（优雅关闭）
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 定期清理过期限流记录
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rateLimitService.CleanupExpired()
			case <-quit:
				return
			}
		}
	}()

	go func() {
		log.Printf("SMTP Lite %s starting on %s", version.Version, addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// 等待退出信号
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	queueService.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}
