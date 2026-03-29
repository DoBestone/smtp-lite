package handler

import (
	"smtp-lite/internal/model"
	"smtp-lite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SmtpHandler struct {
	smtpService *service.SmtpService
}

func NewSmtpHandler(smtpService *service.SmtpService) *SmtpHandler {
	return &SmtpHandler{smtpService: smtpService}
}

func (h *SmtpHandler) List(c *gin.Context) {
	accounts, err := h.smtpService.List()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to list accounts"})
		return
	}

	// 隐藏敏感信息
	for i := range accounts {
		accounts[i].PasswordEncrypted = ""
	}

	c.JSON(200, accounts)
}

type CreateSmtpRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	SmtpHost   string `json:"smtp_host" binding:"required"`
	SmtpPort   int    `json:"smtp_port"`
	DailyLimit int    `json:"daily_limit"`
}

func (h *SmtpHandler) Create(c *gin.Context) {
	var req CreateSmtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.SmtpPort == 0 {
		req.SmtpPort = 587
	}

	account := &model.SmtpAccount{
		Email:             req.Email,
		PasswordEncrypted: req.Password,
		SmtpHost:          req.SmtpHost,
		SmtpPort:          req.SmtpPort,
		DailyLimit:        req.DailyLimit,
	}

	if err := h.smtpService.Create(account); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create account"})
		return
	}

	account.PasswordEncrypted = ""
	c.JSON(201, account)
}

type UpdateSmtpRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	SmtpHost   string `json:"smtp_host"`
	SmtpPort   int    `json:"smtp_port"`
	DailyLimit int    `json:"daily_limit"`
	Status     string `json:"status"`
}

func (h *SmtpHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateSmtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	if req.SmtpHost != "" {
		updates["smtp_host"] = req.SmtpHost
	}
	if req.SmtpPort > 0 {
		updates["smtp_port"] = req.SmtpPort
	}
	if req.DailyLimit > 0 {
		updates["daily_limit"] = req.DailyLimit
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := h.smtpService.Update(id, updates); err != nil {
		c.JSON(500, gin.H{"error": "Failed to update account"})
		return
	}

	c.JSON(200, gin.H{"message": "Updated"})
}

func (h *SmtpHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.smtpService.Delete(id); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(200, gin.H{"message": "Deleted"})
}

func (h *SmtpHandler) Test(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	account, err := h.smtpService.GetByID(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Account not found"})
		return
	}

	password, err := h.smtpService.DecryptAccountPassword(account)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to decrypt password"})
		return
	}

	err = h.smtpService.TestConnection(account.SmtpHost, account.SmtpPort, account.Email, password)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	h.smtpService.ClearError(account.ID)

	c.JSON(200, gin.H{"success": true, "message": "Connection successful"})
}

func (h *SmtpHandler) Toggle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.smtpService.Toggle(id); err != nil {
		c.JSON(500, gin.H{"error": "Failed to toggle account"})
		return
	}

	c.JSON(200, gin.H{"message": "Toggled"})
}

type TestSendRequest struct {
	To string `json:"to" binding:"required,email"`
}

func (h *SmtpHandler) TestSend(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var req TestSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请提供有效的收件邮箱地址"})
		return
	}

	account, err := h.smtpService.GetByID(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "账号不存在"})
		return
	}

	if err := h.smtpService.TestSend(account, req.To); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	h.smtpService.ClearError(account.ID)

	c.JSON(200, gin.H{"success": true, "message": "测试邮件已发送，请检查收件箱"})
}
