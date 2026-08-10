package bootstrap

import (
	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/admin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/altcha"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/live"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/menu"
	robotconfigsvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/role"
	"gorm.io/gorm"
)

type Services struct {
	Admin       admin.Service
	Role        role.Service
	Menu        menu.Service
	Altcha      altcha.Service
	Live        *live.Service
	RobotConfig *robotconfigsvc.Service
}

func InitServices(repo *Repositories, db *gorm.DB, rdb *redis.Client, cfg *config.Config, configCache *robotconfig.Cache) *Services {
	robotConfigSvc := robotconfigsvc.New(repo.RobotConfig, configCache, db)
	return &Services{
		Admin:       *admin.New(repo.Admin, repo.AdminRole, repo.Role, db, rdb),
		Role:        *role.New(repo.Role, repo.Admin, repo.RoleMenu, repo.AdminRole, repo.Menu, db, rdb),
		Menu:        *menu.New(repo.Menu),
		Altcha:      *altcha.New(cfg.Altcha.HMACKey),
		Live:        live.New(cfg.Live, robotConfigSvc, repo.LiveDanmu, repo.LiveGift, repo.LiveSession, repo.LiveUser),
		RobotConfig: robotConfigSvc,
	}
}
