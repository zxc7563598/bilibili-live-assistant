package live

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	liveSvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/live"
)

// Handler 直播控制 HTTP 接口处理器
type Handler struct {
	liveSvc *liveSvc.Service
}

// New 创建 Handler 实例
func New(liveSvc *liveSvc.Service) *Handler {
	return &Handler{liveSvc: liveSvc}
}

// @Summary 获取 B站 扫码登录二维码
// @Description 获取 B站 扫码登录二维码，返回的链接需要在前端转换为二维码由用户使用 B 站客户端进行扫码登陆
// @Tags 直播控制
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.LiveQRCodeResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/live/login/qrcode [post]
func (h *Handler) GetQRCode(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.liveSvc.GetQRCode(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.LiveLogger, "liveSvc.GetQRCode 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, resp.LiveQRCodeResp{
		URL:       svcResp.URL,
		QrcodeKey: svcResp.QrcodeKey,
	})
}

// @Summary 轮询扫码状态
// @Description 根据获取二维码时得到的 qrcodeKey 查询二维码扫描状态，完成用户登录
// @Tags 直播控制
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveQRCodePollReq true "轮询扫码状态参数"
// @Success 200 {object} response.Response{data=resp.LivePollQRCodeResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/live/login/poll [post]
func (h *Handler) PollQRCode(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	var req input.LiveQRCodePollReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.LiveLogger, "PollQRCode 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	svcResp, errCode, err := h.liveSvc.PollQRCode(ctx, req.QrcodeKey)
	if errCode != 0 {
		handler.ErrorLog(logger.LiveLogger, "liveSvc.PollQRCode 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, resp.LivePollQRCodeResp{
		Status:    svcResp.Status,
		Message:   svcResp.Message,
		IsScanned: svcResp.IsScanned,
		IsSuccess: svcResp.IsSuccess,
		IsExpired: svcResp.IsExpired,
	})
}

// @Summary 获取简易机器人信息
// @Description 简单获取当前登录的机器人信息，此为本系统的登录状态，并非代表 B 站状态
// @Tags 直播控制
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.LiveLoginStatusResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/live/login/status [post]
func (h *Handler) GetLoginStatus(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	_ = ctx
	svcResp, errCode, err := h.liveSvc.GetLoginStatus(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.LiveLogger, "liveSvc.GetLoginStatus 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, resp.LiveLoginStatusResp{
		IsLoggedIn: svcResp.IsLoggedIn,
		UID:        svcResp.UID,
		Username:   svcResp.Username,
		Buvid:      svcResp.Buvid,
	})
}

// @Summary 清除机器人登录信息
// @Description 清除机器人登录信息，此为本系统的登录状态，并非代表 B 站状态
// @Tags 直播控制
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/live/login/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	_ = ctx
	errCode, err := h.liveSvc.Logout(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.LiveLogger, "liveSvc.Logout 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 添加/更换监听直播间房间号
// @Description 用于设置监听的直播间房间号，当正在监听直播间时会自动重连，不会影响监听状态
// @Tags 直播控制
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.LiveRoomUpdateReq true "房间号参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/live/room/update [post]
func (h *Handler) UpdateRoom(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	var req input.LiveRoomUpdateReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.LiveLogger, "UpdateRoom 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.liveSvc.UpdateRoom(ctx, req.RoomID)
	if errCode != 0 {
		handler.ErrorLog(logger.LiveLogger, "liveSvc.UpdateRoom 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 启动 WebSocket 监听
// @Description 用于启动 WebSocket 监听，需要在机器人登录 & 已设置直播间房间号 & 未监听 WebSocket 的情况下使用，需要自行验证状态
// @Tags 直播控制
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/live/listener/start [post]
func (h *Handler) StartListener(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	errCode, err := h.liveSvc.StartListener(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.LiveLogger, "liveSvc.StartListener 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 停止 WebSocket 监听
// @Description 用于停止 WebSocket 监听
// @Tags 直播控制
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/live/listener/stop [post]
func (h *Handler) StopListener(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)

	_ = ctx
	errCode, err := h.liveSvc.StopListener()
	if errCode != 0 {
		handler.ErrorLog(logger.LiveLogger, "liveSvc.StopListener 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 获取 WebSocket 状态
// @Description 用于判断当前 WebSocket 状态，以及连接情况下的简易信息
// @Tags 直播控制
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.LiveListenerStatusResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/live/listener/status [post]
func (h *Handler) GetListenerStatus(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	_ = ctx
	svcResp, errCode, err := h.liveSvc.GetListenerStatus(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.LiveLogger, "liveSvc.GetListenerStatus 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, resp.LiveListenerStatusResp{
		IsRunning:  svcResp.IsRunning,
		RoomID:     svcResp.RoomID,
		StartTime:  svcResp.StartTime,
		Uptime:     svcResp.Uptime,
		MsgCount:   svcResp.MsgCount,
		DanmuCount: svcResp.DanmuCount,
		GiftCount:  svcResp.GiftCount,
	})
}
