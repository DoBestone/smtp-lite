package handler

import (
	"fmt"
	"smtp-lite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type APIKeyHandler struct {
	apiKeyService *service.APIKeyService
	auditService  *service.AuditService
}

func NewAPIKeyHandler(apiKeyService *service.APIKeyService, auditService *service.AuditService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService, auditService: auditService}
}

func (h *APIKeyHandler) List(c *gin.Context) {
	keys, err := h.apiKeyService.List()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, keys)
}

type CreateAPIKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	key, fullKey, err := h.apiKeyService.Create(req.Name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	if h.auditService != nil {
		h.auditService.Log("apikey_create", fmt.Sprint(username), c.ClientIP(), "创建 API Key: "+req.Name)
	}

	c.JSON(201, gin.H{
		"id":         key.ID,
		"name":       key.Name,
		"key":        fullKey,
		"key_prefix": key.KeyPrefix,
		"warning":    "Save this key! It won't be shown again.",
		"created_at": key.CreatedAt,
	})
}

func (h *APIKeyHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.apiKeyService.Delete(uid); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	if h.auditService != nil {
		h.auditService.Log("apikey_delete", fmt.Sprint(username), c.ClientIP(), "删除 API Key: "+uid.String())
	}

	c.JSON(200, gin.H{"message": "Deleted"})
}

func (h *APIKeyHandler) Reset(c *gin.Context) {
	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	key, fullKey, err := h.apiKeyService.Reset(uid)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	if h.auditService != nil {
		h.auditService.Log("apikey_reset", fmt.Sprint(username), c.ClientIP(), "重置 API Key: "+uid.String())
	}

	c.JSON(200, gin.H{
		"id":         key.ID,
		"name":       key.Name,
		"key":        fullKey,
		"key_prefix": key.KeyPrefix,
		"warning":    "Save this key! It won't be shown again.",
	})
}
