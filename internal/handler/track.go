package handler

import (
	"net/http"
	"net/url"
	"smtp-lite/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TrackHandler struct {
	trackSvc *service.TrackService
}

func NewTrackHandler(trackSvc *service.TrackService) *TrackHandler {
	return &TrackHandler{trackSvc: trackSvc}
}

// Open 追踪邮件打开（像素请求）
func (h *TrackHandler) Open(c *gin.Context) {
	trackID := c.Param("track_id")

	// 记录打开事件
	h.trackSvc.RecordOpen(trackID)

	// 返回 1x1 透明 GIF
	pixel := h.trackSvc.GetTrackPixel()
	c.Data(http.StatusOK, "image/gif", pixel)
}

// validateRedirectURL 验证重定向 URL 安全性，防止开放重定向
func validateRedirectURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	// 只允许 http/https 协议
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}
	// 不允许空 host
	if parsed.Host == "" {
		return false
	}
	// 不允许包含用户信息（防止 http://evil@legitimate.com 类攻击）
	if parsed.User != nil {
		return false
	}
	return true
}

// Click 追踪链接点击
func (h *TrackHandler) Click(c *gin.Context) {
	trackID := c.Param("track_id")
	redirectURL := c.Query("url")

	if redirectURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL required"})
		return
	}

	// 验证重定向 URL 安全性
	if !validateRedirectURL(redirectURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid redirect URL"})
		return
	}

	// 记录点击事件
	h.trackSvc.RecordClick(trackID, redirectURL)

	// 重定向到目标 URL
	c.Redirect(http.StatusFound, redirectURL)
}

// Stats 获取追踪统计
func (h *TrackHandler) Stats(c *gin.Context) {
	trackID := c.Param("track_id")

	// 从 logID 查找
	logID, err := uuid.Parse(trackID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	stats, err := h.trackSvc.GetTrackStats(logID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Track ID not found"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
