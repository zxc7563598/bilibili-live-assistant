package live_user_order_draft

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.LiveUserOrderDraft]
	// ListByUserID 根据用户ID获取全部订单草稿，按创建时间倒序
	ListByUserID(ctx context.Context, tx *gorm.DB, userID int64) ([]model.LiveUserOrderDraft, error)
	// GetActiveByUserID 获取用户当前 Active 状态的草稿（至多一条；无则返回 nil）
	GetActiveByUserID(ctx context.Context, tx *gorm.DB, userID int64) (*model.LiveUserOrderDraft, error)
	// CancelActiveByID 将 Active 草稿置为 Cancelled，返回是否发生变更（幂等）
	CancelActiveByID(ctx context.Context, tx *gorm.DB, id int64) (bool, error)
	// ListActiveExpired 查询已过期但仍为 Active 的草稿，按到期时间升序
	ListActiveExpired(ctx context.Context, tx *gorm.DB, now int64) ([]model.LiveUserOrderDraft, error)
}

// ListByUserID 根据用户ID获取全部订单草稿
func (r *gormRepo) ListByUserID(ctx context.Context, tx *gorm.DB, userID int64) ([]model.LiveUserOrderDraft, error) {
	db := r.getDB(ctx, tx)
	var list []model.LiveUserOrderDraft
	err := db.Where("user_id = ?", userID).Order("created_at desc, id desc").Find(&list).Error
	return list, err
}

// GetActiveByUserID 获取用户当前 Active 状态的草稿
func (r *gormRepo) GetActiveByUserID(ctx context.Context, tx *gorm.DB, userID int64) (*model.LiveUserOrderDraft, error) {
	db := r.getDB(ctx, tx)
	var draft model.LiveUserOrderDraft
	err := db.Where("user_id = ? AND status = ?", userID, enum.DraftStatusActive).First(&draft).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

// CancelActiveByID 将 Active 草稿置为 Cancelled，返回是否发生变更
func (r *gormRepo) CancelActiveByID(ctx context.Context, tx *gorm.DB, id int64) (bool, error) {
	res := r.getDB(ctx, tx).Model(&model.LiveUserOrderDraft{}).
		Where("id = ? AND status = ?", id, enum.DraftStatusActive).
		Update("status", enum.DraftStatusCancelled)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// ListActiveExpired 查询已过期但仍为 Active 的草稿
func (r *gormRepo) ListActiveExpired(ctx context.Context, tx *gorm.DB, now int64) ([]model.LiveUserOrderDraft, error) {
	db := r.getDB(ctx, tx)
	var list []model.LiveUserOrderDraft
	err := db.Where("status = ?", enum.DraftStatusActive).
		Where("expire_at <= ?", now).
		Order("expire_at asc").
		Find(&list).Error
	return list, err
}
