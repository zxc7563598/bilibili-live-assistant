package order

import (
	"context"
	"errors"

	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_order"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_order_draft"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product_sku"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/product_sku_stock_log"
	"gorm.io/gorm"
)

type Service struct {
	db                     *gorm.DB
	liveUserOrderRepo      live_user_order.Repository
	liveUserOrderDraftRepo live_user_order_draft.Repository
	productRepo            product.Repository
	productSkuRepo         product_sku.Repository
	productSkuStockLogRepo product_sku_stock_log.Repository
}

func New(db *gorm.DB, liveUserOrderRepo live_user_order.Repository, liveUserOrderDraftRepo live_user_order_draft.Repository, productRepo product.Repository, productSkuRepo product_sku.Repository, productSkuStockLogRepo product_sku_stock_log.Repository) *Service {
	return &Service{
		db:                     db,
		liveUserOrderRepo:      liveUserOrderRepo,
		liveUserOrderDraftRepo: liveUserOrderDraftRepo,
		productRepo:            productRepo,
		productSkuRepo:         productSkuRepo,
		productSkuStockLogRepo: productSkuStockLogRepo,
	}
}

// PlaceOrder 用户下单方法：锁定库存并创建待支付草稿
func (s *Service) PlaceOrder(ctx context.Context, userID int64, req PlaceOrderReq) (int64, int, error) {
	if req.SkuID <= 0 || req.Count <= 0 {
		return 0, 11101, errors.New("下单参数不合法")
	}
	return s.placeOrder(ctx, userID, req.SkuID, req.Count)
}

// ReOrder 重新下单：根据历史草稿 ID 找回商品 SKU 与数量，替用户重新锁定库存并创建新的待支付草稿
func (s *Service) ReOrder(ctx context.Context, userID, draftID int64) (int64, int, error) {
	if draftID <= 0 {
		return 0, 11101, errors.New("重新下单参数不合法")
	}
	// 读取历史草稿并校验归属
	prev, err := s.liveUserOrderDraftRepo.GetByID(ctx, nil, draftID)
	if err != nil {
		return 0, 61101, err
	}
	if prev == nil || prev.UserID != userID {
		return 0, 51103, errors.New("待重新购买的订单不存在或不属于当前用户")
	}
	// 按该草稿的 SKU 与数量重新下单
	return s.placeOrder(ctx, userID, prev.ProductSkuID, prev.Quantity)
}

// UserOrderDraft 获取用户下单数据
func (s *Service) UserOrderDraft(ctx context.Context, userID int64) (UserOrderDraftResp, int, error) {
	// 获取用户当前 Active 状态的草稿
	active, err := s.liveUserOrderDraftRepo.GetActiveByUserID(ctx, nil, userID)
	if err != nil {
		return UserOrderDraftResp{}, 61101, err
	}
	if active == nil {
		return UserOrderDraftResp{}, 51102, errors.New("无待支付订单")
	}
	// 获取商品信息
	product, err := s.productRepo.GetByID(ctx, nil, active.ProductID)
	if err != nil {
		return UserOrderDraftResp{}, 61101, err
	}
	if product == nil {
		return UserOrderDraftResp{}, 51101, errors.New("商品不存在")
	}
	// 获取SKU信息
	sku, err := s.productSkuRepo.GetByID(ctx, nil, active.ProductSkuID)
	if err != nil {
		return UserOrderDraftResp{}, 61101, err
	}
	if sku == nil {
		return UserOrderDraftResp{}, 51101, errors.New("商品SKU不存在")
	}
	return UserOrderDraftResp{
		ID:       active.ID,
		ExpireAt: active.ExpireAt,
		Product: ProductItem{
			ID:          product.ID,
			Name:        product.Name,
			Cover:       product.Cover,
			Price:       sku.Price,
			CreditType:  int(product.CreditType),
			ProductType: int(product.ProductType),
			Sku:         sku.SpecProperties,
			Count:       active.Quantity,
		},
	}, 0, nil
}
