package order

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/region"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
	"gorm.io/gorm"
)

// draftExpireSeconds 草稿默认有效期（秒）：下单后在此时间内未完成兑换则自动取消并归还库存
const draftExpireSeconds = 600

// orderSnCounter 进程内自增序号，与纳秒时间戳组合保证单进程并发下单号不重复
var orderSnCounter atomic.Uint64

var (
	errSkuNotFound       = errors.New("商品 SKU 不存在")
	errInsufficientStock = errors.New("库存不足")
	errProductNotFound   = errors.New("商品不存在")
	errDraftNotFound     = errors.New("待兑换的订单不存在或不属于当前用户")
	errDraftStateChanged = errors.New("订单已取消或已完成兑换")
	errAddressNotFound   = errors.New("收货地址不存在或不属于当前用户")
)

// placeOrder 下单公共流程：取消用户已有 Active 草稿并归还库存 → 校验 SKU → 新建待支付草稿并锁定库存。
// skuID / count 由调用方给出（首次下单来自入参，重新下单从历史草稿解析）。返回新草稿 ID。
func (s *Service) placeOrder(ctx context.Context, userID, skuID, count int64) (int64, int, error) {
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
		sku, err := s.productSkuRepo.GetByID(ctx, tx, skuID)
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
			Quantity:     count,
			Status:       enum.DraftStatusActive,
			ExpireAt:     time.Now().Unix() + draftExpireSeconds,
		}
		created, err := s.liveUserOrderDraftRepo.Create(ctx, tx, draft)
		if err != nil {
			return err
		}
		draftID = created.ID
		// 扣减库存，防超卖（流水关联草稿与用户）
		if err := s.deductStock(ctx, tx, sku, count, draftID, userID); err != nil {
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
	if err := s.productSkuRepo.AdjustStock(ctx, tx, skuID, quantity); err != nil {
		return err
	}
	return s.productRepo.AdjustStock(ctx, tx, productID, quantity)
}

// deductBalance 扣减用户余额并写资产流水，需在调用方事务内执行
//
// 复用 liveuser 的 AdjustCredit：原先这里与 liveuser.addCreditLog 是同一套
// 「原子调余额 + 写审计流水」的两份实现，现已收敛为一处。
// 余额不足由数据库条件拦下，ErrInsufficientBalance 经 %w 透传，调用方仍可 errors.Is 判别。
func (s *Service) deductBalance(ctx context.Context, tx *gorm.DB, userID int64, creditType enum.CreditType, amount int64, productName string, quantity int64) error {
	if amount <= 0 {
		return nil // 免费商品不扣减
	}
	return s.liveUserSvc.AdjustCredit(ctx, tx, creditType, liveuser.AdjustCreditParams{
		UserID:       userID,
		ChangeType:   enum.ChangeTypeReduce,
		ChangeAmount: amount,
		BizType:      "order",
		Remark:       fmt.Sprintf("兑换商品 %s x %d", productName, quantity),
		OperatorType: enum.OperatorTypeUser,
		OperatorID:   userID,
	})
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

// generateOrderSn 生成订单号：SO + 纳秒时间戳(base36) + 自增序号(base36) + 随机后缀(base36)
func generateOrderSn() string {
	seq := orderSnCounter.Add(1)
	return "SO" +
		strconv.FormatInt(time.Now().UnixNano(), 36) +
		strconv.FormatUint(seq, 36) +
		randomSuffix()
}

// randomSuffix 返回 base36 编码的 24bit 随机串；crypto/rand 异常时回退时间戳低位
func randomSuffix() string {
	var b [3]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano()&0xFFFFFF, 36)
	}
	n := uint64(b[0])<<16 | uint64(b[1])<<8 | uint64(b[2])
	return strconv.FormatUint(n, 36)
}

func toListPageItems(list []model.LiveUserOrderListItem) []ListPageItem {
	respList := make([]ListPageItem, 0, len(list))
	for _, v := range list {
		item := ListPageItem{
			ID:                    v.ID,
			UserID:                v.UserID,
			UID:                   v.UID,
			Uname:                 v.Uname,
			OrderSn:               v.OrderSn,
			ProductID:             v.ProductID,
			ProductName:           v.ProductName,
			ProductCover:          v.ProductCover,
			ProductSpecProperties: v.ProductSpecProperties,
			Quantity:              v.Quantity,
			CreditType:            v.CreditType,
			Price:                 v.Price,
			ReceiverType:          v.ReceiverType,
			ReceiverName:          v.ReceiverName,
			ReceiverPhone:         v.ReceiverPhone,
			ReceiverRegion:        v.ReceiverRegion,
			ReceiverDetail:        v.ReceiverDetail,
			ReceiverEmail:         v.ReceiverEmail,
			OrderStatus:           v.OrderStatus,
			PayStatus:             v.PayStatus,
			ShipStatus:            v.ShipStatus,
			ExpressCompany:        v.ExpressCompany,
			ExpressNo:             v.ExpressNo,
			PayAt:                 timeutil.Format(v.PayAt),
			ProcessedAt:           timeutil.Format(v.ProcessedAt),
			CancelAt:              timeutil.Format(v.CancelAt),
			CreatedAt:             timeutil.Format(v.CreatedAt),
			Remark:                v.Remark,
		}
		respList = append(respList, item)
	}
	return respList
}

// toDetailsItem 订单联查结果 → 详情出参
func toDetailsItem(v model.LiveUserOrderListItem) DetailsItem {
	return DetailsItem{
		ID:                    v.ID,
		UserID:                v.UserID,
		UID:                   v.UID,
		Uname:                 v.Uname,
		OrderSn:               v.OrderSn,
		ProductID:             v.ProductID,
		ProductSkuID:          v.ProductSkuID,
		ProductName:           v.ProductName,
		ProductCover:          v.ProductCover,
		ProductSpecProperties: v.ProductSpecProperties,
		Quantity:              v.Quantity,
		CreditType:            v.CreditType,
		Price:                 v.Price,
		ReceiverName:          v.ReceiverName,
		ReceiverPhone:         v.ReceiverPhone,
		ReceiverRegionCode:    v.ReceiverRegionCode,
		ReceiverRegion:        v.ReceiverRegion,
		ReceiverDetail:        v.ReceiverDetail,
		ReceiverEmail:         v.ReceiverEmail,
		ReceiverType:          v.ReceiverType,
		OrderStatus:           v.OrderStatus,
		PayStatus:             v.PayStatus,
		ShipStatus:            v.ShipStatus,
		ExpressCompany:        v.ExpressCompany,
		ExpressNo:             v.ExpressNo,
		PayAt:                 timeutil.Format(v.PayAt),
		ProcessedAt:           timeutil.Format(v.ProcessedAt),
		CancelAt:              timeutil.Format(v.CancelAt),
		CreatedAt:             timeutil.Format(v.CreatedAt),
		UpdatedAt:             timeutil.Format(v.UpdatedAt),
		Remark:                v.Remark,
	}
}

// resolveRegionCode 归一化订单收货地区：入参为 JSON 数组字符串（如 ["370000","370100","370116"]），
// 与 live_user_addresses.region_code 存储格式一致 —— 下单时原样拷贝、商城端也是 JSON.stringify 后提交。
// 返回紧凑 JSON 与从 regions.json 派生的文案；"" 或 "[]" 视为清空（返回空串，由调用方按必填规则判断）。
// 与 address 模块的 applyRegion 同源，但不复用其实现：那是 address Service 的方法、参数是 address 专用结构，
// 复用会把 order → address 的 Service 依赖引进来。共享边界是 internal/region 包。
func resolveRegionCode(src string) (regionCode, regionText string, errCode int) {
	src = strings.TrimSpace(src)
	if src == "" {
		return "", "", 0
	}
	var codes []string
	if err := json.Unmarshal([]byte(src), &codes); err != nil {
		return "", "", 11302
	}
	if len(codes) == 0 {
		return "", "", 0
	}
	text, ok := region.Resolve(codes)
	if !ok {
		return "", "", 11302
	}
	canonical, err := json.Marshal(codes)
	if err != nil {
		return "", "", 11302
	}
	return string(canonical), text, 0
}
