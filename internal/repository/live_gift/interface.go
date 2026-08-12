package live_gift

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

// TimeRange 时间范围
type TimeRange struct {
	Start int64 // 起始时间戳（含）
	End   int64 // 结束时间戳（含）
}

// BlindBoxProfit 盲盒盈利统计
type BlindBoxProfit struct {
	Daily   int64 // 本日盈利
	Weekly  int64 // 本周盈利
	Monthly int64 // 本月盈利
	Total   int64 // 总计盈利
}

type Repository interface {
	base.Repository[model.LiveGift]
	// DistinctRoomIDs 获取全表中所有不重复的 RoomID
	DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error)
	// ListPage 分页查询礼物，Uname/GiftName 模糊匹配，SendAt 范围查询，按 SendAt 倒序
	ListPage(ctx context.Context, tx *gorm.DB, query model.LiveGiftListPageQuery) ([]model.LiveGift, int64, error)
	// ListByUID 根据 UID 查询礼物，按 SendAt 倒序，limit 控制最大条数
	ListByUID(ctx context.Context, tx *gorm.DB, uid int64, limit int) ([]model.LiveGift, error)
	// ListByLiveID 根据 LiveID 查询礼物，按 SendAt 倒序，limit 控制最大条数
	ListByLiveID(ctx context.Context, tx *gorm.DB, liveID int64, limit int) ([]model.LiveGift, error)
	// UpdateLiveIDByRoomIDAndTimeRange 将指定房间在时间范围内的礼物关联到直播记录
	UpdateLiveIDByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID, liveID int64) error
	// CountAndRevenueByRoomIDAndTimeRange 统计指定房间在时间范围内的礼物数量与收益（price * num）
	CountAndRevenueByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (count int64, revenue int64, err error)
	// CountGuardByRoomIDAndTimeRange 统计指定房间在时间范围内的大航海数量
	CountGuardByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error)
	// CountSuperChatByRoomIDAndTimeRange 统计指定房间在时间范围内的醒目留言数量
	CountSuperChatByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error)
	// SumBlindBoxProfit 统计盲盒盈利
	SumBlindBoxProfit(ctx context.Context, tx *gorm.DB, uid, roomID int64, day, week, month TimeRange) (*BlindBoxProfit, error)
}

// DistinctRoomIDs 获取全表中所有不重复的 RoomID
func (r *gormRepo) DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error) {
	db := r.getDB(ctx, tx)
	var roomIDs []int64
	if err := db.Model(&model.LiveGift{}).Distinct("room_id").Pluck("room_id", &roomIDs).Error; err != nil {
		return nil, err
	}
	return roomIDs, nil
}

// ListPage 分页查询礼物
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LiveGiftListPageQuery) ([]model.LiveGift, int64, error) {
	var list []model.LiveGift
	var total int64
	db := r.getDB(ctx, tx)
	db = db.Model(&model.LiveGift{})
	if v := query.RoomID; v != nil {
		db = db.Where("room_id = ?", *v)
	}
	if v := query.UID; v != nil {
		db = db.Where("uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("uname LIKE ?", "%"+escapeLike(*v)+"%")
	}
	if v := query.GiftName; v != nil && *v != "" {
		db = db.Where("gift_name LIKE ?", "%"+escapeLike(*v)+"%")
	}
	if v := query.GiftType; v != nil {
		db = db.Where("gift_type = ?", *v)
	}
	if v := query.Original; v != nil {
		db = db.Where("original = ?", *v)
	}
	if v := query.SendAtStart; v != nil {
		db = db.Where("send_at >= ?", *v)
	}
	if v := query.SendAtEnd; v != nil {
		db = db.Where("send_at <= ?", *v)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("send_at desc").Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	return list, total, err
}

// ListByUID 根据 UID 查询礼物
func (r *gormRepo) ListByUID(ctx context.Context, tx *gorm.DB, uid int64, limit int) ([]model.LiveGift, error) {
	db := r.getDB(ctx, tx)
	var list []model.LiveGift
	err := db.Where("uid = ?", uid).Order("send_at desc").Limit(limit).Find(&list).Error
	return list, err
}

// ListByLiveID 根据 LiveID 查询礼物
func (r *gormRepo) ListByLiveID(ctx context.Context, tx *gorm.DB, liveID int64, limit int) ([]model.LiveGift, error) {
	db := r.getDB(ctx, tx)
	var list []model.LiveGift
	err := db.Where("live_id = ?", liveID).Order("send_at desc").Limit(limit).Find(&list).Error
	return list, err
}

// UpdateLiveIDByRoomIDAndTimeRange 将指定房间在时间范围内的礼物关联到直播记录
func (r *gormRepo) UpdateLiveIDByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID, liveID int64) error {
	db := r.getDB(ctx, tx)
	return db.Model(&model.LiveGift{}).
		Where("room_id = ? AND send_at >= ? AND send_at <= ?", roomID, startTime, endTime).
		Update("live_id", liveID).Error
}

// CountAndRevenueByRoomIDAndTimeRange 统计指定房间在时间范围内的礼物数量与收益（price * num）
func (r *gormRepo) CountAndRevenueByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (count int64, revenue int64, err error) {
	db := r.getDB(ctx, tx)
	var result struct {
		Count   int64 `gorm:"column:count"`
		Revenue int64 `gorm:"column:revenue"`
	}
	err = db.Model(&model.LiveGift{}).
		Select("COUNT(*) AS count, COALESCE(SUM(price * num), 0) AS revenue").
		Where("room_id = ? AND send_at >= ? AND send_at <= ?", roomID, startTime, endTime).
		Scan(&result).Error
	return result.Count, result.Revenue, err
}

// CountGuardByRoomIDAndTimeRange 统计指定房间在时间范围内的大航海数量
func (r *gormRepo) CountGuardByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error) {
	db := r.getDB(ctx, tx)
	var count int64
	err := db.Model(&model.LiveGift{}).
		Where("room_id = ? AND send_at >= ? AND send_at <= ? AND gift_type = ?", roomID, startTime, endTime, enum.GiftTypeGuard).
		Count(&count).Error
	return count, err
}

// CountSuperChatByRoomIDAndTimeRange 统计指定房间在时间范围内的醒目留言数量
func (r *gormRepo) CountSuperChatByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error) {
	db := r.getDB(ctx, tx)
	var count int64
	err := db.Model(&model.LiveGift{}).
		Where("room_id = ? AND send_at >= ? AND send_at <= ? AND gift_type = ?", roomID, startTime, endTime, enum.GiftTypeSuperChat).
		Count(&count).Error
	return count, err
}

// SumBlindBoxProfit 统计盲盒盈利
func (r *gormRepo) SumBlindBoxProfit(ctx context.Context, tx *gorm.DB, uid, roomID int64, day, week, month TimeRange) (*BlindBoxProfit, error) {
	db := r.getDB(ctx, tx)
	db = db.Model(&model.LiveGift{}).Where("original = ?", enum.No)
	// 筛选用户或房间
	if uid > 0 {
		db = db.Where("uid = ?", uid)
	}
	if roomID > 0 {
		db = db.Where("room_id = ?", roomID)
	}
	var result BlindBoxProfit
	err := db.Select(`
		COALESCE(SUM(CASE WHEN send_at >= ? AND send_at <= ? THEN (price - original_gift_price) * num ELSE 0 END), 0) AS daily,
		COALESCE(SUM(CASE WHEN send_at >= ? AND send_at <= ? THEN (price - original_gift_price) * num ELSE 0 END), 0) AS weekly,
		COALESCE(SUM(CASE WHEN send_at >= ? AND send_at <= ? THEN (price - original_gift_price) * num ELSE 0 END), 0) AS monthly,
		COALESCE(SUM((price - original_gift_price) * num), 0) AS total`,
		day.Start, day.End,
		week.Start, week.End,
		month.Start, month.End,
	).Scan(&result).Error
	return &result, err
}

// escapeLike 转义 LIKE 查询中的特殊字符 _ %
func escapeLike(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '%' || r == '_' || r == '\\' {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
