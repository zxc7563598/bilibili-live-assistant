package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/docs"
	"github.com/zxc7563598/bilibili-live-assistant/internal/bootstrap"
	"github.com/zxc7563598/bilibili-live-assistant/internal/config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/version"
)

// @title BiliLive Assistant API
// @version 1.0
// @description BiliLive Assistant 系统接口文档
// @contact.name 何俊杰
// @contact.email junjie.he.925@gmail.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @BasePath /
func main() {
	port := flag.Int("port", 25443, "服务端口")
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()
	// 未显式设置 GIN_MODE 时默认使用 release 模式（开发用 make dev-go 传 GIN_MODE=debug）
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	// 首次运行且配置文件缺失时，自动生成默认配置
	created, err := config.EnsureConfigFile(*configPath)
	if err != nil {
		log.Fatalf("初始化配置文件失败: %v", err)
	}
	if created {
		log.Printf("未找到配置文件，已自动生成默认配置: %s", *configPath)
	}
	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}
	// 初始化应用
	app := bootstrap.NewApp(cfg)
	// 确保资源在服务退出时关闭
	defer func() {
		if app.Redis != nil {
			if err := app.Redis.Close(); err != nil {
				log.Printf("关闭 Redis 连接失败: %v", err)
			}
			log.Println("Redis 连接已关闭")
		}
		if db, err := app.DB.DB(); err == nil {
			if err := db.Close(); err != nil {
				log.Printf("关闭数据库连接失败: %v", err)
			}
			log.Println("数据库连接已关闭")
		}
	}()
	// 使用配置中的端口（命令行参数优先）
	serverPort := cfg.Server.Port
	if *port != 9000 {
		serverPort = *port
	}
	// 创建 HTTP Server
	addr := fmt.Sprintf(":%d", serverPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      app.Engine,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}
	// 注册 Swagger
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", serverPort)
	// 启动服务
	go func() {
		log.Printf("服务在 %s 启动 (版本: %s, 提交: %s)\n", addr, version.Version, version.Commit)
		log.Printf("打开浏览器，前往：http://127.0.0.1%s/admin 访问后台\n", addr)
		log.Printf("默认账号：admin\n")
		log.Printf("默认密码：123456\n")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("监听错误: %v", err)
		}
	}()
	// 监听退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(
		quit,
		syscall.SIGINT,  // ctrl+c
		syscall.SIGTERM, // docker stop
	)
	<-quit
	log.Println("关闭服务...")
	// 先停止 Live Listener
	app.LiveService.Shutdown()
	log.Println("Live 服务已关闭")
	// 停止定时任务调度器
	app.Scheduler.Stop()
	log.Println("定时任务调度器已停止")
	// 再关闭 HTTP Server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务被迫关闭: %v", err)
	}
	log.Println("服务已退出")
}
