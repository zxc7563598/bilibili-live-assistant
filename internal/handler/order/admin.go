package order

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/order"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"
	"go.uber.org/zap"
)

// @Summary 后台分页查询订单列表
// @Description 分页查询全部订单，联查用户表返回 uid/uname，支持按 UID（精确）、昵称（模糊）、订单号（模糊）与各状态筛选
// @Tags 订单管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.OrderListPageReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.OrderListPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/order/list [post]
func (h *Handler) ListPage(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员信息
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.OrderListPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.OrderLogger, "ListPage 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.orderSvc.ListPage(ctx, order.ListPageReq{
		PageResp: pagination.PageResp{
			PageNo:    req.PageNo,
			PageSize:  req.PageSize,
			SortField: req.SortField,
			SortOrder: req.SortOrder,
		},
		UID:         req.UID,
		Uname:       req.Uname,
		OrderSn:     req.OrderSn,
		OrderStatus: req.OrderStatus,
		PayStatus:   req.PayStatus,
		ShipStatus:  req.ShipStatus,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.OrderLogger,
			"orderSvc.ListPage 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int("req.pageNo", req.PageNo),
			zap.Int("req.pageSize", req.PageSize),
			zap.Any("req.uid", req.UID),
			zap.Any("req.uname", req.Uname),
			zap.Any("req.order_sn", req.OrderSn),
			zap.Any("req.order_status", req.OrderStatus),
			zap.Any("req.pay_status", req.PayStatus),
			zap.Any("req.ship_status", req.ShipStatus),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.OrderListPageResp{
		Total:    svcResp.Total,
		PageData: toOrderAdminListPageItem(svcResp.PageData),
	})
}

// @Summary 后台获取订单详情
// @Description 根据订单ID返回订单全部字段，以及下单用户的 uid/uname
// @Tags 订单管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.OrderDetailsReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.OrderDetailsResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/order/details [post]
func (h *Handler) Details(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员信息
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.OrderDetailsReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.OrderLogger, "Details 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.orderSvc.Details(ctx, req.ID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.OrderLogger,
			"orderSvc.Details 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.id", req.ID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, toOrderDetailsResp(svcResp))
}

// @Summary 后台变更发货状态
// @Description 变更订单发货状态，并按目标状态联动订单状态：未发货→待发货；已发货→虚拟商品直接完成、实体商品进入待收货；已送达→完成。发货状态未变化时只更新快递信息，不动订单状态与发货时间。快递信息可选且仅实体订单可填
// @Tags 订单管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.OrderUpdateShipStatusReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/order/ship-status [post]
func (h *Handler) UpdateShipStatus(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员信息
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.OrderUpdateShipStatusReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.OrderLogger, "UpdateShipStatus 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.orderSvc.UpdateShipStatus(ctx, order.UpdateShipStatusReq{
		ID:             req.ID,
		ShipStatus:     enum.ShipStatus(*req.ShipStatus),
		ExpressCompany: req.ExpressCompany,
		ExpressNo:      req.ExpressNo,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.OrderLogger,
			"orderSvc.UpdateShipStatus 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.id", req.ID),
			zap.Any("req.ship_status", req.ShipStatus),
			zap.Any("req.express_company", req.ExpressCompany),
			zap.Any("req.express_no", req.ExpressNo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 后台变更订单状态
// @Description 变更订单状态（确认收货、取消订单等），仅变更状态字段，不涉及退款与库存回滚；置为已取消时写取消时间
// @Tags 订单管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.OrderUpdateOrderStatusReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/order/status [post]
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员信息
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.OrderUpdateOrderStatusReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.OrderLogger, "UpdateOrderStatus 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.orderSvc.UpdateOrderStatus(ctx, order.UpdateOrderStatusReq{
		ID:          req.ID,
		OrderStatus: enum.OrderStatus(*req.OrderStatus),
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.OrderLogger,
			"orderSvc.UpdateOrderStatus 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.id", req.ID),
			zap.Any("req.order_status", req.OrderStatus),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 后台变更订单收货信息
// @Description 变更订单收货信息，仅改订单表单条记录。虚拟订单只接受 receiver_email，实体订单只接受 receiver_name/receiver_phone/receiver_region_code/receiver_detail，传了不适用于该类型的字段返回参数错误
// @Tags 订单管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.OrderUpdateReceiverInfoReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/order/receiver [post]
func (h *Handler) UpdateReceiverInfo(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员信息
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.OrderUpdateReceiverInfoReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.OrderLogger, "UpdateReceiverInfo 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.orderSvc.UpdateReceiverInfo(ctx, order.UpdateReceiverInfoReq{
		ID:                 req.ID,
		ReceiverName:       req.ReceiverName,
		ReceiverPhone:      req.ReceiverPhone,
		ReceiverRegionCode: req.ReceiverRegionCode,
		ReceiverDetail:     req.ReceiverDetail,
		ReceiverEmail:      req.ReceiverEmail,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.OrderLogger,
			"orderSvc.UpdateReceiverInfo 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.id", req.ID),
			zap.Any("req.receiver_name", req.ReceiverName),
			zap.Any("req.receiver_phone", req.ReceiverPhone),
			zap.Any("req.receiver_region_code", req.ReceiverRegionCode),
			zap.Any("req.receiver_detail", req.ReceiverDetail),
			zap.Any("req.receiver_email", req.ReceiverEmail),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}
