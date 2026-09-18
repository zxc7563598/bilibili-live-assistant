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
		if err := s.productSkuStockLogRepo.UpdateOrderRefByDraftID(ctx, tx, draftID, created.ID, orderSn); err != nil {
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

// ListPage 后台分页查询全部订单，联查 live_users 返回 uid/uname
func (s *Service) ListPage(ctx context.Context, req ListPageReq) (ListPageResp, int, error) {
	// 获取列表数据
	offset, limit, sortField, sortOrder := req.OffsetLimit()
	list, total, err := s.liveUserOrderRepo.ListPage(ctx, nil, model.LiveUserOrderListPageQuery{
		UID:         req.UID,
		Uname:       req.Uname,
		OrderSn:     req.OrderSn,
		OrderStatus: req.OrderStatus,
		PayStatus:   req.PayStatus,
		ShipStatus:  req.ShipStatus,
		Offset:      offset,
		Limit:       limit,
		SortField:   sortField,
		SortOrder:   sortOrder,
	})
	if err != nil {
		return ListPageResp{}, 61101, err
	}
	// 返回数据
	return ListPageResp{
		Total:    total,
		PageData: toListPageItems(list),
	}, 0, nil
}

// Details 后台查询订单详情，返回订单全部字段与用户 uid/uname
func (s *Service) Details(ctx context.Context, id int64) (DetailsItem, int, error) {
	item, err := s.liveUserOrderRepo.GetDetailByID(ctx, nil, id)
	if err != nil {
		return DetailsItem{}, 61101, err
	}
	if item == nil {
		return DetailsItem{}, 51105, errors.New("订单不存在")
	}
	return toDetailsItem(*item), 0, nil
}

// UpdateShipStatus 后台变更发货状态，并按目标发货状态联动订单状态
func (s *Service) UpdateShipStatus(ctx context.Context, req UpdateShipStatusReq) (int, error) {
	// 读取订单
	order, err := s.liveUserOrderRepo.GetByID(ctx, nil, req.ID)
	if err != nil {
		return 61101, err
	}
	if order == nil {
		return 51105, errors.New("订单不存在")
	}
	if !req.ShipStatus.IsValid() {
		return 11101, errors.New("发货状态不合法")
	}
	// 虚拟商品没有物流环节，不接受快递信息
	if order.ReceiverType != enum.AddressTypeActual && (req.ExpressCompany != nil || req.ExpressNo != nil) {
		return 11101, errors.New("虚拟订单不支持快递信息")
	}
	// 快递信息仅在传了非空值时才覆盖，避免前端提交空串把已填单号冲掉
	if v := strPtr(req.ExpressCompany); v != "" {
		order.ExpressCompany = v
	}
	if v := strPtr(req.ExpressNo); v != "" {
		order.ExpressNo = v
	}
	// 发货状态没变就只落快递信息：此时订单状态与发货时间都不该被动
	if req.ShipStatus != order.ShipStatus {
		if req.ShipStatus == enum.ShipStatusPending {
			// 撤销发货：订单回到待发货，发货时间与快递信息一并清空
			order.OrderStatus = enum.OrderStatusPendingShipment
			order.ProcessedAt = 0
			order.ExpressCompany = ""
			order.ExpressNo = ""
		} else {
			// 已发货：虚拟商品直接完成，实体商品进入待收货；已送达一律完成
			if req.ShipStatus == enum.ShipStatusShipped && order.ReceiverType == enum.AddressTypeActual {
				order.OrderStatus = enum.OrderStatusPendingReceipt
			} else {
				order.OrderStatus = enum.OrderStatusCompleted
			}
			// 首次发货记发货时间，重复变更保留原时间
			if order.ProcessedAt == 0 {
				order.ProcessedAt = time.Now().Unix()
			}
		}
	}
	order.ShipStatus = req.ShipStatus
	// 落库
	if err := s.liveUserOrderRepo.Save(ctx, nil, order); err != nil {
		return 61101, err
	}
	return 0, nil
}

// UpdateOrderStatus 后台变更订单状态（确认收货 / 取消订单等），仅变更状态字段，不涉及退款与库存回滚
func (s *Service) UpdateOrderStatus(ctx context.Context, req UpdateOrderStatusReq) (int, error) {
	// 读取订单
	order, err := s.liveUserOrderRepo.GetByID(ctx, nil, req.ID)
	if err != nil {
		return 61101, err
	}
	if order == nil {
		return 51105, errors.New("订单不存在")
	}
	if !req.OrderStatus.IsValid() {
		return 11101, errors.New("订单状态不合法")
	}
	// 已是目标状态直接返回：避免重复取消把真实的取消时间改晚
	if order.OrderStatus == req.OrderStatus {
		return 0, nil
	}
	order.OrderStatus = req.OrderStatus
	if req.OrderStatus == enum.OrderStatusCancelled {
		order.CancelAt = time.Now().Unix()
	} else {
		order.CancelAt = 0
	}
	// 落库
	if err := s.liveUserOrderRepo.Save(ctx, nil, order); err != nil {
		return 61101, err
	}
	return 0, nil
}

// UpdateReceiverInfo 后台变更订单收货信息，仅改 live_user_orders 单条记录，不涉及 live_user_addresses
func (s *Service) UpdateReceiverInfo(ctx context.Context, req UpdateReceiverInfoReq) (int, error) {
	// 读取订单；存在性校验放在空请求判断之前，
	// 否则用一个不存在的 ID 调本接口会拿到成功，排查问题时容易被误导
	order, err := s.liveUserOrderRepo.GetByID(ctx, nil, req.ID)
	if err != nil {
		return 61101, err
	}
	if order == nil {
		return 51105, errors.New("订单不存在")
	}
	// 未传任何可变更字段，无需落库
	if req.ReceiverName == nil && req.ReceiverPhone == nil && req.ReceiverRegionCode == nil &&
		req.ReceiverDetail == nil && req.ReceiverEmail == nil {
		return 0, nil
	}
	switch order.ReceiverType {
	case enum.AddressTypeVirtual:
		// 虚拟订单只改邮箱，传了收货地址字段说明调用方串错了表单
		if req.ReceiverName != nil || req.ReceiverPhone != nil || req.ReceiverRegionCode != nil || req.ReceiverDetail != nil {
			return 11101, errors.New("虚拟订单只支持变更邮箱")
		}
		if req.ReceiverEmail != nil {
			order.ReceiverEmail = strPtr(req.ReceiverEmail)
		}
		if order.ReceiverEmail == "" {
			return 11308, nil
		}
	case enum.AddressTypeActual:
		// 实体订单只改收货地址
		if req.ReceiverEmail != nil {
			return 11101, errors.New("实体订单只支持变更收货地址")
		}
		if req.ReceiverName != nil {
			order.ReceiverName = strPtr(req.ReceiverName)
		}
		if req.ReceiverPhone != nil {
			order.ReceiverPhone = strPtr(req.ReceiverPhone)
		}
		if req.ReceiverDetail != nil {
			order.ReceiverDetail = strPtr(req.ReceiverDetail)
		}
		// 地区以后端从 region_code 派生的文案为准，不接受前端自由文本
		if req.ReceiverRegionCode != nil {
			regionCode, regionText, errCode := resolveRegionCode(*req.ReceiverRegionCode)
			if errCode != 0 {
				return errCode, errors.New("地区信息不正确")
			}
			order.ReceiverRegionCode = regionCode
			order.ReceiverRegion = regionText
		}
		// 合并后按实体地址必填项校验，顺序与 address.validateEntityFields 保持一致
		if order.ReceiverName == "" {
			return 11304, nil
		}
		if order.ReceiverPhone == "" {
			return 11305, nil
		}
		if order.ReceiverRegionCode == "" {
			return 11306, nil
		}
		if order.ReceiverDetail == "" {
			return 11307, nil
		}
	default:
		return 11303, errors.New("收货人地址类型不合法")
	}
	// 落库
	if err := s.liveUserOrderRepo.Save(ctx, nil, order); err != nil {
		return 61101, err
	}
	return 0, nil
}

// ListPageByUser 根据用户ID获取订单列表信息
func (s *Service) ListPageByUser(ctx context.Context, userID int64, req ListPageByUserReq) (ListPageByUserResp, int, error) {
	// 获取列表数据
	offset, limit, sortField, sortOrder := req.OffsetLimit()
	list, total, err := s.liveUserOrderRepo.ListPage(ctx, nil, model.LiveUserOrderListPageQuery{
		UserID:      &userID,
		OrderStatus: req.OrderStatus,
		Offset:      offset,
		Limit:       limit,
		SortField:   sortField,
		SortOrder:   sortOrder,
	})
	if err != nil {
		return ListPageByUserResp{}, 61101, err
	}
	// 返回数据
	return ListPageByUserResp{
		Total:    total,
		PageData: toListPageItems(list),
	}, 0, nil
}
