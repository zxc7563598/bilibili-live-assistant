package product

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.Product]
	// ListPage 分页查询商品列表，支持按名称模糊、积分类型、启用状态筛选，按排序值倒序
	ListPage(ctx context.Context, tx *gorm.DB, query model.ProductListPageQuery) ([]model.Product, int64, error)
	// ListEnabled 获取全部启用中的商品，按排序值倒序
	ListEnabled(ctx context.Context, tx *gorm.DB) ([]model.Product, error)
	// IncrementSold 原子增加商品销量
	IncrementSold(ctx context.Context, tx *gorm.DB, id, delta int64) error
	// IncrementStock 原子增加商品库存
	IncrementStock(ctx context.Context, tx *gorm.DB, id, delta int64) error
	// DecrementStock 原子扣减商品库存；库存不足（stock < delta）时不修改数据并返回 false
	DecrementStock(ctx context.Context, tx *gorm.DB, id, delta int64) (bool, error)
}

// ListPage 分页查询商品列表，支持按名称模糊、积分类型、启用状态筛选
// 未指定 Enable 时不限制状态（管理端可查看含下架在内的全部商品）；商城端只看启用商品应使用 ListEnabled
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.ProductListPageQuery) ([]model.Product, int64, error) {
	var list []model.Product
	var total int64
	db := r.getDB(ctx, tx)
	db = db.Model(&model.Product{})
	if v := query.Name; v != nil && *v != "" {
		db = db.Where("name LIKE ?", "%"+escapeLike(*v)+"%")
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
	err := db.Order("sort_order desc, id desc").Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListEnabled 获取全部启用中的商品
func (r *gormRepo) ListEnabled(ctx context.Context, tx *gorm.DB) ([]model.Product, error) {
	db := r.getDB(ctx, tx)
	var list []model.Product
	err := db.Where("enable = ?", enum.EnableEnable).Order("sort_order desc, id desc").Find(&list).Error
	return list, err
}

// IncrementSold 原子增加商品销量
func (r *gormRepo) IncrementSold(ctx context.Context, tx *gorm.DB, id, delta int64) error {
	return r.IncrementField(ctx, tx, id, "sold", delta)
}

// IncrementStock 原子增加商品库存
func (r *gormRepo) IncrementStock(ctx context.Context, tx *gorm.DB, id, delta int64) error {
	return r.IncrementField(ctx, tx, id, "stock", delta)
}

// DecrementStock 原子扣减商品库存；库存不足（stock < delta）时不修改数据并返回 false
func (r *gormRepo) DecrementStock(ctx context.Context, tx *gorm.DB, id, delta int64) (bool, error) {
	res := r.getDB(ctx, tx).Model(&model.Product{}).
		Where("id = ? AND stock >= ?", id, delta).
		Update("stock", gorm.Expr("stock - ?", delta))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
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
