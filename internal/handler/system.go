package handler

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
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

// Update 一键更新：需要确认令牌，优先调用 update.sh，否则内联执行
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

	c.JSON(200, gin.H{"message": "更新已启动，服务将在构建完成后自动重启"})

	go func() {
		time.Sleep(300 * time.Millisecond)

		wd, err := os.Getwd()
		if err != nil {
			log.Printf("[update] getwd: %v", err)
			return
		}

		// 优先使用 update.sh（更强大的重启检测）
		script := filepath.Join(wd, "update.sh")
		if _, err := os.Stat(script); err == nil {
			log.Println("[update] 使用 update.sh --force")
			cmd := exec.Command("bash", script, "--force")
			cmd.Dir = wd
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("[update] update.sh 执行失败: %v，回退到内联更新", err)
			} else {
				return // update.sh 自行处理重启
			}
		}

		// 回退：内联 git pull → go build → exec 重启
		log.Println("[update] 内联更新 step 1/3 - git pull")
		pull := exec.Command("git", "pull")
		pull.Dir, pull.Stdout, pull.Stderr = wd, os.Stdout, os.Stderr
		if err := pull.Run(); err != nil {
			log.Printf("[update] git pull failed: %v", err)
			return
		}

		binary, err := os.Executable()
		if err != nil {
			log.Printf("[update] get executable: %v", err)
			return
		}

		log.Println("[update] 内联更新 step 2/3 - go build")
		build := exec.Command("go", "build", "-o", binary, "./cmd/server/")
		build.Dir, build.Stdout, build.Stderr = wd, os.Stdout, os.Stderr
		if err := build.Run(); err != nil {
			log.Printf("[update] go build failed: %v", err)
			return
		}

		log.Println("[update] 内联更新 step 3/3 - exec 重启")
		time.Sleep(200 * time.Millisecond)
		if err := syscall.Exec(binary, os.Args, os.Environ()); err != nil {
			log.Printf("[update] exec failed: %v", err)
		}
	}()
}
