package product_spec

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.ProductSpec]
	// ListByProductID 根据商品ID获取全部规格，按 ID 升序
	ListByProductID(ctx context.Context, tx *gorm.DB, productID int64) ([]model.ProductSpec, error)
}

// ListByProductID 根据商品ID获取全部规格
func (r *gormRepo) ListByProductID(ctx context.Context, tx *gorm.DB, productID int64) ([]model.ProductSpec, error) {
	db := r.getDB(ctx, tx)
	var list []model.ProductSpec
	err := db.Where("product_id = ?", productID).Order("id asc").Find(&list).Error
	return list, err
}
