package live_user_order_draft

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.LiveUserOrderDraft]
	// ListByUserID 根据用户ID获取全部订单草稿，按创建时间倒序
	ListByUserID(ctx context.Context, tx *gorm.DB, userID int64) ([]model.LiveUserOrderDraft, error)
}

// ListByUserID 根据用户ID获取全部订单草稿
func (r *gormRepo) ListByUserID(ctx context.Context, tx *gorm.DB, userID int64) ([]model.LiveUserOrderDraft, error) {
	db := r.getDB(ctx, tx)
	var list []model.LiveUserOrderDraft
	err := db.Where("user_id = ?", userID).Order("created_at desc, id desc").Find(&list).Error
	return list, err
}
