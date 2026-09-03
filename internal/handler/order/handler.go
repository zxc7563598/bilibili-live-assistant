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
	"go.uber.org/zap"
)

type Handler struct {
	orderSvc *order.Service
}

func New(orderSvc *order.Service) *Handler {
	return &Handler{
		orderSvc: orderSvc,
	}
}

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
		response.Error(c, lang, 20001)
		return
	}
	// 获取参数
	var req input.OrderPlaceOrderReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.ProductLogger, "ShopDetail 参数异常", code, err)
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
// @Success 200 {object} response.Response{data=resp.OrderGetConfirmResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/order/confirm [post]
func (h *Handler) GetConfirm(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.orderSvc.UserOrderDraft(ctx, userInfo.UserID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.OrderLogger,
			"orderSvc.UserOrderDraft 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.OrderGetConfirmResp{
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
