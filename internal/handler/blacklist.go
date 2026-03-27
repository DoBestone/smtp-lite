package handler

import (
	"smtp-lite/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BlacklistHandler struct {
	blacklistSvc *service.BlacklistService
}

func NewBlacklistHandler(blacklistSvc *service.BlacklistService) *BlacklistHandler {
	return &BlacklistHandler{blacklistSvc: blacklistSvc}
}

func (h *BlacklistHandler) List(c *gin.Context) {
	list, err := h.blacklistSvc.List()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, list)
}

type AddBlacklistRequest struct {
	Email  string `json:"email" binding:"required,email"`
	Reason string `json:"reason"`
}

func (h *BlacklistHandler) Add(c *gin.Context) {
	var req AddBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.blacklistSvc.Add(req.Email, req.Reason); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Added to blacklist"})
}

type BatchAddBlacklistRequest struct {
	Emails string `json:"emails" binding:"required"` // 逗号或换行分隔
	Reason string `json:"reason"`
}

func (h *BlacklistHandler) BatchAdd(c *gin.Context) {
	var req BatchAddBlacklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 解析邮箱
	emails := strings.Split(req.Emails, "\n")
	if len(emails) == 1 {
		emails = strings.Split(req.Emails, ",")
	}

	cleanEmails := []string{}
	for _, email := range emails {
		email = strings.TrimSpace(email)
		if email != "" && strings.Contains(email, "@") {
			cleanEmails = append(cleanEmails, email)
		}
	}

	count, err := h.blacklistSvc.BatchAdd(cleanEmails, req.Reason)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "Added to blacklist",
		"count":   count,
	})
}

func (h *BlacklistHandler) Remove(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.blacklistSvc.Remove(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Removed from blacklist"})
}

func (h *BlacklistHandler) Check(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(400, gin.H{"error": "Email required"})
		return
	}

	isBlacklisted := h.blacklistSvc.IsBlacklisted(email)
	c.JSON(200, gin.H{
		"email":      email,
		"blacklisted": isBlacklisted,
	})
}