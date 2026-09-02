package product

import (
	"context"
	"errors"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product_image"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product_sku"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product_sku_stock_log"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product_spec"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product_spec_value"
	"gorm.io/gorm"
)

type Service struct {
	db                     *gorm.DB
	productRepo            product.Repository
	productSkuRepo         product_sku.Repository
	productSkuStockLogRepo product_sku_stock_log.Repository
	productImageRepo       product_image.Repository
	productSpecRepo        product_spec.Repository
	productSpecValueRepo   product_spec_value.Repository
}

func New(db *gorm.DB, productRepo product.Repository, productSkuRepo product_sku.Repository, productSkuStockLogRepo product_sku_stock_log.Repository, productImageRepo product_image.Repository, productSpecRepo product_spec.Repository, productSpecValueRepo product_spec_value.Repository) *Service {
	return &Service{
		db:                     db,
		productRepo:            productRepo,
		productSkuRepo:         productSkuRepo,
		productSkuStockLogRepo: productSkuStockLogRepo,
		productImageRepo:       productImageRepo,
		productSpecRepo:        productSpecRepo,
		productSpecValueRepo:   productSpecValueRepo,
	}
}

// ListPage 用于获取商品列表信息
func (s *Service) ListPage(ctx context.Context, req ListPageReq) (ListPageResp, int, error) {
	// 获取列表数据
	offset, limit := req.OffsetLimit()
	products, total, err := s.productRepo.ListPage(ctx, nil, model.ProductListPageQuery{
		Name:       req.Name,
		CreditType: req.CreditType,
		Enable:     req.Enable,
		Offset:     offset,
		Limit:      limit,
	})
	if err != nil {
		return ListPageResp{}, 61001, err
	}
	// 返回数据
	return ListPageResp{
		Total:    total,
		PageData: toListPageItems(products),
	}, 0, nil
}

// Details 获取商品详情（含 SKU、规格、图片）
//
// includeDisabledProduct 为 true 时允许返回已禁用的商品（管理端场景）；
// false 时已禁用商品按不存在处理，适用于商城用户端。
// includeDisabledImages 为 true 时图片返回全部（含已停用）；
// false 时仅返回启用中的图片。
func (s *Service) Details(ctx context.Context, id int64, includeDisabledProduct, includeDisabledImages bool) (DetailsResp, int, error) {
	prod, err := s.productRepo.GetByID(ctx, nil, id)
	if err != nil {
		return DetailsResp{}, 61001, err
	}
	if prod == nil || (!includeDisabledProduct && prod.Enable != enum.EnableEnable) {
		return DetailsResp{}, 51001, errors.New("商品不存在")
	}
	// 查询关联数据
	skus, err := s.productSkuRepo.ListByProductID(ctx, nil, id)
	if err != nil {
		return DetailsResp{}, 61001, err
	}
	specs, err := s.productSpecRepo.ListByProductID(ctx, nil, id)
	if err != nil {
		return DetailsResp{}, 61001, err
	}
	specValues, err := s.productSpecValueRepo.ListByProductID(ctx, nil, id)
	if err != nil {
		return DetailsResp{}, 61001, err
	}
	var images []model.ProductImage
	if includeDisabledImages {
		images, err = s.productImageRepo.ListByProductID(ctx, nil, id)
	} else {
		images, err = s.productImageRepo.ListEnabledByProductID(ctx, nil, id)
	}
	if err != nil {
		return DetailsResp{}, 61001, err
	}
	// 返回数据
	return toDetailsResp(prod, skus, specs, specValues, images), 0, nil
}
