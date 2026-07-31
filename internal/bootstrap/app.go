package bootstrap

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/migrate"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/live"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/jwt"
	"gorm.io/gorm"
)

// App 包含应用运行时依赖
type App struct {
	Engine      *gin.Engine
	DB          *gorm.DB
	Redis       *redis.Client
	LiveService *live.Service
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
	// service
	services := InitServices(repos, db, rdb, cfg)
	// handler
	handlers := InitHandlers(services)
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
	}
}
