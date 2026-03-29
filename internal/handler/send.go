package handler

import (
	"smtp-lite/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SendHandler struct {
	sendService *service.SendService
	smtpService *service.SmtpService
}

func NewSendHandler(sendService *service.SendService, smtpService *service.SmtpService) *SendHandler {
	return &SendHandler{sendService: sendService, smtpService: smtpService}
}

func (h *SendHandler) Send(c *gin.Context) {
	var req service.SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.sendService.Send(&req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if resp.Success {
		c.JSON(200, resp)
	} else {
		c.JSON(200, resp)
	}
}

func (h *SendHandler) Logs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	logs, total, err := h.sendService.Logs(page, pageSize)
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
}

func (h *SendHandler) Stats(c *gin.Context) {
	stats, err := h.sendService.Stats()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, stats)
}
