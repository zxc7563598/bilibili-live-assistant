package bootstrap

import (
	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/address"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/admin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/altcha"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/appconfig"
	feedbackHdlr "github.com/zxc7563598/bilibili-live-assistant/internal/handler/feedback"
	liveHdlr "github.com/zxc7563598/bilibili-live-assistant/internal/handler/live"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/livedanmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/livegift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/livepk"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/menu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/order"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/product"
	robotconfigHdlr "github.com/zxc7563598/bilibili-live-assistant/internal/handler/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/role"
	uploadHdlr "github.com/zxc7563598/bilibili-live-assistant/internal/handler/upload"
)

type Handlers struct {
	Admin       *admin.Handler
	Menu        *menu.Handler
	Role        *role.Handler
	Altcha      *altcha.Handler
	Live        *liveHdlr.Handler
	RobotConfig *robotconfigHdlr.Handler
	LiveDanmu   *livedanmu.Handler
	LiveGift    *livegift.Handler
	LivePk      *livepk.Handler
	LiveUser    *liveuser.Handler
	AppConfig   *appconfig.Handler
	Product     *product.Handler
	Order       *order.Handler
	Address     *address.Handler
	Feedback    *feedbackHdlr.Handler
	Upload      *uploadHdlr.Handler
}

func InitHandlers(svc *Services, rdb *redis.Client) *Handlers {
	return &Handlers{
		Admin:       admin.New(svc.Admin, svc.Altcha),
		Menu:        menu.New(svc.Menu),
		Role:        role.New(svc.Role),
		Altcha:      altcha.New(svc.Altcha),
		Live:        liveHdlr.New(svc.Live, rdb),
		RobotConfig: robotconfigHdlr.New(svc.RobotConfig, svc.Live),
		LiveDanmu:   livedanmu.New(svc.LiveDanmu),
		LiveGift:    livegift.New(svc.LiveGift),
		LivePk:      livepk.New(svc.LivePk),
		LiveUser:    liveuser.New(svc.LiveUser, svc.RobotConfig),
		AppConfig:   appconfig.New(svc.AppConfig),
		Product:     product.New(svc.Product),
		Order:       order.New(svc.Order),
		Address:     address.New(svc.Address),
		Feedback:    feedbackHdlr.New(svc.Feedback),
		Upload:      uploadHdlr.New(svc.Upload),
	}
}
