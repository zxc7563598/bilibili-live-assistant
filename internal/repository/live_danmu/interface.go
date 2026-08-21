package live_danmu

import (
	"context"
	"strings"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.LiveDanmu]
	// DistinctRoomIDs 获取全表中所有不重复的 RoomID
	DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error)
	// ListPage 分页查询弹幕，Uname/Msg 模糊匹配，SendAt 范围查询，按 SendAt 倒序
	ListPage(ctx context.Context, tx *gorm.DB, query model.LiveDanmuListPageQuery) ([]model.LiveDanmu, int64, error)
	// ListByUID 根据 UID 查询弹幕，按 SendAt 倒序，limit 控制最大条数
	ListByUID(ctx context.Context, tx *gorm.DB, uid int64, limit int) ([]model.LiveDanmu, error)
	// ListByLiveID 根据 LiveID 查询弹幕，按 SendAt 倒序，limit 控制最大条数
	ListByLiveID(ctx context.Context, tx *gorm.DB, liveID int64, limit int) ([]model.LiveDanmu, error)
	// UpdateLiveIDByRoomIDAndTimeRange 将指定房间在时间范围内的弹幕关联到直播记录
	UpdateLiveIDByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID, liveID int64) error
	// CountByRoomIDAndTimeRange 统计指定房间在时间范围内的弹幕数量
	CountByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error)
	// CountByUID 统计指定uid的弹幕数量
	CountByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error)
	// CountDailyByUID 根据uid统计用户在时间范围内的每日发言数量。
	// 返回的 map key 为「当月第几天」(1-31)，value 为当日发言数。
	CountDailyByUID(ctx context.Context, tx *gorm.DB, uid int64, startAt int64, endAt int64) (map[int64]int64, error)
	// GetMessagesByUID 根据uid获取用户全部弹幕信息
	GetMessagesByUID(ctx context.Context, tx *gorm.DB, uid int64) ([]string, error)
}

// DistinctRoomIDs 获取全表中所有不重复的 RoomID
func (r *gormRepo) DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error) {
	db := r.getDB(ctx, tx)
	var roomIDs []int64
	if err := db.Model(&model.LiveDanmu{}).Distinct("room_id").Pluck("room_id", &roomIDs).Error; err != nil {
		return nil, err
	}
	return roomIDs, nil
}

// ListPage 分页查询弹幕
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LiveDanmuListPageQuery) ([]model.LiveDanmu, int64, error) {
	var list []model.LiveDanmu
	var total int64
	db := r.getDB(ctx, tx)
	db = db.Model(&model.LiveDanmu{})
	if v := query.RoomID; v != nil {
		db = db.Where("room_id = ?", *v)
	}
	if v := query.UID; v != nil {
		db = db.Where("uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("uname LIKE ?", "%"+escapeLike(*v)+"%")
	}
	if v := query.Msg; v != nil && *v != "" {
		db = db.Where("msg LIKE ?", "%"+escapeLike(*v)+"%")
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

// ListByUID 根据 UID 查询弹幕
func (r *gormRepo) ListByUID(ctx context.Context, tx *gorm.DB, uid int64, limit int) ([]model.LiveDanmu, error) {
	db := r.getDB(ctx, tx)
	var list []model.LiveDanmu
	err := db.Where("uid = ?", uid).Order("send_at desc").Limit(limit).Find(&list).Error
	return list, err
}

// ListByLiveID 根据 LiveID 查询弹幕
func (r *gormRepo) ListByLiveID(ctx context.Context, tx *gorm.DB, liveID int64, limit int) ([]model.LiveDanmu, error) {
	db := r.getDB(ctx, tx)
	var list []model.LiveDanmu
	err := db.Where("live_id = ?", liveID).Order("send_at desc").Limit(limit).Find(&list).Error
	return list, err
}

// UpdateLiveIDByRoomIDAndTimeRange 将指定房间在时间范围内的弹幕关联到直播记录
func (r *gormRepo) UpdateLiveIDByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID, liveID int64) error {
	db := r.getDB(ctx, tx)
	return db.Model(&model.LiveDanmu{}).
		Where("room_id = ? AND send_at >= ? AND send_at <= ?", roomID, startTime, endTime).
		Update("live_id", liveID).Error
}

// CountByRoomIDAndTimeRange 统计指定房间在时间范围内的弹幕数量
func (r *gormRepo) CountByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error) {
	db := r.getDB(ctx, tx)
	var count int64
	err := db.Model(&model.LiveDanmu{}).
		Where("room_id = ? AND send_at >= ? AND send_at <= ?", roomID, startTime, endTime).
		Count(&count).Error
	return count, err
}

// CountByUID 统计指定uid的弹幕数量
func (r *gormRepo) CountByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error) {
	db := r.getDB(ctx, tx)
	var total int64
	if err := db.Model(&model.LiveDanmu{}).Where("uid = ?", uid).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// CountDailyByUID 根据uid统计用户在时间范围内的每日发言数量，
// map key 为「当月第几天」(1-31)，value 为当日发言数
func (r *gormRepo) CountDailyByUID(ctx context.Context, tx *gorm.DB, uid int64, startAt int64, endAt int64) (map[int64]int64, error) {
	db := r.getDB(ctx, tx)
	var sendAts []int64
	if err := db.Model(&model.LiveDanmu{}).Where("uid = ?", uid).Where("send_at >= ?", startAt).Where("send_at < ?", endAt).Pluck("send_at", &sendAts).Error; err != nil {
		return nil, err
	}
	result := make(map[int64]int64)
	for _, sendAt := range sendAts {
		day := int64(time.Unix(sendAt, 0).In(time.Local).Day())
		result[day]++
	}
	return result, nil
}

// GetMessagesByUID 根据uid获取用户全部弹幕信息
func (r *gormRepo) GetMessagesByUID(ctx context.Context, tx *gorm.DB, uid int64) ([]string, error) {
	db := r.getDB(ctx, tx)
	var messages []string
	if err := db.Model(&model.LiveDanmu{}).Where("uid = ?", uid).Pluck("msg", &messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
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
