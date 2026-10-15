package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// LiveGuardRecord 大航海开通台账
//
// 一行 = 一次开通 = 它覆盖的名单月份区间 [CoverFrom, CoverTo)，只追加不修改。
// live_users 的三个到期列只反映当前状态、续费即覆盖，留不下历史，
// 所以「某月谁持有某档位」只能靠这张表。
//
// CoverFrom / CoverTo 都是月首零点，CoverFrom 含、CoverTo 不含，区间长度 = Num 个月。
// 名单按月份求交（查询区间同样是月首到次月首），所以这两个值必须是月首，
// 且同一 (UID, GuardLevel) 下任意一个月最多只有一行命中——插入历史后要整体重排。
type LiveGuardRecord struct {
	ID         int64          `gorm:"primaryKey"`
	RoomID     int64          `gorm:"not null;index;comment:相关房间ID"`
	UID        int64          `gorm:"not null;index;comment:用户uid"`
	Uname      string         `gorm:"type:varchar(100);not null;comment:开通时的用户名称"`
	GuardLevel enum.BadgeType `gorm:"type:smallint;not null;index;comment:大航海等级, 3-舰长, 2-提督, 1-总督"`
	GiftID     int64          `gorm:"not null;default:0;comment:礼物ID"`
	Price      int64          `gorm:"not null;default:0;comment:开通价格(分)"`
	Num        int64          `gorm:"not null;default:1;comment:开通数量(份, 一份覆盖一个月)"`
	OpenAt     int64          `gorm:"not null;index;comment:开通时间(秒级时间戳)"`
	CoverFrom  int64          `gorm:"not null;index;comment:名单月份区间起点, 月首, 含(秒级时间戳)"`
	CoverTo    int64          `gorm:"not null;index;comment:名单月份区间终点, 月首, 不含(秒级时间戳)"`
	BaseModel
}

func (LiveGuardRecord) TableName() string {
	return "live_guard_records"
}

// LiveGuardRecordListPageQuery 大航海名单分页查询入参，不对应数据库表
//
// MonthFrom / MonthTo 是查询月份的 [月首, 次月首) 秒级区间：
// 判断覆盖窗口是否落在当月用它，统计当月开通次数也用它。
type LiveGuardRecordListPageQuery struct {
	GuardLevel *int
	MonthFrom  int64
	MonthTo    int64
	UID        *int64
	Uname      *string
	Offset     int
	Limit      int
	SortField  *string
	SortOrder  *string
}

// LiveGuardRecordOpenStatsQuery 当月开通统计查询入参，不对应数据库表
type LiveGuardRecordOpenStatsQuery struct {
	GuardLevel *int
	OpenFrom   int64
	OpenTo     int64
	UIDs       []int64
}
