package live_user_order

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

// sortColumns 允许参与 ListPage 排序的 DB 列
var sortColumns = map[string]string{
	"id":                   "id",
	"user_id":              "user_id",
	"order_sn":             "order_sn",
	"product_id":           "product_id",
	"product_sku_id":       "product_sku_id",
	"product_name":         "product_name",
	"quantity":             "quantity",
	"credit_type":          "credit_type",
	"price":                "price",
	"receiver_name":        "receiver_name",
	"receiver_phone":       "receiver_phone",
	"receiver_region_code": "receiver_region_code",
	"receiver_email":       "receiver_email",
	"order_status":         "order_status",
	"pay_status":           "pay_status",
	"ship_status":          "ship_status",
	"express_company":      "express_company",
	"express_no":           "express_no",
	"pay_at":               "pay_at",
	"processed_at":         "processed_at",
	"cancel_at":            "cancel_at",
	"created_at":           "created_at",
	"updated_at":           "updated_at",
}

// sortOrder 允许的排序方向 → SQL 方向。
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

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
	// 默认与现状一致；排序参数非法/缺失时静默回退该默认
	orderClause := "created_at desc, id desc"
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

// GetByOrderSn 根据订单号获取单条订单
func (r *gormRepo) GetByOrderSn(ctx context.Context, tx *gorm.DB, orderSn string) (*model.LiveUserOrder, error) {
	return r.FindOneByField(ctx, tx, "order_sn", orderSn)
}
