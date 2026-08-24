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
	"go.uber.org/zap"
)

// Handler 直播控制 HTTP 接口处理器
type Handler struct {
	liveuserSvc *liveuser.Service
}

// New 创建 Handler 实例
func New(liveuserSvc *liveuser.Service) *Handler {
	return &Handler{liveuserSvc: liveuserSvc}
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
