package product_spec_value

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.ProductSpecValue]
	// ListByProductID 根据商品ID获取全部规格值，按 ID 升序
	ListByProductID(ctx context.Context, tx *gorm.DB, productID int64) ([]model.ProductSpecValue, error)
	// ListBySpecID 根据规格ID获取全部规格值，按 ID 升序
	ListBySpecID(ctx context.Context, tx *gorm.DB, specID int64) ([]model.ProductSpecValue, error)
}

// ListByProductID 根据商品ID获取全部规格值
func (r *gormRepo) ListByProductID(ctx context.Context, tx *gorm.DB, productID int64) ([]model.ProductSpecValue, error) {
	db := r.getDB(ctx, tx)
	var list []model.ProductSpecValue
	err := db.Where("product_id = ?", productID).Order("id asc").Find(&list).Error
	return list, err
}

// ListBySpecID 根据规格ID获取全部规格值
func (r *gormRepo) ListBySpecID(ctx context.Context, tx *gorm.DB, specID int64) ([]model.ProductSpecValue, error) {
	db := r.getDB(ctx, tx)
	var list []model.ProductSpecValue
	err := db.Where("product_spec_id = ?", specID).Order("id asc").Find(&list).Error
	return list, err
}
