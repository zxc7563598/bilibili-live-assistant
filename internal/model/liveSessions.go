package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

type LiveSession struct {
	ID             int64            `gorm:"primaryKey"`
	RoomID         int64            `gorm:"not null;index;comment:相关房间ID"`
	UID            int64            `gorm:"not null;comment:主播UID"`
	LiveKey        string           `gorm:"type:varchar(100);comment:直播流标识 key"`
	LivePlatform   string           `gorm:"type:varchar(100);comment:开播平台"`
	StartAt        int64            `gorm:"not null;comment:开播时间"`
	StartSource    enum.StartSource `gorm:"not null;comment:开播来源"`
	DanmuCount     int64            `gorm:"not null;default:0;comment:弹幕总数"`
	GiftCount      int64            `gorm:"not null;default:0;comment:礼物总数(含大航海+醒目留言)"`
	GuardCount     int64            `gorm:"not null;default:0;comment:大航海开通数"`
	SuperChatCount int64            `gorm:"not null;default:0;comment:醒目留言数"`
	TotalRevenue   int64            `gorm:"not null;default:0;comment:总收益(分)"`
	EndAt          int64            `gorm:"comment:下播时间"`
	EndReason      enum.EndReason   `gorm:"comment:下播原因"`
	EndSource      enum.EndSource   `gorm:"comment:下播来源"`
	EndDetail      string           `gorm:"type:varchar(500);comment:下播详情（如切断原因信息）"`
	BaseModel
}

func (LiveSession) TableName() string {
	return "live_sessions"
}

// LiveSessionListPageQuery 直播场次分页查询入参，不对应数据库表
type LiveSessionListPageQuery struct {
	RoomID       *int64
	UID          *int64
	StartAtStart *int64
	StartAtEnd   *int64
	EndAtStart   *int64
	EndAtEnd     *int64
	Offset       int
	Limit        int
}

// LiveSessionUpdateStartForm 更新开播信息，不对应数据库表
type LiveSessionUpdateStartForm struct {
	StartAt      *int64
	LivePlatform *string
}

// LiveSessionUpdateEndForm 更新下播信息，不对应数据库表
type LiveSessionUpdateEndForm struct {
	EndAt     *int64
	EndReason *enum.EndReason
	EndSource *enum.EndSource
	EndDetail *string
}

// LiveSessionUpdateStatsForm 更新统计数据，不对应数据库表
type LiveSessionUpdateStatsForm struct {
	DanmuCount     *int64
	GiftCount      *int64
	GuardCount     *int64
	SuperChatCount *int64
	TotalRevenue   *int64
}
