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
	// ListPage 分页查询启用中的商品，Name 模糊匹配，按排序值倒序
	ListPage(ctx context.Context, tx *gorm.DB, query model.ProductListPageQuery) ([]model.Product, int64, error)
	// ListEnabled 获取全部启用中的商品，按排序值倒序
	ListEnabled(ctx context.Context, tx *gorm.DB) ([]model.Product, error)
	// IncrementSold 原子增加商品销量
	IncrementSold(ctx context.Context, tx *gorm.DB, id, delta int64) error
	// IncrementStock 原子增加商品库存
	IncrementStock(ctx context.Context, tx *gorm.DB, id, delta int64) error
}

// ListPage 分页查询启用中的商品
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.ProductListPageQuery) ([]model.Product, int64, error) {
	var list []model.Product
	var total int64
	db := r.getDB(ctx, tx).Model(&model.Product{}).Where("enable = ?", enum.EnableEnable)
	if v := query.Name; v != nil && *v != "" {
		db = db.Where("name LIKE ?", "%"+escapeLike(*v)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("sort_order desc, id desc").Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	return list, total, err
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
