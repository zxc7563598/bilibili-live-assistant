package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

type LiveGift struct {
	ID                int64          `gorm:"primaryKey"`
	RoomID            int64          `gorm:"not null;index;comment:相关房间ID"`
	UID               int64          `gorm:"not null;index;comment:用户uid"`
	Uname             string         `gorm:"type:varchar(100);not null;comment:用户名称"`
	GiftType          enum.GiftType  `gorm:"not null;comment:礼物类型"`
	GiftID            int64          `gorm:"not null;comment:礼物ID"`
	GiftName          string         `gorm:"type:varchar(100);not null;comment:礼物名称"`
	Price             int64          `gorm:"not null;default:0;comment:礼物价格(分)"`
	Num               int64          `gorm:"not null;default:1;comment:礼物数量"`
	Message           string         `gorm:"type:varchar(200);comment:醒目留言文本"`
	AnchorID          int64          `gorm:"not null;index;comment:主播UID"`
	BadgeUID          int64          `gorm:"comment:勋章主播UID"`
	BadgeName         string         `gorm:"type:varchar(100);comment:勋章名称"`
	BadgeLevel        int64          `gorm:"comment:勋章等级"`
	BadgeType         enum.BadgeType `gorm:"type:smallint;comment:勋章类型"`
	LiveID            int64          `gorm:"not null;default:0;index;comment:关联直播场次ID"`
	SendAt            int64          `gorm:"not null;comment:发送时间（秒级时间戳）"`
	Original          enum.YesNo     `gorm:"type:smallint;comment:是否是原始商品"`
	OriginalGiftID    int64          `gorm:"comment:原始礼物ID"`
	OriginalGiftName  string         `gorm:"type:varchar(100);comment:原始礼物名称"`
	OriginalGiftPrice int64          `gorm:"default:0;comment:原始礼物价格(分)"`
	BaseModel
}

func (LiveGift) TableName() string {
	return "live_gifts"
}

// LiveGiftListPageQuery 礼物分页查询入参，不对应数据库表
type LiveGiftListPageQuery struct {
	RoomID      *int64
	UID         *int64
	Uname       *string
	GiftName    *string
	GiftType    *int
	Original    *int
	SendAtStart *int64
	SendAtEnd   *int64
	Offset      int
	Limit       int
}
