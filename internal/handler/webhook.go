package handler

import (
	"encoding/json"
	"smtp-lite/internal/model"
	"smtp-lite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WebhookHandler struct {
	webhookSvc *service.WebhookService
}

func NewWebhookHandler(webhookSvc *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookSvc: webhookSvc}
}

func (h *WebhookHandler) List(c *gin.Context) {
	webhooks, err := h.webhookSvc.List()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, webhooks)
}

func (h *WebhookHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	webhook, err := h.webhookSvc.GetByID(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Webhook not found"})
		return
	}
	c.JSON(200, webhook)
}

type CreateWebhookRequest struct {
	Name   string   `json:"name" binding:"required"`
	URL    string   `json:"url" binding:"required"`
	Secret string   `json:"secret"`
	Events []string `json:"events"` // ["send_success", "send_failed", "opened", "clicked", "*"]
}

func (h *WebhookHandler) Create(c *gin.Context) {
	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// SSRF 防护：验证 webhook URL
	if err := service.ValidateWebhookURL(req.URL); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	eventsJSON, _ := json.Marshal(req.Events)

	webhook := &model.Webhook{
		Name:   req.Name,
		URL:    req.URL,
		Secret: req.Secret,
		Events: string(eventsJSON),
	}

	if err := h.webhookSvc.Create(webhook); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, webhook)
}

type UpdateWebhookRequest struct {
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Secret string   `json:"secret"`
	Events []string `json:"events"`
}

func (h *WebhookHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.URL != "" {
		// SSRF 防护：验证 webhook URL
		if err := service.ValidateWebhookURL(req.URL); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		updates["url"] = req.URL
	}
	if req.Secret != "" {
		updates["secret"] = req.Secret
	}
	if len(req.Events) > 0 {
		eventsJSON, _ := json.Marshal(req.Events)
		updates["events"] = string(eventsJSON)
	}

	if err := h.webhookSvc.Update(id, updates); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Updated"})
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.webhookSvc.Delete(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Deleted"})
}

func (h *WebhookHandler) Toggle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.webhookSvc.Toggle(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Toggled"})
}

func (h *WebhookHandler) Test(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	webhook, err := h.webhookSvc.GetByID(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Webhook not found"})
		return
	}

	// 发送测试事件
	h.webhookSvc.Trigger("test", map[string]interface{}{
		"message": "This is a test webhook",
	})

	c.JSON(200, gin.H{
		"message": "Test webhook sent",
		"url":     webhook.URL,
	})
}
