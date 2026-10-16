package address

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/address"
	"go.uber.org/zap"
)

// @Summary 后台查询指定用户的收货地址列表
// @Description 按 B站 uid 查询该用户的收货地址，type 省略表示不限类型；供后台代用户下单时选地址
// @Tags 收货地址
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AddressAdminListReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.AddressGetAddressListResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/address/list [post]
func (h *Handler) AdminListAddresses(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员信息
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.AddressAdminListReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.AddressLogger, "AdminListAddresses 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.addressSvc.ListAddressesByUID(ctx, req.UID, req.Type)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AddressLogger,
			"addressSvc.ListAddressesByUID 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.uid", req.UID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	list := make([]resp.AddressItem, 0, len(svcResp))
	for _, item := range svcResp {
		list = append(list, toAddressItemResp(item))
	}
	response.Success(c, lang, resp.AddressGetAddressListResp{List: list})
}

// @Summary 后台给指定用户新建收货地址
// @Description 按 B站 uid 替用户新建一条收货地址，虚拟地址需邮箱、实体地址需手机号与地区详址；返回新地址 ID
// @Tags 收货地址
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AddressAdminSaveReq true "请求参数"
// @Success 200 {object} response.Response{data=int64} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/address/save [post]
func (h *Handler) AdminSaveAddress(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员信息
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.AddressAdminSaveReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.AddressLogger, "AdminSaveAddress 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	addressID, errCode, err := h.addressSvc.SaveAddressByUID(ctx, req.UID, address.AddressReq{
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
			"addressSvc.SaveAddressByUID 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.uid", req.UID),
			zap.Int("req.type", intVal(req.Type)),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, addressID)
}
