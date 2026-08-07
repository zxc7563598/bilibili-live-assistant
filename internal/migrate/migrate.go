package migrate

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Admin{},
		&model.Role{},
		&model.Menu{},
		&model.RoleMenu{},
		&model.AdminRole{},
		&model.LiveUser{},
		&model.LiveDanmu{},
		&model.LiveGift{},
		&model.LiveSession{},
		&model.RobotConfig{},
	)
}
