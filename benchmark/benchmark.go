package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// 性能测试工具
// go run benchmark.go

var (
	baseURL    = "http://localhost:8090/api/v1"
	token      string
	totalReq   int64
	successReq int64
	failedReq  int64
	totalTime  int64
)

func main() {
	fmt.Println("=== SMTP Lite 性能测试 ===")
	fmt.Println()

	// 1. 登录获取 token
	fmt.Print("1. 登录测试... ")
	token = login()
	if token == "" {
		fmt.Println("❌ 登录失败")
		os.Exit(1)
	}
	fmt.Println("✅")

	// 2. 并发请求测试
	concurrency := []int{10, 50, 100, 200}
	requests := 1000

	for _, c := range concurrency {
		fmt.Printf("\n2. 并发测试 (并发=%d, 请求=%d)\n", c, requests)
		runConcurrentTest(c, requests)
	}

	// 3. 持续压力测试
	fmt.Println("\n3. 持续压力测试 (30秒)")
	runSustainedTest(30 * time.Second)

	// 4. 报告
	fmt.Println("\n=== 测试报告 ===")
	printReport()
}

func login() string {
	resp, err := http.PostForm(baseURL+"/auth/login", url.Values{
		"username": {"admin"},
		"password": {"admin123"},
	})
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if token, ok := result["token"].(string); ok {
		return token
	}
	return ""
}

func runConcurrentTest(concurrency, requests int) {
	atomic.StoreInt64(&totalReq, 0)
	atomic.StoreInt64(&successReq, 0)
	atomic.StoreInt64(&failedReq, 0)
	atomic.StoreInt64(&totalTime, 0)

	var wg sync.WaitGroup
	reqPerWorker := requests / concurrency

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < reqPerWorker; j++ {
				makeRequest()
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	total := atomic.LoadInt64(&totalReq)
	success := atomic.LoadInt64(&successReq)
	failed := atomic.LoadInt64(&failedReq)
	avgTime := atomic.LoadInt64(&totalTime) / total

	fmt.Printf("   总请求: %d\n", total)
	fmt.Printf("   成功: %d (%.1f%%)\n", success, float64(success)/float64(total)*100)
	fmt.Printf("   失败: %d (%.1f%%)\n", failed, float64(failed)/float64(total)*100)
	fmt.Printf("   QPS: %.1f\n", float64(total)/elapsed.Seconds())
	fmt.Printf("   平均响应: %dms\n", avgTime/1000000)
	fmt.Printf("   总耗时: %v\n", elapsed)
}

func runSustainedTest(duration time.Duration) {
	atomic.StoreInt64(&totalReq, 0)
	atomic.StoreInt64(&successReq, 0)
	atomic.StoreInt64(&failedReq, 0)
	atomic.StoreInt64(&totalTime, 0)

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// 启动 50 个并发 worker
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					makeRequest()
				}
			}
		}()
	}

	time.Sleep(duration)
	close(stop)
	wg.Wait()

	total := atomic.LoadInt64(&totalReq)
	success := atomic.LoadInt64(&successReq)
	failed := atomic.LoadInt64(&failedReq)
	avgTime := atomic.LoadInt64(&totalTime) / total

	fmt.Printf("   总请求: %d\n", total)
	fmt.Printf("   成功: %d (%.1f%%)\n", success, float64(success)/float64(total)*100)
	fmt.Printf("   失败: %d (%.1f%%)\n", failed, float64(failed)/float64(total)*100)
	fmt.Printf("   QPS: %.1f\n", float64(total)/duration.Seconds())
	fmt.Printf("   平均响应: %dms\n", avgTime/1000000)
}

func makeRequest() {
	atomic.AddInt64(&totalReq, 1)

	start := time.Now()

	req, _ := http.NewRequest("GET", baseURL+"/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	elapsed := time.Since(start)
	atomic.AddInt64(&totalTime, int64(elapsed))

	if err != nil || resp.StatusCode >= 400 {
		atomic.AddInt64(&failedReq, 1)
	} else {
		atomic.AddInt64(&successReq, 1)
	}

	if resp != nil {
		resp.Body.Close()
	}
}

func printReport() {
	fmt.Println("\n测试完成！")
	fmt.Println("\n建议指标：")
	fmt.Println("  - QPS > 500: 优秀")
	fmt.Println("  - QPS 200-500: 良好")
	fmt.Println("  - QPS < 200: 需优化")
	fmt.Println()
	fmt.Println("  - 成功率 > 99%: 优秀")
	fmt.Println("  - 成功率 95-99%: 良好")
	fmt.Println("  - 成功率 < 95%: 需排查")
	fmt.Println()
	fmt.Println("  - 平均响应 < 100ms: 优秀")
	fmt.Println("  - 平均响应 100-300ms: 良好")
	fmt.Println("  - 平均响应 > 300ms: 需优化")
}
