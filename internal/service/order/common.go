package order

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"gorm.io/gorm"
)

// draftExpireSeconds 草稿默认有效期（秒）：下单后在此时间内未完成兑换则自动取消并归还库存
const draftExpireSeconds = 600

var (
	errSkuNotFound       = errors.New("商品 SKU 不存在")
	errInsufficientStock = errors.New("库存不足")
)

// deductStock 扣减 SKU 与商品库存并写入扣减流水（下单场景）
func (s *Service) deductStock(ctx context.Context, tx *gorm.DB, sku *model.ProductSku, count, draftID, userID int64) error {
	if sku.Stock < count {
		return errInsufficientStock
	}
	// 扣减流水（Before/After 取 SKU 库存；OrderID/OrderSn 留空，此时尚无订单）
	if _, err := s.productSkuStockLogRepo.Create(ctx, tx, &model.ProductSkuStockLog{
		ProductID:    sku.ProductID,
		ProductSkuID: sku.ID,
		ChangeNum:    -count,
		BeforeStock:  sku.Stock,
		AfterStock:   sku.Stock - count,
		DraftID:      draftID,
		UserID:       userID,
		Type:         enum.StockChangeTypeOrderDeduct,
	}); err != nil {
		return err
	}
	// SKU 库存条件扣减
	ok, err := s.productSkuRepo.DecrementStock(ctx, tx, sku.ID, count)
	if err != nil {
		return err
	}
	if !ok {
		return errInsufficientStock
	}
	// 商品库存同步扣减
	ok, err = s.productRepo.DecrementStock(ctx, tx, sku.ProductID, count)
	if err != nil {
		return err
	}
	if !ok {
		return errInsufficientStock
	}
	return nil
}

// returnStock 归还 SKU 与商品库存并写入归还流水（取消/过期场景）
func (s *Service) returnStock(ctx context.Context, tx *gorm.DB, productID, skuID, quantity, draftID, userID int64) error {
	sku, err := s.productSkuRepo.GetByID(ctx, tx, skuID)
	if err != nil {
		return err
	}
	if sku == nil {
		return errSkuNotFound
	}
	if _, err := s.productSkuStockLogRepo.Create(ctx, tx, &model.ProductSkuStockLog{
		ProductID:    productID,
		ProductSkuID: skuID,
		ChangeNum:    quantity,
		BeforeStock:  sku.Stock,
		AfterStock:   sku.Stock + quantity,
		DraftID:      draftID,
		UserID:       userID,
		Type:         enum.StockChangeTypeReturnRefund,
	}); err != nil {
		return err
	}
	if err := s.productSkuRepo.IncrementStock(ctx, tx, skuID, quantity); err != nil {
		return err
	}
	return s.productRepo.IncrementStock(ctx, tx, productID, quantity)
}

// cancelDraft 取消草稿并归还库存（status=Active → Cancelled，幂等）
func (s *Service) cancelDraft(ctx context.Context, tx *gorm.DB, draft *model.LiveUserOrderDraft) error {
	ok, err := s.liveUserOrderDraftRepo.CancelActiveByID(ctx, tx, draft.ID)
	if err != nil {
		return err
	}
	if !ok {
		return nil // 已非 Active，无需处理
	}
	return s.returnStock(ctx, tx, draft.ProductID, draft.ProductSkuID, draft.Quantity, draft.ID, draft.UserID)
}

// expireDraftByID 按 ID 加载草稿并取消归还（AfterFunc 主路径，幂等）
func (s *Service) expireDraftByID(ctx context.Context, draftID int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		draft, err := s.liveUserOrderDraftRepo.GetByID(ctx, tx, draftID)
		if err != nil {
			return err
		}
		if draft == nil {
			return nil
		}
		return s.cancelDraft(ctx, tx, draft)
	})
}

// scheduleDraftExpiry 下单提交后注册 600s 定时器，精准取消归还（进程内主路径）
func (s *Service) scheduleDraftExpiry(draftID int64) {
	time.AfterFunc(draftExpireSeconds*time.Second, func() {
		if err := s.expireDraftByID(context.Background(), draftID); err != nil {
			log.Printf("[order.Expire] 草稿 %d 过期取消失败: %v", draftID, err)
		}
	})
}

// ExpireDrafts 定时任务（兜底）：扫描并取消所有已过期的 Active 草稿、归还库存。
// 用于进程重启后 AfterFunc 丢失的场景；每条草稿独立事务，单条失败仅记日志不中断。
func (s *Service) ExpireDrafts(ctx context.Context) error {
	list, err := s.liveUserOrderDraftRepo.ListActiveExpired(ctx, nil, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("查询过期草稿失败: %w", err)
	}
	for i := range list {
		draft := list[i]
		if err := s.db.Transaction(func(tx *gorm.DB) error {
			return s.cancelDraft(ctx, tx, &draft)
		}); err != nil {
			log.Printf("[order.Expire] 草稿 %d 过期取消失败: %v", draft.ID, err)
		}
	}
	return nil
}
