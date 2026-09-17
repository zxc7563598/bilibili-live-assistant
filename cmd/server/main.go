package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/docs"
	"github.com/zxc7563598/bilibili-live-assistant/internal/bootstrap"
	"github.com/zxc7563598/bilibili-live-assistant/internal/config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/migrate"
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
	seedProducts := flag.Bool("seed-products", false, "填充商城商品测试数据后退出，不启动服务（仅供测试环境）")
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
	// 手动填充测试数据：仅初始化数据库 + 迁移 + 写入测试商品，然后退出，不启动服务
	if *seedProducts {
		db, err := config.InitDB(cfg)
		if err != nil {
			log.Fatalf("无法初始化数据库: %v", err)
		}
		if err := migrate.Run(db); err != nil {
			log.Fatalf("数据库迁移失败: %v", err)
		}
		if err := migrate.SeedProducts(db); err != nil {
			log.Fatalf("测试数据填充失败: %v", err)
		}
		log.Println("测试数据填充完成，服务未启动")
		return
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
		remoteVersion, needUpdate, err := version.CheckUpdate()
		log.Printf("服务在 %s 启动 (版本: %s, 提交: %s)", addr, version.Version, version.Commit)
		log.Printf("打开浏览器，前往：http://127.0.0.1%s/admin 访问后台", addr)
		log.Printf("默认账号：admin")
		log.Printf("默认密码：123456")
		log.Printf("关闭该窗口后软件会自行退出，下次启动重新打开软件即可")
		log.Println(strings.Repeat("-", 50))
		if err != nil {
			log.Printf("[警告] 检查更新失败: %s (请检查网络连接)", err)
		} else {
			if needUpdate {
				log.Printf("[提示] 发现新版本！当前版本: %s，最新版本: %s，建议前往下载更新", version.Version, remoteVersion)
			} else {
				log.Printf("[提示] 当前已是最新版本 (%s)", version.Version)
			}
		}
		log.Println(strings.Repeat("-", 50))
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
