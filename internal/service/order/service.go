package order

import (
	"context"
	"errors"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
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
	var draftID int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 取消已有 Active 草稿并归还库存（重复下单 = 放弃旧单）
		active, err := s.liveUserOrderDraftRepo.GetActiveByUserID(ctx, tx, userID)
		if err != nil {
			return err
		}
		if active != nil {
			if err := s.cancelDraft(ctx, tx, active); err != nil {
				return err
			}
		}
		// 获取 SKU
		sku, err := s.productSkuRepo.GetByID(ctx, tx, req.SkuID)
		if err != nil {
			return err
		}
		if sku == nil {
			return errSkuNotFound
		}
		// 创建草稿锁定库存
		draft := &model.LiveUserOrderDraft{
			UserID:       userID,
			ProductID:    sku.ProductID,
			ProductSkuID: sku.ID,
			Quantity:     req.Count,
			Status:       enum.DraftStatusActive,
			ExpireAt:     time.Now().Unix() + draftExpireSeconds,
		}
		created, err := s.liveUserOrderDraftRepo.Create(ctx, tx, draft)
		if err != nil {
			return err
		}
		draftID = created.ID
		// 扣减库存，防超卖（流水关联草稿与用户）
		if err := s.deductStock(ctx, tx, sku, req.Count, draftID, userID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, errInsufficientStock):
			return 0, 41101, err
		case errors.Is(err, errSkuNotFound):
			return 0, 51101, err
		default:
			return 0, 61101, err
		}
	}
	// 事务提交成功后注册定时器
	s.scheduleDraftExpiry(draftID)
	return draftID, 0, nil
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
