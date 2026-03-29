package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"smtp-lite/internal/version"
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

// Update 一键更新：confirm_token 可选（兼容旧版前端），调用 update.sh 或 Go 自更新
func (h *SystemHandler) Update(c *gin.Context) {
	var req struct {
		ConfirmToken string `json:"confirm_token"` // 可选：旧版客户端不发送令牌
	}
	// 允许空 body 或无 confirm_token（忽略绑定错误）
	_ = c.ShouldBindJSON(&req)

	// 若提供了 confirm_token，则验证；否则直接放行（兼容旧版客户端）
	if req.ConfirmToken != "" {
		h.mu.Lock()
		validToken := h.confirmToken != "" &&
			req.ConfirmToken == h.confirmToken &&
			time.Now().Before(h.tokenExpiry)
		h.confirmToken = ""
		h.tokenExpiry = time.Time{}
		h.mu.Unlock()

		if !validToken {
			c.JSON(403, gin.H{"error": "确认令牌无效或已过期，请重新获取"})
			return
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		c.JSON(500, gin.H{"error": "无法获取工作目录: " + err.Error()})
		return
	}

	script := filepath.Join(wd, "update.sh")
	_, scriptErr := os.Stat(script)

	c.JSON(200, gin.H{"message": "更新已启动，服务将在下载完成后自动重启"})

	go func() {
		time.Sleep(300 * time.Millisecond)
		if scriptErr == nil {
			// update.sh 存在：使用脚本更新
			log.Println("[update] 调用 update.sh --force")
			cmd := exec.Command("bash", script, "--force")
			cmd.Dir = wd
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("[update] update.sh 执行失败: %v", err)
			}
		} else {
			// update.sh 不存在（旧版本部署）：Go 原生自更新
			log.Println("[update] update.sh 不存在，使用 Go 自更新下载预编译二进制")
			if err := goSelfUpdate(""); err != nil {
				log.Printf("[update] Go 自更新失败: %v", err)
			}
		}
	}()
}

// goSelfUpdate 从 GitHub Releases 下载当前平台的预编译二进制并替换自身。
// latestTag 为空时自动查询最新版本。
func goSelfUpdate(latestTag string) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	suffix := fmt.Sprintf("%s-%s", goos, goarch)
	if goarch == "arm" {
		// 项目只发布 ARMv7
		suffix = fmt.Sprintf("%s-armv7", goos)
	}
	assetName := "smtp-lite-" + suffix

	var downloadURL string

	if latestTag != "" {
		downloadURL = fmt.Sprintf(
			"https://github.com/DoBestone/smtp-lite/releases/download/%s/%s",
			latestTag, assetName,
		)
	} else {
		type asset struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}
		type release struct {
			TagName string  `json:"tag_name"`
			Assets  []asset `json:"assets"`
		}
		apiClient := &http.Client{Timeout: 15 * time.Second}
		resp, err := apiClient.Get("https://api.github.com/repos/DoBestone/smtp-lite/releases/latest")
		if err != nil {
			return fmt.Errorf("获取最新版本失败: %w", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var rel release
		if err := json.Unmarshal(body, &rel); err != nil {
			return fmt.Errorf("解析版本信息失败: %w", err)
		}
		for _, a := range rel.Assets {
			if a.Name == assetName {
				downloadURL = a.BrowserDownloadURL
				break
			}
		}
		if downloadURL == "" {
			return fmt.Errorf("未找到平台 %s 的预编译二进制，请手动更新", assetName)
		}
		log.Printf("[update] 最新版本 %s，下载 %s", rel.TagName, assetName)
	}

	// 获取当前可执行文件路径
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法获取当前执行文件路径: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("无法解析执行文件路径: %w", err)
	}

	log.Printf("[update] 下载新二进制到 %s", exe)

	// 下载新二进制
	dlClient := &http.Client{Timeout: 5 * time.Minute}
	dlResp, err := dlClient.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer dlResp.Body.Close()

	tmpFile := exe + ".new"
	f, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("无法创建临时文件: %w", err)
	}
	if _, err := io.Copy(f, dlResp.Body); err != nil {
		f.Close()
		os.Remove(tmpFile)
		return fmt.Errorf("写入新版本失败: %w", err)
	}
	f.Close()

	// 原子替换二进制
	oldFile := exe + ".old"
	if err := os.Rename(exe, oldFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("备份当前二进制失败: %w", err)
	}
	if err := os.Rename(tmpFile, exe); err != nil {
		_ = os.Rename(oldFile, exe) // 回滚
		os.Remove(tmpFile)
		return fmt.Errorf("替换二进制文件失败: %w", err)
	}
	os.Remove(oldFile)

	// 通过 syscall.Exec 重启（替换当前进程）
	log.Println("[update] 二进制替换完成，正在重启服务...")
	return syscall.Exec(exe, os.Args, os.Environ())
}
