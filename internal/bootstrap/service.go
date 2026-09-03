package bootstrap

import (
	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/admin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/altcha"
	appconfigsvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/live"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livedanmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livegift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/menu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/order"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/product"
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
	LiveDanmu   livedanmu.Service
	LiveGift    livegift.Service
	LiveUser    *liveuser.Service
	AppConfig   appconfigsvc.Service
	Product     product.Service
	Order       *order.Service
}

func InitServices(repo *Repositories, db *gorm.DB, rdb *redis.Client, cfg *config.Config, configCache *robotconfig.Cache, appConfigCache *appconfig.Cache) *Services {
	robotConfigSvc := robotconfigsvc.New(repo.RobotConfig, configCache, db)
	liveUserSvc := liveuser.New(db, rdb, appConfigCache, repo.LiveUser, repo.LiveUserCreditLog, repo.LiveDanmu, repo.LiveGift, repo.LiveSession)
	return &Services{
		Admin:       *admin.New(repo.Admin, repo.AdminRole, repo.Role, db, rdb),
		Role:        *role.New(repo.Role, repo.Admin, repo.RoleMenu, repo.AdminRole, repo.Menu, db, rdb),
		Menu:        *menu.New(repo.Menu),
		Altcha:      *altcha.New(cfg.Altcha.HMACKey),
		Live:        live.New(cfg.Live, robotConfigSvc, liveUserSvc, configCache, repo.LiveDanmu, repo.LiveGift, repo.LiveSession, repo.LiveUser, repo.LiveUserSignLog, repo.LiveUserBlacklist, repo.LiveInteractWord),
		RobotConfig: robotConfigSvc,
		LiveDanmu:   *livedanmu.New(repo.LiveDanmu),
		LiveGift:    *livegift.New(repo.LiveGift),
		LiveUser:    liveUserSvc,
		AppConfig:   *appconfigsvc.New(appConfigCache, repo.AppConfig),
		Product:     *product.New(db, repo.Product, repo.ProductSku, repo.ProductSkuStockLog, repo.ProductImage, repo.ProductSpec, repo.ProductSpecValue),
		Order:       order.New(db, repo.LiveUserOrder, repo.LiveUserOrderDraft, repo.Product, repo.ProductSku, repo.ProductSkuStockLog),
	}
}
