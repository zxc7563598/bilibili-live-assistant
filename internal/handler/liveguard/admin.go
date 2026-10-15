package liveguard

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveguard"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"
	"go.uber.org/zap"
)

// @Summary 分页查询大航海名单
// @Description 按月份与大航海等级查询当月持有该身份的人员名单，粒度为「用户 × 月份」
// @Tags 大航海名单
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveGuardListPageReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveGuardListPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/liveguard/list [post]
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
	// 解析并校验请求参数
	var req input.LiveGuardListPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.LiveGuardLogger, "参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.liveguardSvc.ListPage(ctx, liveguard.ListPageReq{
		PageResp: pagination.PageResp{
			PageNo:    req.PageNo,
			PageSize:  req.PageSize,
			SortField: req.SortField,
			SortOrder: req.SortOrder,
		},
		GuardLevel: &req.GuardLevel,
		Month:      req.Month,
		UID:        req.UID,
		Uname:      req.Uname,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveGuardLogger,
			"liveguardSvc.ListPage 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.String("month", req.Month),
			zap.Int("guardLevel", req.GuardLevel),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveGuardListPageResp{
		Total:    svcResp.Total,
		PageData: toLiveGuardListItems(svcResp.PageData),
	})
}

// @Summary 补录大航海开通记录
// @Description 补录一条漏记的大航海开通：补齐礼物记录、累计消费金额、到期时间、开通台账与积分/星光奖励，不发弹幕、不走黑名单赎回
// @Tags 大航海名单
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveGuardBackfillReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/liveguard/backfill [post]
func (h *Handler) Backfill(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 解析并校验请求参数
	var req input.LiveGuardBackfillReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.LiveGuardLogger, "参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.liveguardSvc.Backfill(ctx, liveguard.BackfillReq{
		UID:        req.UID,
		Uname:      req.Uname,
		RoomID:     req.RoomID,
		GuardLevel: req.GuardLevel,
		OpenAt:     req.OpenAt,
		Num:        req.Num,
		Price:      req.Price,
		GiftID:     req.GiftID,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveGuardLogger,
			"liveguardSvc.Backfill 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("uid", req.UID),
			zap.Int("guardLevel", req.GuardLevel),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}
