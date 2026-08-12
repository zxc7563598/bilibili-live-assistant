package bootstrap

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/admin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/admin_role"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_session"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_credit_log"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_sign_log"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/menu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/robot_config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/role"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/role_menu"
	"gorm.io/gorm"
)

type Repositories struct {
	Admin             admin.Repository
	Role              role.Repository
	Menu              menu.Repository
	AdminRole         admin_role.Repository
	RoleMenu          role_menu.Repository
	LiveDanmu         live_danmu.Repository
	LiveGift          live_gift.Repository
	LiveSession       live_session.Repository
	LiveUser          live_user.Repository
	RobotConfig       robot_config.Repository
	LiveUserCreditLog live_user_credit_log.Repository
	LiveUserSignLog   live_user_sign_log.Repository
}

func InitRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Admin:             admin.New(db),
		Role:              role.New(db),
		Menu:              menu.New(db),
		AdminRole:         admin_role.New(db),
		RoleMenu:          role_menu.New(db),
		LiveDanmu:         live_danmu.New(db),
		LiveGift:          live_gift.New(db),
		LiveSession:       live_session.New(db),
		LiveUser:          live_user.New(db),
		RobotConfig:       robot_config.New(db),
		LiveUserCreditLog: live_user_credit_log.New(db),
		LiveUserSignLog:   live_user_sign_log.New(db),
	}
}
