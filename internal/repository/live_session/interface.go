package live_session

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

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
	// ListActive 获取所有未下播的记录
	ListActive(ctx context.Context, tx *gorm.DB) ([]model.LiveSession, error)
}

// DistinctRoomIDs 获取全表中所有不重复的 RoomID
func (r *gormRepo) DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error) {
	db := r.getDB(ctx, tx)
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
	db := r.getDB(ctx, tx)
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
	err := db.Order("start_at desc").Offset(query.Offset).Limit(query.Limit).Find(&list).Error
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
	db := r.getDB(ctx, tx)
	var list []model.LiveSession
	err := db.Where("end_at = 0").Order("start_at asc").Find(&list).Error
	return list, err
}
