package address

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/address"
	"go.uber.org/zap"
)

type Handler struct {
	addressSvc *address.Service
}

func New(addressSvc *address.Service) *Handler {
	return &Handler{
		addressSvc: addressSvc,
	}
}

// @Summary 获取指定收货地址类型的默认地址
// @Description 返回当前用户在指定收货地址类型（0 虚拟，1 实体）下的默认收货地址，供下单确认等页面默认回填使用；该类型下暂无默认地址时返回空对象。
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AddressGetDefaultAddressReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.AddressGetDefaultAddressResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/address/default [post]
func (h *Handler) GetDefaultAddress(c *gin.Context) {
	var req input.AddressGetDefaultAddressReq
	ur, ok := handler.BindUserRequest(c, logger.AddressLogger, "GetDefaultAddress", &req)
	if !ok {
		return
	}
	// 执行请求
	svcResp, errCode, err := h.addressSvc.GetDefaultAddress(ur.Ctx, ur.User.UserID, req.Type)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AddressLogger,
			"addressSvc.GetDefaultAddress 调用失败",
			errCode,
			err,
			zap.Any("userInfo", ur.User),
			zap.Int("req.type", req.Type),
		)
		response.Error(c, ur.Lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, ur.Lang, resp.AddressGetDefaultAddressResp{
		AddressItem: toAddressItemResp(svcResp),
	})
}

// @Summary 获取收货地址列表
// @Description 按收货地址类型获取当前用户的全部收货地址，不分页返回，常用于地址管理页面的列表展示；type 缺省时返回全部类型。
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AddressGetAddressListReq false "请求参数（可省略；省略或未传 type 时返回全部类型）"
// @Success 200 {object} response.Response{data=resp.AddressGetAddressListResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/address/list [post]
func (h *Handler) GetAddressList(c *gin.Context) {
	var req input.AddressGetAddressListReq
	ur, ok := handler.BindUserRequestAllowEmpty(c, logger.AddressLogger, "GetAddressList", &req)
	if !ok {
		return
	}
	// 执行请求
	svcResp, errCode, err := h.addressSvc.GetAddressList(ur.Ctx, ur.User.UserID, req.Type)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AddressLogger,
			"addressSvc.GetAddressList 调用失败",
			errCode,
			err,
			zap.Any("userInfo", ur.User),
			zap.Any("req.type", req.Type),
		)
		response.Error(c, ur.Lang, errCode)
		return
	}
	// 返回结果
	items := make([]resp.AddressItem, 0, len(svcResp))
	for _, item := range svcResp {
		items = append(items, toAddressItemResp(item))
	}
	response.Success(c, ur.Lang, resp.AddressGetAddressListResp{
		List: items,
	})
}

// @Summary 获取收货地址详情
// @Description 返回当前用户某条收货地址的完整信息，常用于编辑前的信息回显；仅可查询归属当前用户本人的地址，否则返回错误。
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AddressGetAddressByIDReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.AddressGetAddressByIDResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/address/detail [post]
func (h *Handler) GetAddressByID(c *gin.Context) {
	var req input.AddressGetAddressByIDReq
	ur, ok := handler.BindUserRequest(c, logger.AddressLogger, "GetAddressByID", &req)
	if !ok {
		return
	}
	// 执行请求
	svcResp, errCode, err := h.addressSvc.GetAddressByID(ur.Ctx, ur.User.UserID, int64(req.ID))
	if errCode != 0 {
		handler.ErrorLog(
			logger.AddressLogger,
			"addressSvc.GetAddressByID 调用失败",
			errCode,
			err,
			zap.Any("userInfo", ur.User),
			zap.Int("req.id", req.ID),
		)
		response.Error(c, ur.Lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, ur.Lang, resp.AddressGetAddressByIDResp{
		AddressItem: toAddressItemResp(svcResp),
	})
}

// @Summary 新增/修改收货地址
// @Description 新增或修改当前用户的收货地址：携带 id 视为修改该条地址，不携带 id 视为新增；修改时省略的字段保留原值，is_default 为 1 时该地址被设为对应类型下的默认地址。
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AddressSaveAddressReq true "请求参数"
// @Success 200 {object} response.Response{data=int64} "统一响应（code=0成功，data 为地址 ID）"
// @Router /api/shop/address/save [post]
func (h *Handler) SaveAddress(c *gin.Context) {
	var req input.AddressSaveAddressReq
	ur, ok := handler.BindUserRequest(c, logger.AddressLogger, "SaveAddress", &req)
	if !ok {
		return
	}
	// 执行请求
	addressID, errCode, err := h.addressSvc.SaveAddress(ur.Ctx, ur.User.UserID, address.AddressReq{
		ID:         req.ID,
		Name:       req.Name,
		Phone:      req.Phone,
		RegionCode: req.RegionCode,
		Region:     req.Region,
		Detail:     req.Detail,
		Email:      req.Email,
		Type:       req.Type,
		IsDefault:  req.IsDefault,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.AddressLogger,
			"addressSvc.SaveAddress 调用失败",
			errCode,
			err,
			zap.Any("userInfo", ur.User),
			zap.Any("req.id", req.ID),
			zap.Int("req.type", intVal(req.Type)),
			zap.Int("req.is_default", intVal(req.IsDefault)),
		)
		response.Error(c, ur.Lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, ur.Lang, addressID)
}

// @Summary 删除收货地址
// @Description 删除当前用户的某条收货地址（软删除）；仅可删除归属当前用户本人且存在的地址，否则返回错误。
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AddressDeleteAddressReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/shop/address/delete [post]
func (h *Handler) DeleteAddress(c *gin.Context) {
	var req input.AddressDeleteAddressReq
	ur, ok := handler.BindUserRequest(c, logger.AddressLogger, "DeleteAddress", &req)
	if !ok {
		return
	}
	// 执行请求
	errCode, err := h.addressSvc.DeleteAddress(ur.Ctx, ur.User.UserID, req.ID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AddressLogger,
			"addressSvc.DeleteAddress 调用失败",
			errCode,
			err,
			zap.Any("userInfo", ur.User),
			zap.Int64("req.id", req.ID),
		)
		response.Error(c, ur.Lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, ur.Lang, nil)
}
