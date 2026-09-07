package order

import (
	"context"
	"errors"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_address"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_credit_log"
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
	liveUserAddress        live_user_address.Repository
	productRepo            product.Repository
	productSkuRepo         product_sku.Repository
	productSkuStockLogRepo product_sku_stock_log.Repository
	liveUserRepo           live_user.Repository
	liveUserCreditLogRepo  live_user_credit_log.Repository
}

func New(db *gorm.DB, liveUserOrderRepo live_user_order.Repository, liveUserOrderDraftRepo live_user_order_draft.Repository, liveUserAddress live_user_address.Repository, productRepo product.Repository, productSkuRepo product_sku.Repository, productSkuStockLogRepo product_sku_stock_log.Repository, liveUserRepo live_user.Repository, liveUserCreditLogRepo live_user_credit_log.Repository) *Service {
	return &Service{
		db:                     db,
		liveUserOrderRepo:      liveUserOrderRepo,
		liveUserOrderDraftRepo: liveUserOrderDraftRepo,
		liveUserAddress:        liveUserAddress,
		productRepo:            productRepo,
		productSkuRepo:         productSkuRepo,
		productSkuStockLogRepo: productSkuStockLogRepo,
		liveUserRepo:           liveUserRepo,
		liveUserCreditLogRepo:  liveUserCreditLogRepo,
	}
}

// PlaceOrder 用户下单方法：锁定库存并创建待支付草稿
func (s *Service) PlaceOrder(ctx context.Context, userID int64, req PlaceOrderReq) (int64, int, error) {
	return s.placeOrder(ctx, userID, req.SkuID, req.Count)
}

// ReOrder 重新下单：根据历史草稿 ID 找回商品 SKU 与数量，替用户重新锁定库存并创建新的待支付草稿
func (s *Service) ReOrder(ctx context.Context, userID, draftID int64) (int64, int, error) {
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
	draft, err := s.liveUserOrderDraftRepo.GetActiveByUserID(ctx, nil, userID)
	if err != nil {
		return UserOrderDraftResp{}, 61101, err
	}
	// 让确认页能展示上次超时/取消的订单并引导重新购买
	if draft == nil {
		draft, err = s.liveUserOrderDraftRepo.GetLatestCancelledByUserID(ctx, nil, userID)
		if err != nil {
			return UserOrderDraftResp{}, 61101, err
		}
	}
	if draft == nil {
		return UserOrderDraftResp{}, 51102, errors.New("无待支付订单")
	}
	// 获取商品信息
	product, err := s.productRepo.GetByID(ctx, nil, draft.ProductID)
	if err != nil {
		return UserOrderDraftResp{}, 61101, err
	}
	if product == nil {
		return UserOrderDraftResp{}, 51101, errors.New("商品不存在")
	}
	// 获取SKU信息
	sku, err := s.productSkuRepo.GetByID(ctx, nil, draft.ProductSkuID)
	if err != nil {
		return UserOrderDraftResp{}, 61101, err
	}
	if sku == nil {
		return UserOrderDraftResp{}, 51101, errors.New("商品SKU不存在")
	}
	return UserOrderDraftResp{
		ID:       draft.ID,
		ExpireAt: draft.ExpireAt,
		Product: ProductItem{
			ID:          product.ID,
			Name:        product.Name,
			Cover:       product.Cover,
			Price:       sku.Price,
			CreditType:  int(product.CreditType),
			ProductType: int(product.ProductType),
			Sku:         sku.SpecProperties,
			Count:       draft.Quantity,
		},
	}, 0, nil
}

// ConfirmPayment 确认兑换并完成支付
func (s *Service) ConfirmPayment(ctx context.Context, userID, draftID, addressID int64) (int64, int, error) {
	var orderID int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 读取草稿并校验归属
		draft, err := s.liveUserOrderDraftRepo.GetByID(ctx, tx, draftID)
		if err != nil {
			return err
		}
		if draft == nil || draft.UserID != userID {
			return errDraftNotFound
		}
		if draft.Status != enum.DraftStatusActive {
			return errDraftStateChanged
		}
		// 读取收货地址并校验归属
		address, err := s.liveUserAddress.GetByID(ctx, tx, addressID)
		if err != nil {
			return err
		}
		if address == nil || address.UserID != userID {
			return errAddressNotFound
		}
		// 读取商品与 SKU
		product, err := s.productRepo.GetByID(ctx, tx, draft.ProductID)
		if err != nil {
			return err
		}
		if product == nil {
			return errProductNotFound
		}
		sku, err := s.productSkuRepo.GetByID(ctx, tx, draft.ProductSkuID)
		if err != nil {
			return err
		}
		if sku == nil {
			return errSkuNotFound
		}
		// 原子占用草稿：Active→Redeemed。与到期取消互斥，占用失败说明已被取消或已兑换
		ok, err := s.liveUserOrderDraftRepo.RedeemActiveByID(ctx, tx, draftID)
		if err != nil {
			return err
		}
		if !ok {
			return errDraftStateChanged
		}
		// 生成订单号并创建订单
		orderSn := generateOrderSn()
		created, err := s.liveUserOrderRepo.Create(ctx, tx, &model.LiveUserOrder{
			UserID:                userID,
			OrderSn:               orderSn,
			ProductID:             draft.ProductID,
			ProductSkuID:          draft.ProductSkuID,
			ProductName:           product.Name,
			ProductCover:          product.Cover,
			ProductSpecProperties: sku.SpecProperties,
			Quantity:              draft.Quantity,
			CreditType:            product.CreditType,
			Price:                 sku.Price,
			ReceiverName:          address.Name,
			ReceiverPhone:         address.Phone,
			ReceiverRegionCode:    address.RegionCode,
			ReceiverRegion:        address.Region,
			ReceiverDetail:        address.Detail,
			ReceiverEmail:         address.Email,
			ReceiverType:          address.Type,
			OrderStatus:           enum.OrderStatusPendingShipment,
			PayStatus:             enum.PayStatusPaid,
			ShipStatus:            enum.ShipStatusPending,
			PayAt:                 time.Now().Unix(),
		})
		if err != nil {
			return err
		}
		orderID = created.ID
		// 回填下单锁定库存时的扣减流水，关联到订单
		if err := s.productSkuStockLogRepo.UpdateByDraftID(ctx, tx, draftID, created.ID, orderSn); err != nil {
			return err
		}
		// 扣减用户余额并写资产流水
		if err := s.deductBalance(ctx, tx, userID, product.CreditType, sku.Price*draft.Quantity, product.Name, draft.Quantity); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, errDraftNotFound):
			return 0, 51103, err
		case errors.Is(err, errDraftStateChanged):
			return 0, 51104, err
		case errors.Is(err, errAddressNotFound):
			return 0, 51301, err
		case errors.Is(err, errProductNotFound), errors.Is(err, errSkuNotFound):
			return 0, 51101, err
		case errors.Is(err, live_user.ErrInsufficientBalance):
			return 0, 40803, err
		default:
			return 0, 61101, err
		}
	}
	return orderID, 0, nil
}
