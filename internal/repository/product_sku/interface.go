package product_sku

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.ProductSku]
	// ListByProductID 根据商品ID获取全部SKU，按 ID 升序
	ListByProductID(ctx context.Context, tx *gorm.DB, productID int64) ([]model.ProductSku, error)
	// IncrementStock 原子增加SKU库存
	IncrementStock(ctx context.Context, tx *gorm.DB, id, delta int64) error
	// DecrementStock 原子扣减SKU库存；库存不足（stock < delta）时不修改数据并返回 false
	DecrementStock(ctx context.Context, tx *gorm.DB, id, delta int64) (bool, error)
}

// ListByProductID 根据商品ID获取全部SKU
func (r *gormRepo) ListByProductID(ctx context.Context, tx *gorm.DB, productID int64) ([]model.ProductSku, error) {
	db := r.getDB(ctx, tx)
	var list []model.ProductSku
	err := db.Where("product_id = ?", productID).Order("id asc").Find(&list).Error
	return list, err
}

// IncrementStock 原子增加SKU库存
func (r *gormRepo) IncrementStock(ctx context.Context, tx *gorm.DB, id, delta int64) error {
	return r.IncrementField(ctx, tx, id, "stock", delta)
}

// DecrementStock 原子扣减SKU库存；库存不足（stock < delta）时不修改数据并返回 false
func (r *gormRepo) DecrementStock(ctx context.Context, tx *gorm.DB, id, delta int64) (bool, error) {
	res := r.getDB(ctx, tx).Model(&model.ProductSku{}).
		Where("id = ? AND stock >= ?", id, delta).
		Update("stock", gorm.Expr("stock - ?", delta))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
