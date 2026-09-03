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

// PlaceOrder 用户下单方法：锁定库存并创建待支付草稿。
// 返回值：新草稿 ID、错误码（0 成功）、原始错误。
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
