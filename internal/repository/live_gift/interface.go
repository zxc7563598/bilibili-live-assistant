package live_gift

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

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
