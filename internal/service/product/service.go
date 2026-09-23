package product

import (
	"context"
	"errors"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product_image"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product_sku"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product_spec"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product_spec_value"
	"gorm.io/gorm"
)

type Service struct {
	db                   *gorm.DB
	productRepo          product.Repository
	productSkuRepo       product_sku.Repository
	productImageRepo     product_image.Repository
	productSpecRepo      product_spec.Repository
	productSpecValueRepo product_spec_value.Repository
}

func New(db *gorm.DB, productRepo product.Repository, productSkuRepo product_sku.Repository, productImageRepo product_image.Repository, productSpecRepo product_spec.Repository, productSpecValueRepo product_spec_value.Repository) *Service {
	return &Service{
		db:                   db,
		productRepo:          productRepo,
		productSkuRepo:       productSkuRepo,
		productImageRepo:     productImageRepo,
		productSpecRepo:      productSpecRepo,
		productSpecValueRepo: productSpecValueRepo,
	}
}

// ListPage 用于获取商品列表信息
func (s *Service) ListPage(ctx context.Context, req ListPageReq) (ListPageResp, int, error) {
	// 获取列表数据
	offset, limit, sortField, sortOrder := req.OffsetLimit()
	products, total, err := s.productRepo.ListPage(ctx, nil, model.ProductListPageQuery{
		Name:       req.Name,
		CreditType: req.CreditType,
		Enable:     req.Enable,
		Offset:     offset,
		Limit:      limit,
		SortField:  sortField,
		SortOrder:  sortOrder,
	})
	if err != nil {
		return ListPageResp{}, CodeQueryFailed, err
	}
	// 返回数据
	return ListPageResp{
		Total:    total,
		PageData: toListPageItems(products),
	}, 0, nil
}

// Details 获取商品详情（含 SKU、规格、图片）
func (s *Service) Details(ctx context.Context, id int64, includeDisabledProduct, includeDisabledImages bool) (DetailsResp, int, error) {
	prod, err := s.productRepo.GetByID(ctx, nil, id)
	if err != nil {
		return DetailsResp{}, CodeQueryFailed, err
	}
	if prod == nil || (!includeDisabledProduct && prod.Enable != enum.EnableEnable) {
		return DetailsResp{}, CodeNotFound, errors.New("商品不存在")
	}
	// 查询关联数据
	skus, err := s.productSkuRepo.ListByProductID(ctx, nil, id)
	if err != nil {
		return DetailsResp{}, CodeQueryFailed, err
	}
	specs, err := s.productSpecRepo.ListByProductID(ctx, nil, id)
	if err != nil {
		return DetailsResp{}, CodeQueryFailed, err
	}
	specValues, err := s.productSpecValueRepo.ListByProductID(ctx, nil, id)
	if err != nil {
		return DetailsResp{}, CodeQueryFailed, err
	}
	var images []model.ProductImage
	if includeDisabledImages {
		images, err = s.productImageRepo.ListByProductID(ctx, nil, id)
	} else {
		images, err = s.productImageRepo.ListEnabledByProductID(ctx, nil, id)
	}
	if err != nil {
		return DetailsResp{}, CodeQueryFailed, err
	}
	// 返回数据
	return toDetailsResp(prod, skus, specs, specValues, images), 0, nil
}

// UpdateEnable 快速变更商品是否启用
func (s *Service) UpdateEnable(ctx context.Context, id int64, enable bool) (int, error) {
	prod, err := s.productRepo.GetByID(ctx, nil, id)
	if err != nil {
		return CodeQueryFailed, err
	}
	if prod == nil {
		return CodeNotFound, errors.New("商品不存在")
	}
	if err := s.productRepo.UpdateField(ctx, nil, id, "enable", enum.BoolToEnablePtr(&enable)); err != nil {
		return CodeSaveFailed, err
	}
	return 0, nil
}

// UpdateMinMemberLevel 变更商品限购档位
func (s *Service) UpdateMinMemberLevel(ctx context.Context, id int64, level int) (int, error) {
	lv := enum.MinMemberLevel(level)
	if !lv.IsValid() {
		return CodeMinMemberLevelInvalid, nil
	}
	// 预检商品存在：UpdateField 不返回影响行数，没有这一步「商品不存在」会被当成成功
	prod, err := s.productRepo.GetByID(ctx, nil, id)
	if err != nil {
		return CodeQueryFailed, err
	}
	if prod == nil {
		return CodeNotFound, errors.New("商品不存在")
	}
	if err := s.productRepo.UpdateField(ctx, nil, id, "min_member_level", lv); err != nil {
		return CodeSaveFailed, err
	}
	return 0, nil
}

// Save 创建或变更商品
func (s *Service) Save(ctx context.Context, req SaveReq) (SaveResp, int, error) {
	if errCode, err := validateSaveReq(req); errCode != 0 {
		return SaveResp{}, errCode, err
	}
	var productID int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var err error
		if req.ID == 0 {
			productID, err = s.add(ctx, tx, req)
		} else {
			productID, err = s.update(ctx, tx, req)
		}
		if err != nil {
			return err
		}
		if err := s.syncRelations(ctx, tx, productID, req); err != nil {
			return err
		}
		return s.refreshProductStock(ctx, tx, productID)
	})
	if err != nil {
		if errors.Is(err, errProductNotFound) {
			return SaveResp{}, CodeNotFound, err
		}
		return SaveResp{}, CodeSaveFailed, err
	}
	return SaveResp{ID: productID}, 0, nil
}

// refreshProductStock 按落库后的 SKU 汇总刷新商品库存
func (s *Service) refreshProductStock(ctx context.Context, tx *gorm.DB, productID int64) error {
	skus, err := s.productSkuRepo.ListByProductID(ctx, tx, productID)
	if err != nil {
		return err
	}
	var total int64
	for _, sku := range skus {
		total += sku.Stock
	}
	return s.productRepo.UpdateField(ctx, tx, productID, "stock", total)
}

// add 新增商品主表记录，返回新商品 ID
func (s *Service) add(ctx context.Context, tx *gorm.DB, req SaveReq) (int64, error) {
	prod := toProductEntity(req)
	created, err := s.productRepo.Create(ctx, tx, &prod)
	if err != nil {
		return 0, err
	}
	return created.ID, nil
}

// update 变更商品主表记录，返回商品 ID
func (s *Service) update(ctx context.Context, tx *gorm.DB, req SaveReq) (int64, error) {
	prod, err := s.productRepo.GetByID(ctx, tx, req.ID)
	if err != nil {
		return 0, err
	}
	if prod == nil {
		return 0, errProductNotFound
	}
	// 只覆盖请求里携带的字段，CreatedAt 等由实体自身保留；
	// Stock 由 refreshProductStock 在 SKU 落库后单独刷新
	updated := toProductEntity(req)
	prod.Name = updated.Name
	prod.Cover = updated.Cover
	prod.Price = updated.Price
	prod.CreditType = updated.CreditType
	prod.ProductType = updated.ProductType
	prod.Sold = updated.Sold
	prod.Tags = updated.Tags
	prod.Describe = updated.Describe
	prod.SortOrder = updated.SortOrder
	prod.Enable = updated.Enable
	prod.MinMemberLevel = updated.MinMemberLevel
	if err := s.productRepo.Save(ctx, tx, prod); err != nil {
		return 0, err
	}
	return prod.ID, nil
}
