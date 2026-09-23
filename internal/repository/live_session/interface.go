package live_session

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
	"gorm.io/gorm"
)

// sortColumns 允许参与 ListPage 排序的 DB 列
var sortColumns = map[string]string{
	"id":               "id",
	"room_id":          "room_id",
	"uid":              "uid",
	"live_key":         "live_key",
	"start_at":         "start_at",
	"start_source":     "start_source",
	"danmu_count":      "danmu_count",
	"gift_count":       "gift_count",
	"guard_count":      "guard_count",
	"super_chat_count": "super_chat_count",
	"total_revenue":    "total_revenue",
	"end_at":           "end_at",
	"end_detail":       "end_detail",
	"created_at":       "created_at",
	"updated_at":       "updated_at",
}

// sortOrder 允许的排序方向 → SQL 方向。
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

type Repository interface {
	base.Repository[model.LiveSession]
	// DistinctRoomIDs 获取全表中所有不重复的 RoomID
	DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error)
	// ListPage 分页查询直播场次，StartAt/EndAt 范围查询，按 StartAt 倒序
	ListPage(ctx context.Context, tx *gorm.DB, query model.LiveSessionListPageQuery) ([]model.LiveSession, int64, error)
	// UpdateStartByID 根据 ID 更新开播信息（StartAt / LivePlatform）
	UpdateStartByID(ctx context.Context, tx *gorm.DB, id int64, form model.LiveSessionUpdateStartForm) error
	// UpdateEndByID 根据 ID 更新下播信息（EndAt / EndReason / EndSource / EndDetail）
	UpdateEndByID(ctx context.Context, tx *gorm.DB, id int64, form model.LiveSessionUpdateEndForm) error
	// UpdateStatsByID 根据 ID 更新统计数据（DanmuCount / GiftCount / GuardCount / SuperChatCount / TotalRevenue）
	UpdateStatsByID(ctx context.Context, tx *gorm.DB, id int64, form model.LiveSessionUpdateStatsForm) error
	// ListActive 获取所有未下播的记录（判据是 end_at = 0），按 StartAt 升序
	ListActive(ctx context.Context, tx *gorm.DB) ([]model.LiveSession, error)
	// ListActiveByRoomID 获取指定房间所有未下播的记录（判据是 end_at = 0），按 StartAt 升序
	ListActiveByRoomID(ctx context.Context, tx *gorm.DB, roomID int64) ([]model.LiveSession, error)
	// DistinctLiveDays 统计时间范围内有哪些天开播过（不区分主播/房间，通常只有一个主播），
	// 返回「当月第几天」(1-31) 的集合作为 key；时间区间为闭开 [startAt, endAt)。
	// 注意 key 是日号而非日期，调用方需保证查询区间不跨月，否则不同月的同一天号会合并。
	DistinctLiveDays(ctx context.Context, tx *gorm.DB, startAt int64, endAt int64) (map[int64]struct{}, error)
}

// DistinctRoomIDs 获取全表中所有不重复的 RoomID
func (r *gormRepo) DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error) {
	db := r.ResolveDB(ctx, tx)
	var roomIDs []int64
	if err := db.Model(&model.LiveSession{}).Distinct("room_id").Pluck("room_id", &roomIDs).Error; err != nil {
		return nil, err
	}
	return roomIDs, nil
}

// ListPage 分页查询直播场次
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LiveSessionListPageQuery) ([]model.LiveSession, int64, error) {
	var list []model.LiveSession
	var total int64
	db := r.ResolveDB(ctx, tx)
	db = db.Model(&model.LiveSession{})
	if v := query.RoomID; v != nil {
		db = db.Where("room_id = ?", *v)
	}
	if v := query.UID; v != nil {
		db = db.Where("uid = ?", *v)
	}
	if v := query.StartAtStart; v != nil {
		db = db.Where("start_at >= ?", *v)
	}
	if v := query.StartAtEnd; v != nil {
		db = db.Where("start_at <= ?", *v)
	}
	if v := query.EndAtStart; v != nil {
		db = db.Where("end_at >= ?", *v)
	}
	if v := query.EndAtEnd; v != nil {
		db = db.Where("end_at <= ?", *v)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 默认与现状一致；排序参数非法/缺失时静默回退该默认
	orderClause := "start_at desc"
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

// UpdateStartByID 根据 ID 更新开播信息
func (r *gormRepo) UpdateStartByID(ctx context.Context, tx *gorm.DB, id int64, form model.LiveSessionUpdateStartForm) error {
	updateMap := make(map[string]any, 2)
	if v := form.StartAt; v != nil {
		updateMap["start_at"] = *v
	}
	if v := form.LivePlatform; v != nil && *v != "" {
		updateMap["live_platform"] = *v
	}
	if len(updateMap) == 0 {
		return nil
	}
	return r.UpdateMap(ctx, tx, "id", id, updateMap)
}

// UpdateEndByID 根据 ID 更新下播信息
func (r *gormRepo) UpdateEndByID(ctx context.Context, tx *gorm.DB, id int64, form model.LiveSessionUpdateEndForm) error {
	updateMap := make(map[string]any, 4)
	if v := form.EndAt; v != nil {
		updateMap["end_at"] = *v
	}
	if v := form.EndReason; v != nil {
		updateMap["end_reason"] = *v
	}
	if v := form.EndSource; v != nil {
		updateMap["end_source"] = *v
	}
	if v := form.EndDetail; v != nil && *v != "" {
		updateMap["end_detail"] = *v
	}
	if len(updateMap) == 0 {
		return nil
	}
	return r.UpdateMap(ctx, tx, "id", id, updateMap)
}

// UpdateStatsByID 根据 ID 更新统计数据
func (r *gormRepo) UpdateStatsByID(ctx context.Context, tx *gorm.DB, id int64, form model.LiveSessionUpdateStatsForm) error {
	updateMap := make(map[string]any, 5)
	if v := form.DanmuCount; v != nil {
		updateMap["danmu_count"] = *v
	}
	if v := form.GiftCount; v != nil {
		updateMap["gift_count"] = *v
	}
	if v := form.GuardCount; v != nil {
		updateMap["guard_count"] = *v
	}
	if v := form.SuperChatCount; v != nil {
		updateMap["super_chat_count"] = *v
	}
	if v := form.TotalRevenue; v != nil {
		updateMap["total_revenue"] = *v
	}
	if len(updateMap) == 0 {
		return nil
	}
	return r.UpdateMap(ctx, tx, "id", id, updateMap)
}

// ListActive 获取所有未下播的记录（EndAt = 0），按 StartAt 升序
func (r *gormRepo) ListActive(ctx context.Context, tx *gorm.DB) ([]model.LiveSession, error) {
	db := r.ResolveDB(ctx, tx)
	var list []model.LiveSession
	err := db.Where("end_at = 0").Order("start_at asc").Find(&list).Error
	return list, err
}

// ListActiveByRoomID 获取指定房间所有未下播的记录（EndAt = 0），按 StartAt 升序
func (r *gormRepo) ListActiveByRoomID(ctx context.Context, tx *gorm.DB, roomID int64) ([]model.LiveSession, error) {
	db := r.ResolveDB(ctx, tx)
	var list []model.LiveSession
	err := db.Where("end_at = 0 AND room_id = ?", roomID).Order("start_at asc").Find(&list).Error
	return list, err
}

// DistinctLiveDays 统计时间范围内有哪些天开播过（不区分主播/房间，通常只有一个主播）
// 返回「当月第几天」(1-31) 的集合；时间区间为闭开 [startAt, endAt)
func (r *gormRepo) DistinctLiveDays(ctx context.Context, tx *gorm.DB, startAt int64, endAt int64) (map[int64]struct{}, error) {
	db := r.ResolveDB(ctx, tx)
	var startAts []int64
	if err := db.Model(&model.LiveSession{}).
		Where("start_at >= ?", startAt).
		Where("start_at < ?", endAt).
		Pluck("start_at", &startAts).Error; err != nil {
		return nil, err
	}
	result := make(map[int64]struct{})
	for _, start := range startAts {
		day := timeutil.LocalDayOfMonth(start)
		result[day] = struct{}{}
	}
	return result, nil
}
