package live_user_order

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.LiveUserOrder]
	// ListPage 分页查询用户订单，按 OrderStatus 筛选，按创建时间倒序
	ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserOrderListPageQuery) ([]model.LiveUserOrder, int64, error)
	// GetByOrderSn 根据订单号获取单条订单
	GetByOrderSn(ctx context.Context, tx *gorm.DB, orderSn string) (*model.LiveUserOrder, error)
}

// ListPage 分页查询用户订单
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserOrderListPageQuery) ([]model.LiveUserOrder, int64, error) {
	var list []model.LiveUserOrder
	var total int64
	db := r.getDB(ctx, tx).Model(&model.LiveUserOrder{}).Where("user_id = ?", query.UserID)
	if v := query.OrderStatus; v != nil {
		s := enum.OrderStatus(*v)
		if s.IsValid() {
			db = db.Where("order_status = ?", s)
		}
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("created_at desc, id desc").Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	return list, total, err
}

// GetByOrderSn 根据订单号获取单条订单
func (r *gormRepo) GetByOrderSn(ctx context.Context, tx *gorm.DB, orderSn string) (*model.LiveUserOrder, error) {
	return r.FindOneByField(ctx, tx, "order_sn", orderSn)
}
