package bootstrap

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/migrate"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/live"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/cron"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/jwt"
	"gorm.io/gorm"
)

// App 包含应用运行时依赖
type App struct {
	Engine      *gin.Engine
	DB          *gorm.DB
	Redis       *redis.Client
	LiveService *live.Service
	ConfigCache *robotconfig.Cache
	Scheduler   *cron.Scheduler
}

func NewApp(cfg *config.Config) *App {
	// 初始化日志
	logger.InitAll()
	// 初始化redis
	rdb, err := config.InitRedis(cfg)
	if err != nil {
		log.Fatalf("Redis配置存在但连接失败: %v", err)
	}
	if rdb != nil {
		log.Println("Redis已启用")
	}
	// 初始化数据库
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("无法初始化数据库: %v", err)
	}
	// 自动建表
	if err := migrate.Run(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	// 初始化填充数据
	if err := migrate.Seed(db); err != nil {
		log.Fatalf("数据填充失败: %v", err)
	}
	// 初始化jwt
	jwt.Init(cfg.JWT)
	// 处理依赖注入
	// repository
	repos := InitRepositories(db)
	// 初始化机器人配置缓存
	configCache := robotconfig.New(repos.RobotConfig)
	if err := configCache.Init(context.Background()); err != nil {
		log.Fatalf("机器人配置加载失败: %v", err)
	}
	// service
	services := InitServices(repos, db, rdb, cfg, configCache)
	// 定时任务调度器（项目启动后常驻，退出时在 main.go 中统一停止）
	scheduler := cron.New(cron.Job{
		Name:     "live-unmute",
		Interval: time.Minute,
		Run:      services.Live.UnmuteDueUsers,
	})
	scheduler.Start()
	// handler
	handlers := InitHandlers(services, rdb)
	// 如果配置了自动监听（is_listening=1）且 room_id>0，启动监听
	if isListening, ok := configCache.Get("room", "is_listening"); ok && isListening == "1" {
		if code, err := services.Live.StartListener(context.Background()); err != nil {
			log.Printf("[App] 自动启动监听失败 (code=%d): %v", code, err)
		} else if code == 0 {
			log.Println("[App] 已根据配置自动启动直播间监听")
		}
	}
	// i18n
	if err := i18n.InitLocales(); err != nil {
		log.Fatalf("无法初始化 i18n: %v", err)
	}
	// 注册路由
	r := gin.New()
	r = RouteRegister(r, rdb, handlers, cfg.CORS)
	return &App{
		Engine:      r,
		DB:          db,
		Redis:       rdb,
		LiveService: services.Live,
		ConfigCache: configCache,
		Scheduler:   scheduler,
	}
}
