package handler

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler { return &SystemHandler{} }

// Update 一键更新：优先调用 update.sh，否则内联执行
func (h *SystemHandler) Update(c *gin.Context) {
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
