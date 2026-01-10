// @title mdblog API
// @version 1.0
// @description 轻量级 Markdown 博客系统 API
// @host localhost:8080
// @BasePath /

package main

import (
	"context"
	"flag"
	"log"
	"mdblog/internal/pkg"
	"mdblog/internal/router"
	"mdblog/internal/theme"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 命令行参数
	buildMode := flag.Bool("build", false, "生成静态站点")
	outputDir := flag.String("output", "public", "静态站点输出目录")
	flag.Parse()

	// 1. Initialize Config
	pkg.InitConfig()

	// 2. Set Gin Mode (release for production)
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 3. Initialize Markdown Processor
	pkg.InitMarkdown()

	// 4. Initialize Store (Load posts into memory & Index for search)
	pkg.InitStore()

	// 静态生成模式
	if *buildMode {
		generator := pkg.NewStaticGenerator(*outputDir)
		if err := generator.Generate(); err != nil {
			log.Fatalf("生成静态站点失败: %v", err)
		}
		return
	}

	// 5. Initialize Comments
	pkg.InitComments()

	// 6. Initialize Stats
	pkg.InitStats()

	// 5. Load Theme Templates
	theme.InitPongo2()         // 前台 Pongo2
	theme.LoadAdminTemplates() // 后台 原生模板

	// 6. Setup Router
	r := router.SetupRouter()

	// 7. Create HTTP Server
	srv := &http.Server{
		Addr:    ":" + pkg.AppConfig.Port,
		Handler: r,
	}

	// 8. Start Server in goroutine
	go func() {
		log.Printf("Server starting on port %s...", pkg.AppConfig.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 9. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}
