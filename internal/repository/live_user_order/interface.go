package live_user_order

import (
	"context"
	"errors"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/sqlutil"
	"gorm.io/gorm"
)

var sortColumns = map[string]string{
	"id":                   "live_user_orders.id",
	"user_id":              "live_user_orders.user_id",
	"order_sn":             "live_user_orders.order_sn",
	"product_id":           "live_user_orders.product_id",
	"product_sku_id":       "live_user_orders.product_sku_id",
	"product_name":         "live_user_orders.product_name",
	"quantity":             "live_user_orders.quantity",
	"credit_type":          "live_user_orders.credit_type",
	"price":                "live_user_orders.price",
	"receiver_name":        "live_user_orders.receiver_name",
	"receiver_phone":       "live_user_orders.receiver_phone",
	"receiver_region_code": "live_user_orders.receiver_region_code",
	"receiver_email":       "live_user_orders.receiver_email",
	"order_status":         "live_user_orders.order_status",
	"pay_status":           "live_user_orders.pay_status",
	"ship_status":          "live_user_orders.ship_status",
	"express_company":      "live_user_orders.express_company",
	"express_no":           "live_user_orders.express_no",
	"pay_at":               "live_user_orders.pay_at",
	"processed_at":         "live_user_orders.processed_at",
	"cancel_at":            "live_user_orders.cancel_at",
	"created_at":           "live_user_orders.created_at",
	"updated_at":           "live_user_orders.updated_at",
	"uid":                  "lu.uid",
	"uname":                "lu.uname",
}

// sortOrder 允许的排序方向 → SQL 方向。
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

type Repository interface {
	base.Repository[model.LiveUserOrder]
	// ListPage 分页查询订单，联查 live_users 补充 uid/uname，
	// 支持 UserID/UID/Uname/OrderSn/OrderStatus/PayStatus/ShipStatus 筛选与白名单字段排序
	ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserOrderListPageQuery) ([]model.LiveUserOrderListItem, int64, error)
	// CountFiltered 统计命中行数。limit > 0 时只保证「不超过 limit」的语义，
	// 实现上用「子查询 + LIMIT limit」早停，避免在大表上做全量 COUNT。
	// 口径与 ListPage 一致（含 live_users 的 LEFT JOIN）
	CountFiltered(ctx context.Context, tx *gorm.DB, query model.LiveUserOrderListPageQuery, limit int) (int64, error)
	// ExportChunk 导出用的分块读取：主键 < afterID（afterID 为 0 表示不限）且满足筛选条件，
	// 按主键倒序取至多 limit 行；口径与 ListPage 一致（含 live_users 的 LEFT JOIN）
	ExportChunk(ctx context.Context, tx *gorm.DB, query model.LiveUserOrderListPageQuery, afterID int64, limit int) ([]model.LiveUserOrderListItem, error)
	// GetByOrderSn 根据订单号获取单条订单
	GetByOrderSn(ctx context.Context, tx *gorm.DB, orderSn string) (*model.LiveUserOrder, error)
	// GetDetailByID 按主键查询订单详情，联查 live_users 补充 uid/uname；不存在返回 (nil, nil)
	GetDetailByID(ctx context.Context, tx *gorm.DB, id int64) (*model.LiveUserOrderListItem, error)
}

// orderBase 订单查询的基准：联查 live_users 补充 uid/uname。
//
// 订单是历史快照、用户可能被移除，用 LEFT JOIN 保证订单不丢；
// JOIN 打在 live_users 主键上不放大行数，count(*) 依然精确，无需 DISTINCT。
// lu 侧不过滤 deleted_at 是有意为之（对齐 live_user_credit_log 的处理）。
//
// 列表 / 计数 / 导出三个入口都必须用这一份：漏掉 Select + Joins 会让 uid/uname 全为
// 零值，而过滤条件里的 lu.uid 还会引用到未声明的别名，直接报 SQL 错。
func (r *gormRepo) orderBase(ctx context.Context, tx *gorm.DB) *gorm.DB {
	return r.ResolveDB(ctx, tx).Model(&model.LiveUserOrder{}).
		Select("live_user_orders.*, lu.uid, lu.uname").
		Joins("LEFT JOIN live_users lu ON lu.id = live_user_orders.user_id")
}

// applyOrderFilters 应用筛选条件，与 ListPage 共用同一份拼装
func (r *gormRepo) applyOrderFilters(db *gorm.DB, query model.LiveUserOrderListPageQuery) *gorm.DB {
	if v := query.UserID; v != nil {
		db = db.Where("live_user_orders.user_id = ?", *v)
	}
	if v := query.UID; v != nil {
		db = db.Where("lu.uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("lu.uname LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if v := query.OrderSn; v != nil && *v != "" {
		db = db.Where("live_user_orders.order_sn LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if v := query.OrderStatus; v != nil {
		s := enum.OrderStatus(*v)
		if s.IsValid() {
			db = db.Where("live_user_orders.order_status = ?", s)
		}
	}
	if v := query.PayStatus; v != nil {
		s := enum.PayStatus(*v)
		if s.IsValid() {
			db = db.Where("live_user_orders.pay_status = ?", s)
		}
	}
	if v := query.ShipStatus; v != nil {
		s := enum.ShipStatus(*v)
		if s.IsValid() {
			db = db.Where("live_user_orders.ship_status = ?", s)
		}
	}
	return db
}

// ListPage 分页查询订单，联查 live_users 补充 uid/uname
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserOrderListPageQuery) ([]model.LiveUserOrderListItem, int64, error) {
	var list []model.LiveUserOrderListItem
	var total int64
	db := r.applyOrderFilters(r.orderBase(ctx, tx), query)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 默认与现状一致；排序参数非法/缺失时静默回退该默认
	orderClause := "live_user_orders.created_at desc, live_user_orders.id desc"
	if query.SortField != nil && query.SortOrder != nil {
		// 方向先小写归一化再查白名单，非法值直接回退默认
		if field, ok := sortColumns[*query.SortField]; ok {
			if dir, ok := sortOrder[strings.ToLower(*query.SortOrder)]; ok {
				// field/dir 均来自字面量白名单，杜绝注入；id asc 保证同键值时翻页稳定
				orderClause = field + " " + dir + ", live_user_orders.id asc"
			}
		}
	}
	err := db.Order(orderClause).Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	return list, total, err
}

// CountFiltered 统计命中行数，limit > 0 时提前停止
func (r *gormRepo) CountFiltered(ctx context.Context, tx *gorm.DB, query model.LiveUserOrderListPageQuery, limit int) (int64, error) {
	db := r.ResolveDB(ctx, tx)
	sub := r.applyOrderFilters(r.orderBase(ctx, tx), query)
	if limit > 0 {
		sub = sub.Limit(limit)
	}
	var total int64
	// 子查询包一层再 count：GORM 的 Count 会剥掉外层 Limit，必须用派生表把早停固定下来
	err := db.Table("(?) AS t", sub).Count(&total).Error
	return total, err
}

// ExportChunk 导出用的分块读取。
//
// 游标走 live_user_orders.id：JOIN 打在 live_users 主键上是一对一、不放大行数，
// 所以按主键递减分块不会跳行也不会重复。
func (r *gormRepo) ExportChunk(ctx context.Context, tx *gorm.DB, query model.LiveUserOrderListPageQuery, afterID int64, limit int) ([]model.LiveUserOrderListItem, error) {
	var list []model.LiveUserOrderListItem
	db := r.applyOrderFilters(r.orderBase(ctx, tx), query)
	if afterID > 0 {
		db = db.Where("live_user_orders.id < ?", afterID)
	}
	err := db.Order("live_user_orders.id desc").Limit(limit).Find(&list).Error
	return list, err
}

// GetDetailByID 按主键查询订单详情，联查 live_users 补充 uid/uname；不存在返回 (nil, nil)
func (r *gormRepo) GetDetailByID(ctx context.Context, tx *gorm.DB, id int64) (*model.LiveUserOrderListItem, error) {
	var item model.LiveUserOrderListItem
	err := r.ResolveDB(ctx, tx).Model(&model.LiveUserOrder{}).
		Select("live_user_orders.*, lu.uid, lu.uname").
		Joins("LEFT JOIN live_users lu ON lu.id = live_user_orders.user_id").
		Where("live_user_orders.id = ?", id).
		Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetByOrderSn 根据订单号获取单条订单
func (r *gormRepo) GetByOrderSn(ctx context.Context, tx *gorm.DB, orderSn string) (*model.LiveUserOrder, error) {
	return r.GetByField(ctx, tx, "order_sn", orderSn)
}
