package product_sku_stock_log

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.ProductSkuStockLog]
	// ListByProductSkuID 根据SKU ID获取库存流水，按创建时间倒序，limit 控制最大条数
	ListByProductSkuID(ctx context.Context, tx *gorm.DB, skuID int64, limit int) ([]model.ProductSkuStockLog, error)
	// ListByProductID 根据商品ID获取库存流水，按创建时间倒序，limit 控制最大条数
	ListByProductID(ctx context.Context, tx *gorm.DB, productID int64, limit int) ([]model.ProductSkuStockLog, error)
}

// ListByProductSkuID 根据SKU ID获取库存流水
func (r *gormRepo) ListByProductSkuID(ctx context.Context, tx *gorm.DB, skuID int64, limit int) ([]model.ProductSkuStockLog, error) {
	db := r.getDB(ctx, tx)
	var list []model.ProductSkuStockLog
	err := db.Where("product_sku_id = ?", skuID).Order("created_at desc").Limit(limit).Find(&list).Error
	return list, err
}

// ListByProductID 根据商品ID获取库存流水
func (r *gormRepo) ListByProductID(ctx context.Context, tx *gorm.DB, productID int64, limit int) ([]model.ProductSkuStockLog, error) {
	db := r.getDB(ctx, tx)
	var list []model.ProductSkuStockLog
	err := db.Where("product_id = ?", productID).Order("created_at desc").Limit(limit).Find(&list).Error
	return list, err
}
