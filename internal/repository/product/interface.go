package product

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/sqlutil"
	"gorm.io/gorm"
)

// sortColumns 允许参与 ListPage 排序的 DB 列
var sortColumns = map[string]string{
	"id":           "id",
	"name":         "name",
	"price":        "price",
	"credit_type":  "credit_type",
	"product_type": "product_type",
	"sold":         "sold",
	"stock":        "stock",
	"sort_order":   "sort_order",
	"enable":       "enable",
	"created_at":   "created_at",
	"updated_at":   "updated_at",
}

// sortOrder 允许的排序方向 → SQL 方向。
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

type Repository interface {
	base.Repository[model.Product]
	// ListPage 分页查询商品列表，支持按名称模糊、积分类型、启用状态筛选，按排序值倒序
	ListPage(ctx context.Context, tx *gorm.DB, query model.ProductListPageQuery) ([]model.Product, int64, error)
	// ListEnabled 获取全部启用中的商品，按排序值倒序
	ListEnabled(ctx context.Context, tx *gorm.DB) ([]model.Product, error)
	// AdjustStock 原子增减商品库存，delta 可为负
	AdjustStock(ctx context.Context, tx *gorm.DB, id, delta int64) error
	// DecrementStock 原子扣减商品库存；库存不足（stock < delta）时不修改数据并返回 false
	DecrementStock(ctx context.Context, tx *gorm.DB, id, delta int64) (bool, error)
}

// ListPage 分页查询商品列表，支持按名称模糊、积分类型、启用状态筛选
// 未指定 Enable 时不限制状态（管理端可查看含下架在内的全部商品）；商城端只看启用商品应使用 ListEnabled
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.ProductListPageQuery) ([]model.Product, int64, error) {
	var list []model.Product
	var total int64
	db := r.ResolveDB(ctx, tx)
	db = db.Model(&model.Product{})
	if v := query.Name; v != nil && *v != "" {
		db = db.Where("name LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if v := query.CreditType; v != nil {
		ct := enum.CreditType(*v)
		if ct.IsValid() {
			db = db.Where("credit_type = ?", ct)
		}
	}
	if v := query.Enable; v != nil {
		e := enum.Enable(*v)
		if e.IsValid() {
			db = db.Where("enable = ?", e)
		}
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 默认与现状一致；排序参数非法/缺失时静默回退该默认
	orderClause := "sort_order desc, id desc"
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
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListEnabled 获取全部启用中的商品
func (r *gormRepo) ListEnabled(ctx context.Context, tx *gorm.DB) ([]model.Product, error) {
	db := r.ResolveDB(ctx, tx)
	var list []model.Product
	err := db.Where("enable = ?", enum.EnableEnable).Order("sort_order desc, id desc").Find(&list).Error
	return list, err
}

// AdjustStock 原子增减商品库存，delta 可为负
func (r *gormRepo) AdjustStock(ctx context.Context, tx *gorm.DB, id, delta int64) error {
	return r.AdjustField(ctx, tx, id, "stock", delta)
}

// DecrementStock 原子扣减商品库存；库存不足（stock < delta）时不修改数据并返回 false
func (r *gormRepo) DecrementStock(ctx context.Context, tx *gorm.DB, id, delta int64) (bool, error) {
	res := r.ResolveDB(ctx, tx).Model(&model.Product{}).
		Where("id = ? AND stock >= ?", id, delta).
		Update("stock", gorm.Expr("stock - ?", delta))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
