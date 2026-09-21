package bootstrap

import (
	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/address"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/admin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/altcha"
	appconfigsvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
	feedbacksvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/feedback"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/live"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livedanmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livegift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livepk"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/menu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/order"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/product"
	robotconfigsvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/role"
	uploadsvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/upload"
	"gorm.io/gorm"
)

// Services 各 Service 一律用指针：
// 一是各包的 New 本来就返回 *Service，二是 live.Service 内含 sync.Mutex 与 channel，
// 用值类型会被复制（go vet 的 copylocks 会报），且 RobotConfig / LiveUser 这两个
// 被 live.New 持有的共享实例会分裂成两份。
type Services struct {
	Address     *address.Service
	Admin       *admin.Service
	Role        *role.Service
	Menu        *menu.Service
	Altcha      *altcha.Service
	Live        *live.Service
	RobotConfig *robotconfigsvc.Service
	LiveDanmu   *livedanmu.Service
	LiveGift    *livegift.Service
	LivePk      *livepk.Service
	LiveUser    *liveuser.Service
	AppConfig   *appconfigsvc.Service
	Product     *product.Service
	Order       *order.Service
	Feedback    *feedbacksvc.Service
	Upload      *uploadsvc.Service
	Export      *export.Service
}

func InitServices(repo *Repositories, db *gorm.DB, rdb *redis.Client, cfg *config.Config, configCache *robotconfig.Cache, appConfigCache *appconfig.Cache) *Services {
	robotConfigSvc := robotconfigsvc.New(repo.RobotConfig, configCache, db)
	liveUserSvc := liveuser.New(db, rdb, appConfigCache, repo.LiveUser, repo.LiveUserCreditLog, repo.LiveDanmu, repo.LiveGift, repo.LiveSession)
	services := &Services{
		Address:     address.New(db, repo.LiveUserAddress),
		Admin:       admin.New(repo.Admin, repo.AdminRole, repo.Role, db, rdb),
		Role:        role.New(repo.Role, repo.Admin, repo.RoleMenu, repo.AdminRole, repo.Menu, db, rdb),
		Menu:        menu.New(repo.Menu),
		Altcha:      altcha.New(cfg.Altcha.HMACKey),
		Live:        live.New(cfg.Live, robotConfigSvc, liveUserSvc, configCache, repo.LiveDanmu, repo.LiveGift, repo.LiveSession, repo.LiveUser, repo.LiveUserSignLog, repo.LiveUserBlacklist, repo.LiveInteractWord, repo.LivePkLog),
		RobotConfig: robotConfigSvc,
		LiveDanmu:   livedanmu.New(repo.LiveDanmu),
		LiveGift:    livegift.New(repo.LiveGift),
		LivePk:      livepk.New(repo.LivePkLog),
		LiveUser:    liveUserSvc,
		AppConfig:   appconfigsvc.New(appConfigCache, repo.AppConfig),
		Product:     product.New(db, repo.Product, repo.ProductSku, repo.ProductImage, repo.ProductSpec, repo.ProductSpecValue),
		Order:       order.New(db, repo.LiveUserOrder, repo.LiveUserOrderDraft, repo.LiveUserAddress, repo.Product, repo.ProductSku, repo.ProductSkuStockLog, liveUserSvc),
		Feedback:    feedbacksvc.New(repo.Feedback),
		Upload:      uploadsvc.New(appConfigCache),
	}
	// 导出服务需要消费各业务模块的导出数据源，因此在服务组装完之后再装配
	services.Export = initExportService(cfg, services)
	return services
}
