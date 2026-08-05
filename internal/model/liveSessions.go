package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

type LiveSession struct {
	ID             uint64           `gorm:"primaryKey"`
	RoomID         uint64           `gorm:"not null;index;comment:相关房间ID"`
	UID            uint64           `gorm:"not null;comment:主播UID"`
	LiveKey        string           `gorm:"type:varchar(100);comment:直播流标识 key"`
	LivePlatform   string           `gorm:"type:varchar(100);comment:开播平台"`
	StartAt        int64            `gorm:"not null;comment:开播时间"`
	StartSource    enum.StartSource `gorm:"not null;comment:开播来源（1=事件 2=轮询 3=人工）"`
	DanmuCount     uint64           `gorm:"not null;default:0;comment:弹幕总数"`
	GiftCount      uint64           `gorm:"not null;default:0;comment:礼物总数(含大航海+醒目留言)"`
	GuardCount     uint64           `gorm:"not null;default:0;comment:大航海开通数"`
	SuperChatCount uint64           `gorm:"not null;default:0;comment:醒目留言数"`
	TotalRevenue   uint64           `gorm:"not null;default:0;comment:总收益(分)"`
	EndAt          int64            `gorm:"comment:下播时间"`
	EndReason      enum.EndReason   `gorm:"comment:下播原因"`
	EndSource      enum.EndSource   `gorm:"comment:下播来源"`
	EndDetail      string           `gorm:"type:varchar(500);comment:下播详情（如切断原因信息）"`
	BaseModel
}

func (LiveSession) TableName() string {
	return "live_sessions"
}
