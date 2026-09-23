package live_danmu

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/sqlutil"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
	"gorm.io/gorm"
)

// sortColumns 允许参与 ListPage 排序的 DB 列
var sortColumns = map[string]string{
	"id":         "id",
	"room_id":    "room_id",
	"uid":        "uid",
	"uname":      "uname",
	"msg":        "msg",
	"send_at":    "send_at",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

// sortOrder 允许的排序方向 → SQL 方向。
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

type Repository interface {
	base.Repository[model.LiveDanmu]
	// DistinctRoomIDs 获取全表中所有不重复的 RoomID
	DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error)
	// ListPage 分页查询弹幕，Uname/Msg 模糊匹配，SendAt 范围查询，按 SendAt 倒序
	ListPage(ctx context.Context, tx *gorm.DB, query model.LiveDanmuListPageQuery) ([]model.LiveDanmu, int64, error)
	// CountFiltered 统计命中行数。limit > 0 时只保证「不超过 limit」的语义，实现上用「子查询 + LIMIT limit」早停，避免在大表上做全量 COUNT
	CountFiltered(ctx context.Context, tx *gorm.DB, query model.LiveDanmuListPageQuery, limit int) (int64, error)
	// ExportChunk 导出用的分块读取：主键 < afterID（afterID 为 0 表示不限）且满足筛选条件，按主键倒序取至多 limit 行。
	ExportChunk(ctx context.Context, tx *gorm.DB, query model.LiveDanmuListPageQuery, afterID int64, limit int) ([]model.LiveDanmu, error)
	// ListByUID 根据 UID 查询弹幕，按 SendAt 倒序，limit 控制最大条数
	ListByUID(ctx context.Context, tx *gorm.DB, uid int64, limit int) ([]model.LiveDanmu, error)
	// ListByLiveID 根据 LiveID 查询弹幕，按 SendAt 倒序，limit 控制最大条数
	ListByLiveID(ctx context.Context, tx *gorm.DB, liveID int64, limit int) ([]model.LiveDanmu, error)
	// UpdateLiveIDByRoomIDAndTimeRange 将指定房间在时间范围内的弹幕批量关联到直播记录
	// 判据 room_id + 时间区间，可能命中多行；区间为闭闭 [startTime, endTime]
	UpdateLiveIDByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID, liveID int64) error
	// CountByRoomIDAndTimeRange 统计指定房间在时间范围内的弹幕数量
	CountByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error)
	// CountByUID 统计指定uid的弹幕数量
	CountByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error)
	// CountDailyByUID 根据uid统计用户在时间范围内的每日发言数量。
	// 返回的 map key 为「当月第几天」(1-31)，value 为当日发言数。
	// 区间为闭开 [startAt, endAt)（注意与 CountByRoomIDAndTimeRange 的闭闭区间不同）；
	// key 是日号而非日期，调用方需保证区间不跨月，否则不同月的同一天号会合并。
	CountDailyByUID(ctx context.Context, tx *gorm.DB, uid int64, startAt int64, endAt int64) (map[int64]int64, error)
	// ListMessagesByUID 获取指定用户的全部弹幕内容（msg 列），无排序、无条数上限
	ListMessagesByUID(ctx context.Context, tx *gorm.DB, uid int64) ([]string, error)
}

// DistinctRoomIDs 获取全表中所有不重复的 RoomID
func (r *gormRepo) DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error) {
	db := r.ResolveDB(ctx, tx)
	var roomIDs []int64
	if err := db.Model(&model.LiveDanmu{}).Distinct("room_id").Pluck("room_id", &roomIDs).Error; err != nil {
		return nil, err
	}
	return roomIDs, nil
}

// applyFilters 应用筛选条件。
//
// ListPage 与导出共用同一份条件拼装 —— 两边口径必须一致，否则导出结果与页面上看到的对不上。
func applyFilters(db *gorm.DB, query model.LiveDanmuListPageQuery) *gorm.DB {
	if v := query.RoomID; v != nil {
		db = db.Where("room_id = ?", *v)
	}
	if v := query.UID; v != nil {
		db = db.Where("uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("uname LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if v := query.Msg; v != nil && *v != "" {
		db = db.Where("msg LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if v := query.SendAtStart; v != nil {
		db = db.Where("send_at >= ?", *v)
	}
	if v := query.SendAtEnd; v != nil {
		db = db.Where("send_at <= ?", *v)
	}
	return db
}

// ListPage 分页查询弹幕
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LiveDanmuListPageQuery) ([]model.LiveDanmu, int64, error) {
	var list []model.LiveDanmu
	var total int64
	db := applyFilters(r.ResolveDB(ctx, tx).Model(&model.LiveDanmu{}), query)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 默认与现状一致；排序参数非法/缺失时静默回退该默认
	orderClause := "send_at desc"
	if query.SortField != nil && query.SortOrder != nil {
		// 方向先小写归一化再查白名单，非法值直接回退默认
		if field, ok := sortColumns[*query.SortField]; ok {
			if dir, ok := sortOrder[strings.ToLower(*query.SortOrder)]; ok {
				// field/dir 均来自字面量白名单，杜绝注入；id asc 保证同键值时翻页稳定
				orderClause = field + " " + dir + ", id asc"
			}
		}
	}
	err := db.Order(orderClause).Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	return list, total, err
}

// CountFiltered 统计命中行数，limit > 0 时提前停止
func (r *gormRepo) CountFiltered(ctx context.Context, tx *gorm.DB, query model.LiveDanmuListPageQuery, limit int) (int64, error) {
	db := r.ResolveDB(ctx, tx)
	sub := applyFilters(db.Model(&model.LiveDanmu{}).Select("1"), query)
	if limit > 0 {
		sub = sub.Limit(limit)
	}
	var total int64
	// 子查询包一层再 count：GORM 的 Count 会剥掉外层 Limit，必须用派生表把早停固定下来
	err := db.Table("(?) AS t", sub).Count(&total).Error
	return total, err
}

// ExportChunk 导出用的分块读取
func (r *gormRepo) ExportChunk(ctx context.Context, tx *gorm.DB, query model.LiveDanmuListPageQuery, afterID int64, limit int) ([]model.LiveDanmu, error) {
	var list []model.LiveDanmu
	db := applyFilters(r.ResolveDB(ctx, tx).Model(&model.LiveDanmu{}), query)
	if afterID > 0 {
		db = db.Where("id < ?", afterID)
	}
	err := db.Order("id desc").Limit(limit).Find(&list).Error
	return list, err
}

// ListByUID 根据 UID 查询弹幕
func (r *gormRepo) ListByUID(ctx context.Context, tx *gorm.DB, uid int64, limit int) ([]model.LiveDanmu, error) {
	db := r.ResolveDB(ctx, tx)
	var list []model.LiveDanmu
	err := db.Where("uid = ?", uid).Order("send_at desc").Limit(limit).Find(&list).Error
	return list, err
}

// ListByLiveID 根据 LiveID 查询弹幕
func (r *gormRepo) ListByLiveID(ctx context.Context, tx *gorm.DB, liveID int64, limit int) ([]model.LiveDanmu, error) {
	db := r.ResolveDB(ctx, tx)
	var list []model.LiveDanmu
	err := db.Where("live_id = ?", liveID).Order("send_at desc").Limit(limit).Find(&list).Error
	return list, err
}

// UpdateLiveIDByRoomIDAndTimeRange 将指定房间在时间范围内的弹幕批量关联到直播记录
func (r *gormRepo) UpdateLiveIDByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID, liveID int64) error {
	db := r.ResolveDB(ctx, tx)
	return db.Model(&model.LiveDanmu{}).
		Where("room_id = ? AND send_at >= ? AND send_at <= ?", roomID, startTime, endTime).
		Update("live_id", liveID).Error
}

// CountByRoomIDAndTimeRange 统计指定房间在时间范围内的弹幕数量
func (r *gormRepo) CountByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error) {
	db := r.ResolveDB(ctx, tx)
	var count int64
	err := db.Model(&model.LiveDanmu{}).
		Where("room_id = ? AND send_at >= ? AND send_at <= ?", roomID, startTime, endTime).
		Count(&count).Error
	return count, err
}

// CountByUID 统计指定uid的弹幕数量
func (r *gormRepo) CountByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error) {
	db := r.ResolveDB(ctx, tx)
	var total int64
	if err := db.Model(&model.LiveDanmu{}).Where("uid = ?", uid).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// CountDailyByUID 根据uid统计用户在时间范围内的每日发言数量，
// map key 为「当月第几天」(1-31)，value 为当日发言数
func (r *gormRepo) CountDailyByUID(ctx context.Context, tx *gorm.DB, uid int64, startAt int64, endAt int64) (map[int64]int64, error) {
	db := r.ResolveDB(ctx, tx)
	var sendAts []int64
	if err := db.Model(&model.LiveDanmu{}).Where("uid = ?", uid).Where("send_at >= ?", startAt).Where("send_at < ?", endAt).Pluck("send_at", &sendAts).Error; err != nil {
		return nil, err
	}
	result := make(map[int64]int64)
	for _, sendAt := range sendAts {
		day := timeutil.LocalDayOfMonth(sendAt)
		result[day]++
	}
	return result, nil
}

// ListMessagesByUID 获取指定用户的全部弹幕内容（msg 列），无排序、无条数上限
func (r *gormRepo) ListMessagesByUID(ctx context.Context, tx *gorm.DB, uid int64) ([]string, error) {
	db := r.ResolveDB(ctx, tx)
	var messages []string
	if err := db.Model(&model.LiveDanmu{}).Where("uid = ?", uid).Pluck("msg", &messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
