package product_image

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.ProductImage]
	// ListByProductID 根据商品ID获取全部图片，按排序值倒序
	ListByProductID(ctx context.Context, tx *gorm.DB, productID int64) ([]model.ProductImage, error)
	// ListEnabledByProductID 根据商品ID获取全部启用图片，按排序值倒序
	ListEnabledByProductID(ctx context.Context, tx *gorm.DB, productID int64) ([]model.ProductImage, error)
}

// ListByProductID 根据商品ID获取全部图片
func (r *gormRepo) ListByProductID(ctx context.Context, tx *gorm.DB, productID int64) ([]model.ProductImage, error) {
	db := r.getDB(ctx, tx)
	var list []model.ProductImage
	err := db.Where("product_id = ?", productID).Order("sort_order desc, id asc").Find(&list).Error
	return list, err
}

// ListEnabledByProductID 根据商品ID获取全部启用图片
func (r *gormRepo) ListEnabledByProductID(ctx context.Context, tx *gorm.DB, productID int64) ([]model.ProductImage, error) {
	db := r.getDB(ctx, tx)
	var list []model.ProductImage
	err := db.Where("product_id = ?", productID).
		Where("enable = ?", enum.EnableEnable).
		Order("sort_order desc, id asc").Find(&list).Error
	return list, err
}
