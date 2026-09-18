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

// @Summary 分页查询用户列表
// @Description 分页查询后台用户列表，支持按 UID 精确匹配、昵称模糊匹配筛选，并可按指定字段排序
// @Tags 用户管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserListPageReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserListPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/liveuser/list [post]
func (h *Handler) ListPage(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取请求参数
	var req input.LiveUserListPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"ListPage 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.liveuserSvc.ListPage(ctx, liveuser.ListPageReq{
		PageResp: pagination.PageResp{
			PageNo:    req.PageNo,
			PageSize:  req.PageSize,
			SortField: req.SortField,
			SortOrder: req.SortOrder,
		},
		UID:   req.UID,
		Uname: req.Uname,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.ListPage 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int("req.pageNo", req.PageNo),
			zap.Int("req.pageSize", req.PageSize),
			zap.Any("req.uid", req.UID),
			zap.Any("req.uname", req.Uname),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveUserListPageResp{
		Total:    svcResp.Total,
		PageData: toLiveUserListItems(svcResp.PageData),
	})
}

// @Summary 获取用户月度数据统计
// @Description 统计指定用户在某一自然月内的每日弹幕数、礼物数、礼物金额，以及当月每天是否开播，用于后台用户分析页的图表展示
// @Tags 用户管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserUserMonthlyAnalysisReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserUserMonthlyAnalysisResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/liveuser/monthly [post]
func (h *Handler) GetUserMonthlyAnalysis(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取请求参数
	var req input.LiveUserUserMonthlyAnalysisReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"GetUserMonthlyAnalysis 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.liveuserSvc.GetUserMonthlyAnalysis(ctx, req.UID, req.Year, req.Month)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.GetUserMonthlyAnalysis 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Any("req.uid", req.UID),
			zap.Any("req.year", req.Year),
			zap.Any("req.month", req.Month),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveUserUserMonthlyAnalysisResp{
		DanmuCount: svcResp.DanmuCount,
		GiftCount:  svcResp.GiftCount,
		GiftAmount: svcResp.GiftAmount,
		LiveDays:   svcResp.LiveDays,
	})
}

// @Summary 获取用户弹幕词频分析
// @Description 统计指定用户历史弹幕中的高频内容，分别返回单词、双词、三词与完整短句四个维度的词频列表，用于后台用户画像分析
// @Tags 用户管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserUserDanmuAnalysisReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserUserDanmuAnalysisResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/liveuser/danmu [post]
func (h *Handler) GetUserDanmuAnalysis(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取请求参数
	var req input.LiveUserUserDanmuAnalysisReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"GetUserDanmuAnalysis 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.liveuserSvc.GetUserDanmuAnalysis(ctx, req.UID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.GetUserDanmuAnalysis 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Any("req.uid", req.UID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveUserUserDanmuAnalysisResp{
		Words:    toLiveUserWordFrequencyItems(svcResp.Words),
		Bigrams:  toLiveUserWordFrequencyItems(svcResp.Bigrams),
		Trigrams: toLiveUserWordFrequencyItems(svcResp.Trigrams),
		Messages: toLiveUserWordFrequencyItems(svcResp.Messages),
	})
}

// @Summary 获取用户详细信息
// @Description 按用户表主键（user_id，非 B站 UID）查询用户基础信息（UID、昵称、头像）及当前剩余积分、星光，用于后台用户详情页展示
// @Tags 用户管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserDetailsReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserUserInfoResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/liveuser/details [post]
func (h *Handler) Details(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取请求参数
	var req input.LiveUserDetailsReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"Details 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.liveuserSvc.GetUserInfo(ctx, req.UserID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.GetUserInfo 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Any("req.user_id", req.UserID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveUserUserInfoResp{
		UID:    svcResp.UID,
		Avatar: svcResp.Avatar,
		Name:   svcResp.Name,
		Points: svcResp.Points,
		Stars:  svcResp.Stars,
	})
}

// @Summary 分页查询用户积分/星光变动记录
// @Description 按用户表主键（user_id，非 B站 UID）分页查询该用户的积分/星光变动流水，支持按资产类型筛选（credit_type：0-星光，1-积分），用于后台用户详情页的资产明细
// @Tags 用户管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserAssetsPageByIdReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserAssetsPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/liveuser/assets [post]
func (h *Handler) ListAssetsPageByID(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取请求参数
	var req input.LiveUserAssetsPageByIdReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"ListAssetsPageByID 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.liveuserSvc.ListUserAssets(ctx, req.UserID, liveuser.UserAssetsPageReq{
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
			zap.Any("adminInfo", adminInfo),
			zap.Int("req.pageNo", req.PageNo),
			zap.Int("req.pageSize", req.PageSize),
			zap.Any("req.user_id", req.UserID),
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

// @Summary 手动调整用户余额
// @Description 按用户表主键（user_id，非 B站 UID）手动变更指定用户的积分/星光余额：credit_type 指定资产类型（0-星光，1-积分），change_type 指定变动方向（0-减少，1-增加），后两者必传且不可省略，change_amount 传正数；服务端原子更新余额并写入变动流水，扣减时余额不足则本次操作失败
// @Tags 用户管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserSaveBalanceReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/liveuser/save-assets [post]
func (h *Handler) SaveBalance(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取请求参数
	var req input.LiveUserSaveBalanceReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"SaveBalance 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	// credit_type / change_type 已由 binding:"required" 保证非空，此处可安全解引用
	errCode, err := h.liveuserSvc.SaveBalance(ctx, adminInfo.AdminID, req.UserID, *req.CreditType, *req.ChangeType, req.ChangeAmount, req.Remark)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.SaveBalance 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Any("req.user_id", req.UserID),
			zap.Any("req.credit_type", *req.CreditType),
			zap.Any("req.change_type", *req.ChangeType),
			zap.Any("req.change_amount", req.ChangeAmount),
			zap.Any("req.remark", req.Remark),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 重置用户密码
// @Description 按用户表主键（user_id，非 B站 UID）直接重置指定用户的登录密码，无需校验旧密码；重置成功后服务端会同时清除该用户的登录态（未启用 Redis 时 access_token 在有效期届满前仍可用），用户需重新登录
// @Tags 用户管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserResetPasswordReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/liveuser/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取请求参数
	var req input.LiveUserResetPasswordReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"ResetPassword 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.liveuserSvc.ResetPassword(ctx, req.UserID, req.Password)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.ResetPassword 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Any("req.user_id", req.UserID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 密码已落库，本次操作即成功；踢登录态只是附带动作，失败仅记日志不影响返回
	if logoutCode, logoutErr := h.liveuserSvc.Logout(ctx, req.UserID); logoutCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.Logout 调用失败（密码已重置，登录态可能未失效）",
			logoutCode,
			logoutErr,
			zap.Any("req.user_id", req.UserID),
		)
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// 移动端 ------------------
