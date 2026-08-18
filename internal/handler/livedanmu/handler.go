package livedanmu

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livedanmu"
	"go.uber.org/zap"
)

// Handler 直播控制 HTTP 接口处理器
type Handler struct {
	livedanmuSvc *livedanmu.Service
}

// New 创建 Handler 实例
func New(livedanmuSvc *livedanmu.Service) *Handler {
	return &Handler{livedanmuSvc: livedanmuSvc}
}

// @Summary 获取全部房间ID
// @Description 获取弹幕记录的所有房间ID，用于列表选定房间搜索
// @Tags 弹幕管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.LiveDanmuFetchRoomGroupsResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/livedanmu/room [post]
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
	svcResp, errCode, err := h.livedanmuSvc.FetchRoomGroups(ctx)
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveDanmuLogger,
			"livedanmuSvc.FetchRoomGroups 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveDanmuFetchRoomGroupsResp{
		Option: toFetchRoomGroupsItems(svcResp),
	})
}

// @Summary 分页查询弹幕列表
// @Description 分页获取弹幕列表，支持按房间ID，用户信息，弹幕信息，发送时间进行筛选
// @Tags 弹幕管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveDanmuListPageReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.LiveDanmuListPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/livedanmu/list [post]
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
	var req input.LiveDanmuListPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.LiveDanmuLogger,
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
		end := (ts[len(ts)-1] / 1000) + 86399
		SendAtStart = &start
		SendAtEnd = &end
	}
	// 执行请求
	svcResp, errCode, err := h.livedanmuSvc.ListPage(ctx, livedanmu.ListPageReq{
		PageResp: livedanmu.PageResp{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
		},
		RoomID:      req.RoomID,
		UID:         req.UID,
		Uname:       req.Uname,
		Msg:         req.Msg,
		SendAtStart: SendAtStart,
		SendAtEnd:   SendAtEnd,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.LiveDanmuLogger,
			"livedanmuSvc.ListPage 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int("req.pageNo", req.PageNo),
			zap.Int("req.pageSize", req.PageSize),
			zap.Any("req.room_id", req.RoomID),
			zap.Any("req.uid", req.UID),
			zap.Any("req.uname", req.Uname),
			zap.Any("req.msg", req.Msg),
			zap.Any("req.send_at", req.SendAt),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.LiveDanmuListPageResp{
		Total:    svcResp.Total,
		PageData: toLiveDanmuListItems(svcResp.PageData),
	})
}
