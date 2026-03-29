package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"smtp-lite/internal/version"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	mu           sync.Mutex
	confirmToken string
	tokenExpiry  time.Time
}

func NewSystemHandler() *SystemHandler { return &SystemHandler{} }

// UpdatePrepare 生成更新确认令牌（两步确认，防止误操作或 token 泄露导致 RCE）
func (h *SystemHandler) UpdatePrepare(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 生成随机确认令牌，60 秒有效
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		c.JSON(500, gin.H{"error": "无法生成确认令牌"})
		return
	}
	h.confirmToken = hex.EncodeToString(b)
	h.tokenExpiry = time.Now().Add(60 * time.Second)

	c.JSON(200, gin.H{
		"confirm_token": h.confirmToken,
		"message":       "请在 60 秒内使用此 confirm_token 确认更新",
	})
}

// UpdateCheck 查询 GitHub 最新版本，返回是否有可用更新及更新日志
func (h *SystemHandler) UpdateCheck(c *gin.Context) {
	current := version.Version

	type releaseResp struct {
		TagName     string `json:"tag_name"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		HTMLURL     string `json:"html_url"`
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/DoBestone/smtp-lite/releases/latest")
	if err != nil {
		c.JSON(200, gin.H{
			"current":    current,
			"latest":     "",
			"has_update": false,
			"error":      "unable to reach GitHub: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var rel releaseResp
	if err := json.Unmarshal(body, &rel); err != nil || rel.TagName == "" {
		c.JSON(200, gin.H{
			"current":    current,
			"latest":     "",
			"has_update": false,
			"error":      "invalid GitHub response",
		})
		return
	}

	c.JSON(200, gin.H{
		"current":      current,
		"latest":       rel.TagName,
		"has_update":   rel.TagName != current,
		"changelog":    rel.Body,
		"published_at": rel.PublishedAt,
		"release_url":  rel.HTMLURL,
	})
}

// Changelog 获取最近的版本更新日志列表
func (h *SystemHandler) Changelog(c *gin.Context) {
	type releaseItem struct {
		TagName     string `json:"tag_name"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		HTMLURL     string `json:"html_url"`
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/DoBestone/smtp-lite/releases?per_page=10")
	if err != nil {
		c.JSON(500, gin.H{"error": "无法获取更新日志: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var releases []releaseItem
	if err := json.Unmarshal(body, &releases); err != nil {
		c.JSON(500, gin.H{"error": "解析更新日志失败"})
		return
	}

	c.JSON(200, gin.H{"releases": releases})
}

// Update 一键更新：需要确认令牌，调用 update.sh --force 下载预编译二进制
func (h *SystemHandler) Update(c *gin.Context) {
	var req struct {
		ConfirmToken string `json:"confirm_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "需要提供 confirm_token 确认更新"})
		return
	}

	h.mu.Lock()
	validToken := h.confirmToken != "" &&
		req.ConfirmToken == h.confirmToken &&
		time.Now().Before(h.tokenExpiry)
	// 一次性令牌，使用后立即清除
	h.confirmToken = ""
	h.tokenExpiry = time.Time{}
	h.mu.Unlock()

	if !validToken {
		c.JSON(403, gin.H{"error": "确认令牌无效或已过期，请重新获取"})
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		c.JSON(500, gin.H{"error": "无法获取工作目录: " + err.Error()})
		return
	}

	script := filepath.Join(wd, "update.sh")
	if _, err := os.Stat(script); err != nil {
		c.JSON(500, gin.H{"error": "update.sh 不存在，请在服务器上下载最新的 update.sh"})
		return
	}

	c.JSON(200, gin.H{"message": "更新已启动，服务将在下载完成后自动重启"})

	go func() {
		time.Sleep(300 * time.Millisecond)
		log.Println("[update] 调用 update.sh --force")
		cmd := exec.Command("bash", script, "--force")
		cmd.Dir = wd
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("[update] update.sh 执行失败: %v", err)
		}
	}()
}
