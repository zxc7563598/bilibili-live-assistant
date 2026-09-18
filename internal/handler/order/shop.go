package order

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/order"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"
	"go.uber.org/zap"
)

// @Summary 用户下单
// @Description 用户在商城选择下单，锁定库存并允许用户在指定时间支付/下单
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.OrderPlaceOrderReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/shop/order/place [post]
func (h *Handler) PlaceOrder(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.OrderPlaceOrderReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.OrderLogger, "PlaceOrder 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	_, errCode, err := h.orderSvc.PlaceOrder(ctx, userInfo.UserID, order.PlaceOrderReq{
		SkuID: req.SkuID,
		Count: req.Count,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.OrderLogger,
			"orderSvc.PlaceOrder 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
			zap.Any("req.sku_id", req.SkuID),
			zap.Any("req.count", req.Count),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 获取用户下单数据
// @Description 获取用户已经下单尚未支付的数据
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.OrderGetDraftResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/order/confirm [post]
func (h *Handler) GetOrderDraft(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.orderSvc.GetUserOrderDraft(ctx, userInfo.UserID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.OrderLogger,
			"orderSvc.GetUserOrderDraft 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.OrderGetDraftResp{
		ID:       svcResp.ID,
		ExpireAt: svcResp.ExpireAt * 1000,
		Product: resp.ProductItem{
			ID:          svcResp.Product.ID,
			Name:        svcResp.Product.Name,
			Cover:       svcResp.Product.Cover,
			Price:       svcResp.Product.Price,
			CreditType:  svcResp.Product.CreditType,
			ProductType: svcResp.Product.ProductType,
			Sku:         svcResp.Product.Sku,
			Count:       svcResp.Product.Count,
		},
	})
}

// @Summary 用户重新下单
// @Description 根据历史草稿ID找回商品与数量，重新锁定库存创建新的待支付草稿（用于已超时/已取消订单重新购买）
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.OrderReOrderReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/shop/order/again [post]
func (h *Handler) ReOrder(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.OrderReOrderReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.OrderLogger, "ReOrder 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	_, errCode, err := h.orderSvc.ReOrder(ctx, userInfo.UserID, req.ID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.OrderLogger,
			"orderSvc.ReOrder 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
			zap.Any("req.id", req.ID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 用户确认兑换并完成支付
// @Description 用户实际进行下单/支付
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.OrderConfirmPaymentReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/shop/order/payment [post]
func (h *Handler) ConfirmPayment(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.OrderConfirmPaymentReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.OrderLogger, "ConfirmPayment 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	_, errCode, err := h.orderSvc.ConfirmPayment(ctx, userInfo.UserID, req.DraftID, req.AddressID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.OrderLogger,
			"orderSvc.ConfirmPayment 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
			zap.Any("req.draft_id", req.DraftID),
			zap.Any("req.address_id", req.AddressID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 分页查询我的订单
// @Description 按状态分页查询当前用户的历史订单列表
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.OrderListPageByUserReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.OrderListPageByUserResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/order/list [post]
func (h *Handler) ListPageByUser(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.OrderListPageByUserReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.OrderLogger, "ListPageByUser 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.orderSvc.ListPageByUser(ctx, userInfo.UserID, order.ListPageByUserReq{
		PageResp: pagination.PageResp{
			PageNo:    req.PageNo,
			PageSize:  req.PageSize,
			SortField: req.SortField,
			SortOrder: req.SortOrder,
		},
		OrderStatus: req.OrderStatus,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.OrderLogger,
			"orderSvc.ListPageByUser 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
			zap.Int("req.pageNo", req.PageNo),
			zap.Int("req.pageSize", req.PageSize),
			zap.Any("req.order_status", req.OrderStatus),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.OrderListPageByUserResp{
		Total:    svcResp.Total,
		PageData: toOrderListPageByUserItem(svcResp.PageData),
	})
}
