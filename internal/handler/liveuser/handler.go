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
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/robotconfig"
	"go.uber.org/zap"
)

// Handler 直播控制 HTTP 接口处理器
type Handler struct {
	liveuserSvc    *liveuser.Service
	robotConfigSvc *robotconfig.Service
}

// New 创建 Handler 实例
func New(liveuserSvc *liveuser.Service, robotConfigSvc *robotconfig.Service) *Handler {
	return &Handler{
		liveuserSvc:    liveuserSvc,
		robotConfigSvc: robotConfigSvc,
	}
}

// @Summary 分页查询用户列表
// @Description 分页获取用户列表，支持按用户信息进行筛选
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
		response.Error(c, lang, 20001)
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
		PageResp: liveuser.PageResp{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
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

// @Summary 获取用户每日分析数据
// @Description 获取用户指定月份在直播间的发言/打赏汇总
// @Tags 用户管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserUserMonthlyAnalysisReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserUserMonthlyAnalysisResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/liveuser/monthly [post]
func (h *Handler) UserMonthlyAnalysis(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取请求参数
	var req input.LiveUserUserMonthlyAnalysisReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"UserMonthlyAnalysis 参数异常",
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

// @Summary 获取用户弹幕分析
// @Description 获取用户弹幕发言分析
// @Tags 用户管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserUserDanmuAnalysisReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserUserDanmuAnalysisResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/liveuser/danmu [post]
func (h *Handler) UserDanmuAnalysis(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取请求参数
	var req input.LiveUserUserDanmuAnalysisReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"UserDanmuAnalysis 参数异常",
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

// 移动端 ------------------

// @Summary 判断用户账号是否存在
// @Description 校验指定账号（UID）是否已存在，供登录页在提交前进行预检
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
// @Description 用户使用账号（UID）与密码登录，成功后返回 access_token 与 refresh_token，用于后续接口鉴权
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
// @Description 使用 refresh_token 换取新的 access_token 与 refresh_token，以延长会话有效期
// @Tags 移动端
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveUserRefreshReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveUserLoginResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/liveuser/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取请求参数
	var req input.LiveUserRefreshReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"Refresh 参数异常",
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
// @Description 清除用户登录态，使当前 access_token 与 refresh_token 立即失效
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
		response.Error(c, lang, 20001)
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

// @Summary 获取用户基本信息
// @Description 获取当前登录用户的基本信息（头像、昵称、积分、星光等），供移动端商城展示
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
		response.Error(c, lang, 20001)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.liveuserSvc.UserInfo(ctx, userInfo.UserID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveUserLogger,
			"liveuserSvc.UserInfo 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
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
		response.Error(c, lang, 20001)
		return
	}
	// 执行请求
	roomID := h.robotConfigSvc.GetRoomID()
	// 返回结果
	response.Success(c, lang, resp.LiveUserGetRoomIDResp{
		RoomID: roomID,
	})
}
