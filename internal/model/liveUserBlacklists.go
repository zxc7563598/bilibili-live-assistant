package model

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
)

type LiveUserBlacklist struct {
	ID              int64           `gorm:"primaryKey"`
	RoomID          int64           `gorm:"not null;index;comment:相关房间ID"`
	UID             int64           `gorm:"not null;index;comment:用户uid"`
	Uname           string          `gorm:"type:varchar(100);not null;comment:用户名称"`
	Msg             string          `gorm:"type:varchar(200);not null;comment:涉案弹幕"`
	RansomAmount    int64           `gorm:"not null;default:0;index;comment:赎回需要的金额"`
	MuteDuration    int64           `gorm:"not null;default:0;index;comment:禁言时长(分钟)"`
	MuteExpiresAt   int64           `gorm:"not null;comment:禁言自动解除时间"`
	UnmuteFailCount int64           `gorm:"not null;default:0;index;comment:解禁失败次数"`
	Status          enum.MuteStatus `gorm:"type:smallint;comment:解禁状态"`
	BaseModel
}

func (LiveUserBlacklist) TableName() string {
	return "live_user_blacklists"
}
