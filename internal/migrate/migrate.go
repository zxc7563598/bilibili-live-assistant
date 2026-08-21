package migrate

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	// 兼容历史数据：为 live_user_sign_logs 回填 sign_date 字段，并对重复记录去重
	if err := backfillLiveUserSignLogSignDate(db); err != nil {
		return err
	}
	if err := dedupeLiveUserSignLogSignDate(db); err != nil {
		return err
	}
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
		&model.LiveUserSignLog{},
		&model.LiveUserBlacklist{},
		&model.LiveInteractWord{},
		&model.LiveUserCreditLog{},
	)
}
