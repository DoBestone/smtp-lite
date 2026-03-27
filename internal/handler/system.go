package handler

import (
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler { return &SystemHandler{} }

// Update 一键更新：git pull → go build → exec 重启
func (h *SystemHandler) Update(c *gin.Context) {
	c.JSON(200, gin.H{"message": "更新已启动，服务将在构建完成后自动重启"})

	go func() {
		time.Sleep(300 * time.Millisecond) // 等待响应发出

		wd, err := os.Getwd()
		if err != nil {
			log.Printf("[update] getwd: %v", err)
			return
		}

		log.Println("[update] step 1/3 - git pull")
		pull := exec.Command("git", "pull")
		pull.Dir = wd
		pull.Stdout = os.Stdout
		pull.Stderr = os.Stderr
		if err := pull.Run(); err != nil {
			log.Printf("[update] git pull failed: %v", err)
			return
		}

		binary, err := os.Executable()
		if err != nil {
			log.Printf("[update] get executable: %v", err)
			return
		}

		log.Println("[update] step 2/3 - go build")
		build := exec.Command("go", "build", "-o", binary, "./cmd/server/")
		build.Dir = wd
		build.Stdout = os.Stdout
		build.Stderr = os.Stderr
		if err := build.Run(); err != nil {
			log.Printf("[update] go build failed: %v", err)
			return
		}

		log.Println("[update] step 3/3 - restarting")
		time.Sleep(200 * time.Millisecond)
		if err := syscall.Exec(binary, os.Args, os.Environ()); err != nil {
			log.Printf("[update] exec restart failed: %v", err)
		}
	}()
}
