package liveuser

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"
	"go.uber.org/zap"
)

// @Summary 判断用户账号是否存在
// @Description 校验指定账号（UID）在用户表中是否已存在，供移动端登录页在提交前预检，不会请求 B 站接口
// @Tags 移动端
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserExistsAccountReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserExistsAccountResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/liveuser/account [post]
func (h *Handler) ExistsAccount(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取请求参数
	var req input.LiveUserExistsAccountReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"ExistsAccount 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	exist, errCode, err := h.liveuserSvc.ExistsAccount(ctx, req.Account)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.ExistsAccount 调用失败",
			errCode,
			err,
			zap.Any("req.account", req.Account),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveUserExistsAccountResp{
		Exist: exist,
	})
}

// @Summary 用户登录
// @Description 用户使用账号（UID）与密码登录，成功后返回 access_token 与 refresh_token，用于后续接口鉴权；首次登录且系统开启自动注册时会自动创建账号，尚无密码的账号则以本次输入的密码作为登录密码
// @Tags 移动端
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserLoginReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserLoginResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/liveuser/login [post]
func (h *Handler) Login(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取请求参数
	var req input.LiveUserLoginReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"Login 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.liveuserSvc.Login(ctx, req.Account, req.Password)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.Login 调用失败",
			errCode,
			err,
			zap.Any("req.account", req.Account),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveUserLoginResp{
		AccessToken:  svcResp.AccessToken,
		RefreshToken: svcResp.RefreshToken,
	})
}

// @Summary 刷新登录凭证
// @Description 使用 refresh_token 换取新的 access_token 与 refresh_token，以延长登录有效期；刷新后原 refresh_token 随即失效，请改用本次返回的新凭证
// @Tags 移动端
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserRefreshReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserLoginResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/liveuser/refresh [post]
func (h *Handler) RefreshLogin(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取请求参数
	var req input.LiveUserRefreshReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"RefreshLogin 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.liveuserSvc.RefreshLogin(ctx, req.Token)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.RefreshLogin 调用失败",
			errCode,
			err,
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveUserLoginResp{
		AccessToken:  svcResp.AccessToken,
		RefreshToken: svcResp.RefreshToken,
	})
}

// @Summary 退出登录
// @Description 退出当前登录，清除服务端保存的登录凭证，使当前 access_token 与 refresh_token 立即失效
// @Tags 移动端
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Security BearerAuth
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/shop/liveuser/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 执行请求
	errCode, err := h.liveuserSvc.Logout(ctx, userInfo.UserID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.Logout 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 修改用户密码
// @Description 已登录用户校验旧密码后修改登录密码，旧密码错误会返回对应错误码；修改成功后当前登录态仍然有效，是否重新登录由前端自行决定
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserChangePasswordReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/shop/liveuser/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取请求参数
	var req input.LiveUserChangePasswordReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"ChangePassword 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.liveuserSvc.ChangePassword(ctx, userInfo.UserID, req.OldPassword, req.NewPassword)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.ChangePassword 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 获取用户基本信息
// @Description 获取当前登录用户自身的基本信息（UID、昵称、头像、剩余积分与星光），用户身份由登录态解析，供移动端商城展示
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.LiveUserUserInfoResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/liveuser/info [post]
func (h *Handler) GetUserInfo(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.liveuserSvc.GetUserInfo(ctx, userInfo.UserID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.GetUserInfo 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveUserUserInfoResp{
		UID:     svcResp.UID,
		Avatar:  svcResp.Avatar,
		Name:    svcResp.Name,
		Points:  svcResp.Points,
		Stars:   svcResp.Stars,
		VipType: svcResp.VipType,
	})
}

// @Summary 获取直播间房间号
// @Description 获取系统当前监听的直播间房间号，供移动端商城跳转至主播直播间使用
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.LiveUserGetRoomIDResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/liveuser/room-id [post]
func (h *Handler) GetRoomID(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	_, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 执行请求
	roomID := h.robotConfigSvc.GetRoomID()
	// 返回结果
	response.Success(c, lang, resp.LiveUserGetRoomIDResp{
		RoomID: roomID,
	})
}

// @Summary 分页查询我的积分/星光变动记录
// @Description 分页查询积分/星光变动流水，支持按资产类型筛选（credit_type：0-星光，1-积分）；查询对象固定为当前登录用户本人，无需也不能传入用户 ID
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserAssetsPageReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserAssetsPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/liveuser/assets [post]
func (h *Handler) ListAssetsPage(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取请求参数
	var req input.LiveUserAssetsPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"ListAssetsPage 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.liveuserSvc.ListUserAssets(ctx, userInfo.UserID, liveuser.UserAssetsPageReq{
		PageResp: pagination.PageResp{
			PageNo:    req.PageNo,
			PageSize:  req.PageSize,
			SortField: req.SortField,
			SortOrder: req.SortOrder,
		},
		CreditType: req.CreditType,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.ListUserAssets 调用失败",
			errCode,
			err,
			zap.Any("userinfo", userInfo),
			zap.Int("req.pageNo", req.PageNo),
			zap.Int("req.pageSize", req.PageSize),
			zap.Any("req.credit_type", req.CreditType),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveUserAssetsPageResp{
		Total:    svcResp.Total,
		PageData: toLiveUserAssetsPageItems(svcResp.PageData),
	})
}
