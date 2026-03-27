package handler

import (
	"net/http"
	"smtp-lite/internal/service"

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

// Click 追踪链接点击
func (h *TrackHandler) Click(c *gin.Context) {
	trackID := c.Param("track_id")
	url := c.Query("url")

	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL required"})
		return
	}

	// 记录点击事件
	h.trackSvc.RecordClick(trackID, url)

	// 重定向到目标 URL
	c.Redirect(http.StatusFound, url)
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