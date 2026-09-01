package bootstrap

import (
	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/admin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/altcha"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/appconfig"
	liveHdlr "github.com/zxc7563598/bilibili-live-assistant/internal/handler/live"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/livedanmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/livegift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/menu"
	robotconfigHdlr "github.com/zxc7563598/bilibili-live-assistant/internal/handler/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler/role"
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
	LiveUser    *liveuser.Handler
	AppConfig   *appconfig.Handler
}

func InitHandlers(svc *Services, rdb *redis.Client) *Handlers {
	return &Handlers{
		Admin:       admin.New(&svc.Admin, &svc.Altcha),
		Menu:        menu.New(&svc.Menu),
		Role:        role.New(&svc.Role),
		Altcha:      altcha.New(&svc.Altcha),
		Live:        liveHdlr.New(svc.Live, rdb),
		RobotConfig: robotconfigHdlr.New(svc.RobotConfig, svc.Live),
		LiveDanmu:   livedanmu.New(&svc.LiveDanmu),
		LiveGift:    livegift.New(&svc.LiveGift),
		LiveUser:    liveuser.New(svc.LiveUser),
		AppConfig:   appconfig.New(&svc.AppConfig, rdb),
	}
}
