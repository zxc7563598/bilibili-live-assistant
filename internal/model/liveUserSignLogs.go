package model

import (
	"errors"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"gorm.io/gorm"
)

type LiveUserSignLog struct {
	ID           int64           `gorm:"primaryKey"`
	UID          int64           `gorm:"not null;uniqueIndex:uk_uid_sign_date;comment:用户UID"`
	SignDate     string          `gorm:"type:varchar(10);not null;uniqueIndex:uk_uid_sign_date;comment:签到日期(YYYY-MM-DD)"`
	Uname        string          `gorm:"type:varchar(100);comment:用户名"`
	Msg          string          `gorm:"type:varchar(255);comment:签到弹幕内容"`
	BadgeUID     int64           `gorm:"not null;default:0;comment:勋章主播UID"`
	BadgeUname   string          `gorm:"type:varchar(100);comment:勋章主播名"`
	BadgeRoomID  int64           `gorm:"not null;default:0;comment:勋章房间ID"`
	BadgeName    string          `gorm:"type:varchar(100);comment:勋章名称"`
	BadgeLevel   int64           `gorm:"not null;default:0;comment:勋章等级"`
	BadgeType    enum.BadgeType  `gorm:"not null;default:0;comment:勋章类型"`
	RewardType   enum.CreditType `gorm:"not null;default:0;comment:奖励类型"`
	RewardAmount int64           `gorm:"not null;default:0;comment:奖励数量"`
	BaseModel
}

func (LiveUserSignLog) TableName() string {
	return "live_user_sign_logs"
}

// BeforeUpdate 禁止修改签到日志
func (l *LiveUserSignLog) BeforeUpdate(tx *gorm.DB) error {
	return errors.New("live_user_sign_logs 不允许修改")
}

// BeforeDelete 禁止删除签到日志
func (l *LiveUserSignLog) BeforeDelete(tx *gorm.DB) error {
	return errors.New("live_user_sign_logs 不允许删除")
}
