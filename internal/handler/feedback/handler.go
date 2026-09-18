package feedback

import "github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/feedback"
	"go.uber.org/zap"
)

type Handler struct {
	feedbackSvc *feedback.Service
}

func New(feedbackSvc *feedback.Service) *Handler {
	return &Handler{
		feedbackSvc: feedbackSvc,
	}
}

// @Summary 提交用户投诉/反馈
// @Description 当前登录用户提交一条投诉/反馈（问题类型、内容、联系方式），提交后由平台侧处理；联系方式仅用于平台回访
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.FeedbackSubmitReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/shop/feedback/submit [post]
func (h *Handler) Submit(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取请求参数
	var req input.FeedbackSubmitReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.FeedbackLogger,
			"Submit 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.feedbackSvc.Add(ctx, userInfo.UserID, feedback.CreateReq{
		Type:    req.Type,
		Content: req.Content,
		Contact: req.Contact,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.FeedbackLogger,
			"feedbackSvc.Add 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
			zap.String("req.type", req.Type),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 后台分页查询投诉列表
// @Description 分页查询用户提交的投诉/反馈，联查用户表返回用户 UID 与昵称；支持按 UID 精确、按昵称模糊筛选。列表不下发投诉正文，正文请调详情接口
// @Tags 投诉管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.FeedbackListPageReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.FeedbackListPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/feedback/list [post]
func (h *Handler) ListPage(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员信息
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取参数
	var req input.FeedbackListPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.FeedbackLogger, "ListPage 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.feedbackSvc.ListPage(ctx, feedback.ListPageReq{
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
			logger.FeedbackLogger,
			"feedbackSvc.ListPage 调用失败",
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
	response.Success(c, lang, resp.FeedbackListPageResp{
		Total:    svcResp.Total,
		PageData: toFeedbackListPageItem(svcResp.PageData),
	})
}

// @Summary 后台获取投诉详情
// @Description 根据投诉ID返回投诉全部字段（含投诉正文），以及投诉用户的 uid/uname
// @Tags 投诉管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.FeedbackDetailsReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.FeedbackDetailsResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/feedback/details [post]
func (h *Handler) Details(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员信息
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取参数
	var req input.FeedbackDetailsReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.FeedbackLogger, "Details 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.feedbackSvc.Details(ctx, req.ID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.FeedbackLogger,
			"feedbackSvc.Details 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.id", req.ID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, toFeedbackDetailsResp(svcResp))
}
