package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

type LiveDanmu struct {
	ID          int64         `gorm:"primaryKey"`
	RoomID      int64         `gorm:"not null;index;comment:相关房间ID"`
	UID         int64         `gorm:"not null;index;comment:用户uid"`
	Uname       string         `gorm:"type:varchar(100);not null;comment:用户名称"`
	Msg         string         `gorm:"type:varchar(200);not null;comment:发送弹幕"`
	LiveID      int64         `gorm:"not null;default:0;index;comment:相关直播ID"`
	BadgeUID    int64         `gorm:"comment:勋章主播UID"`
	BadgeUname  string         `gorm:"type:varchar(100);comment:勋章主播名"`
	BadgeRoomID int64         `gorm:"comment:勋章房间ID"`
	BadgeName   string         `gorm:"type:varchar(100);comment:勋章名称"`
	BadgeLevel  int64         `gorm:"comment:勋章等级"`
	BadgeType   enum.BadgeType `gorm:"type:smallint;comment:勋章类型"`
	SendAt      int64          `gorm:"not null;comment:弹幕发送时间"`
	BaseModel
}

func (LiveDanmu) TableName() string {
	return "live_danmus"
}
