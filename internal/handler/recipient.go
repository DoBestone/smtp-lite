package handler

import (
	"net/mail"
	"smtp-lite/internal/model"
	"smtp-lite/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RecipientHandler struct {
	recipientSvc *service.RecipientService
}

func NewRecipientHandler(recipientSvc *service.RecipientService) *RecipientHandler {
	return &RecipientHandler{recipientSvc: recipientSvc}
}

// GroupList 获取分组列表
func (h *RecipientHandler) GroupList(c *gin.Context) {
	groups, err := h.recipientSvc.GroupList()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, groups)
}

// GroupGet 获取分组详情
func (h *RecipientHandler) GroupGet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	group, err := h.recipientSvc.GroupGetByID(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Group not found"})
		return
	}
	c.JSON(200, group)
}

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *RecipientHandler) GroupCreate(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	group := &model.RecipientGroup{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.recipientSvc.GroupCreate(group); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, group)
}

func (h *RecipientHandler) GroupUpdate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
	}

	if err := h.recipientSvc.GroupUpdate(id, updates); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Updated"})
}

func (h *RecipientHandler) GroupDelete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.recipientSvc.GroupDelete(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Deleted"})
}

// RecipientListByGroup 通过 query 参数获取收件人列表
func (h *RecipientHandler) RecipientListByGroup(c *gin.Context) {
	groupIDStr := c.Query("group_id")
	if groupIDStr == "" {
		c.JSON(400, gin.H{"error": "group_id required"})
		return
	}

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid group ID"})
		return
	}

	recipients, err := h.recipientSvc.RecipientList(groupID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, recipients)
}

type CreateRecipientRequest struct {
	GroupID string `json:"group_id" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Name    string `json:"name"`
}

func (h *RecipientHandler) RecipientCreate(c *gin.Context) {
	var req CreateRecipientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	groupID, err := uuid.Parse(req.GroupID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid group ID"})
		return
	}

	recipient := &model.Recipient{
		GroupID: groupID,
		Email:   req.Email,
		Name:    req.Name,
	}

	if err := h.recipientSvc.RecipientCreate(recipient); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, recipient)
}

type BatchImportRequest struct {
	GroupID string `json:"group_id" binding:"required"`
	Emails  string `json:"emails" binding:"required"` // 逗号或换行分隔的邮箱
}

func (h *RecipientHandler) RecipientBatchImport(c *gin.Context) {
	var req BatchImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	groupID, err := uuid.Parse(req.GroupID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid group ID"})
		return
	}

	// 解析邮箱列表
	emails := strings.Split(req.Emails, "\n")
	if len(emails) == 1 {
		emails = strings.Split(req.Emails, ",")
	}

	// 清理邮箱
	cleanEmails := []string{}
	for _, email := range emails {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}
		// 使用 net/mail 严格校验邮箱格式
		if _, err := mail.ParseAddress(email); err == nil {
			cleanEmails = append(cleanEmails, email)
		}
	}

	successCount, blacklistedCount, err := h.recipientSvc.RecipientBatchCreate(groupID, cleanEmails)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":     "Import completed",
		"total":       len(cleanEmails),
		"success":     successCount,
		"blacklisted": blacklistedCount,
	})
}

func (h *RecipientHandler) RecipientDelete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.recipientSvc.RecipientDelete(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Deleted"})
}
