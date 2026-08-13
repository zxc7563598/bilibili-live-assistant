package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

type LiveInteractWord struct {
	ID         int64             `gorm:"primaryKey"`
	RoomID     int64             `gorm:"not null;index;comment:直播间ID"`
	UID        int64             `gorm:"not null;index;comment:用户UID"`
	Uname      string            `gorm:"type:varchar(100);not null;comment:用户名"`
	MsgType    enum.InteractType `gorm:"not null;index;comment:消息类型 1=进房欢迎 2=关注 3=分享"`
	Timestamp  int64             `gorm:"not null;index;comment:事件时间戳"`
	BadgeUID   int64             `gorm:"not null;default:0;comment:勋章主播UID"`
	BadgeName  string            `gorm:"type:varchar(100);comment:勋章名称"`
	BadgeLevel int64             `gorm:"not null;default:0;comment:勋章等级"`
	BadgeType  enum.BadgeType    `gorm:"type:smallint;comment:勋章类型"`
	BaseModel
}

func (LiveInteractWord) TableName() string {
	return "live_interact_words"
}
