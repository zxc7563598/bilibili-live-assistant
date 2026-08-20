package livegift

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livegift"
	"go.uber.org/zap"
)

// secondsPerDay 一天的秒数，用于将日期范围结束时间戳推到当天最后一秒（23:59:59）
const secondsPerDay = 24 * 60 * 60

// Handler 礼物列表 HTTP 接口处理器
type Handler struct {
	livegiftSvc *livegift.Service
}

// New 创建 Handler 实例
func New(livegiftSvc *livegift.Service) *Handler {
	return &Handler{livegiftSvc: livegiftSvc}
}

// @Summary 获取全部房间ID
// @Description 获取礼物记录的所有房间ID，用于列表选定房间搜索
// @Tags 礼物管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.LiveGiftFetchRoomGroupsResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/livegift/room [post]
func (h *Handler) FetchRoomGroups(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.livegiftSvc.FetchRoomGroups(ctx)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveGiftLogger,
			"livegiftSvc.FetchRoomGroups 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveGiftFetchRoomGroupsResp{
		Option: toFetchRoomGroupsItems(svcResp),
	})
}

// @Summary 分页查询礼物列表
// @Description 分页获取礼物列表，支持按房间ID，用户信息，礼物信息，发送时间进行筛选
// @Tags 礼物管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveGiftListPageReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveGiftListPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/livegift/list [post]
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
	var req input.LiveGiftListPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveGiftLogger,
			"ListPage 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 处理时间
	var SendAtStart, SendAtEnd *int64
	if req.SendAt != nil && len(*req.SendAt) >= 2 {
		ts := *req.SendAt
		start := ts[0] / 1000
		end := (ts[len(ts)-1] / 1000) + secondsPerDay - 1
		SendAtStart = &start
		SendAtEnd = &end
	}
	// 执行请求
	svcResp, errCode, err := h.livegiftSvc.ListPage(ctx, livegift.ListPageReq{
		PageResp: livegift.PageResp{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
		},
		RoomID:      req.RoomID,
		UID:         req.UID,
		Uname:       req.Uname,
		GiftName:    req.GiftName,
		GiftType:    req.GiftType,
		Original:    req.Original,
		SendAtStart: SendAtStart,
		SendAtEnd:   SendAtEnd,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveGiftLogger,
			"livegiftSvc.ListPage 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int("req.pageNo", req.PageNo),
			zap.Int("req.pageSize", req.PageSize),
			zap.Any("req.room_id", req.RoomID),
			zap.Any("req.uid", req.UID),
			zap.Any("req.uname", req.Uname),
			zap.Any("req.gift_name", req.GiftName),
			zap.Any("req.gift_type", req.GiftType),
			zap.Any("req.original", req.Original),
			zap.Any("req.send_at", req.SendAt),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveGiftListPageResp{
		Total:    svcResp.Total,
		PageData: toLiveGiftListItems(svcResp.PageData),
		Stats: resp.LiveGiftListPageStats{
			TotalNum:    svcResp.Stats.TotalNum,
			TotalAmount: svcResp.Stats.TotalAmount,
		},
	})
}
