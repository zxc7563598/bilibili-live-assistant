package livepk

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livepk"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
	"go.uber.org/zap"
)

// @Summary 获取全部房间ID
// @Description 获取礼物记录的所有房间ID，用于列表选定房间搜索
// @Tags 礼物管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.RoomGroupOptionsResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/pk/room [post]
func (h *Handler) FetchRoomGroups(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.livepkSvc.FetchRoomGroups(ctx)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LivePkLogger,
			"livepkSvc.FetchRoomGroups 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.RoomGroupOptionsResp{
		Option: handler.ToRoomGroupOptions(svcResp),
	})
}

// @Summary 分页查询 PK 对战记录
// @Description 分页获取 PK 对战记录列表，支持按房间ID、对方UID、对方昵称、我方胜负、PK开始时间进行筛选
// @Tags PK管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LivePkLogListPageReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LivePkLogListPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/pk/list [post]
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
	var req input.LivePkLogListPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LivePkLogger,
			"ListPage 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 处理时间区间：起始取当天 0 点，结束推到当天最后一秒
	var startAtStart, startAtEnd *int64
	if req.StartAt != nil {
		startAtStart, startAtEnd = timeutil.SecondRange(*req.StartAt)
	}
	// 执行请求
	svcResp, errCode, err := h.livepkSvc.ListPage(ctx, livepk.ListPageReq{
		PageResp: pagination.PageResp{
			PageNo:    req.PageNo,
			PageSize:  req.PageSize,
			SortField: req.SortField,
			SortOrder: req.SortOrder,
		},
		RoomID:       req.RoomID,
		RivalUID:     req.RivalUID,
		RivalUname:   req.RivalUname,
		Result:       req.Result,
		StartAtStart: startAtStart,
		StartAtEnd:   startAtEnd,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.LivePkLogger,
			"livepkSvc.ListPage 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int("req.pageNo", req.PageNo),
			zap.Int("req.pageSize", req.PageSize),
			zap.Any("req.room_id", req.RoomID),
			zap.Any("req.rival_uid", req.RivalUID),
			zap.Any("req.rival_uname", req.RivalUname),
			zap.Any("req.result", req.Result),
			zap.Any("req.start_at", req.StartAt),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LivePkLogListPageResp{
		Total:    svcResp.Total,
		PageData: toLivePkLogListItems(svcResp.PageData),
		Stats: resp.LivePkLogListPageStats{
			TotalNum: svcResp.Stats.TotalNum,
			WinNum:   svcResp.Stats.WinNum,
			LoseNum:  svcResp.Stats.LoseNum,
		},
	})
}
