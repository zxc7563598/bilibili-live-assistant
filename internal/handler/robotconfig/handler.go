package robotconfig

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	robotconfigsvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/robotconfig"
)

type Handler struct {
	robotConfigSvc *robotconfigsvc.Service
}

func New(robotConfigSvc *robotconfigsvc.Service) *Handler {
	return &Handler{robotConfigSvc: robotConfigSvc}
}

// @Summary 获取房间模块配置
// @Description 获取房间模块的监听、用户名裁剪等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.RoomConfigResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/room/get [post]
func (h *Handler) GetRoom(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.robotConfigSvc.GetRoomConfig(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.GetRoomConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, toRoomConfigResp(svcResp))
}

// @Summary 更新房间模块配置
// @Description 更新房间模块的监听、用户名裁剪等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.RoomConfigReq true "房间配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/room/apply [post]
func (h *Handler) ApplyRoom(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)

	var req input.RoomConfigReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.RobotConfigLogger, "ApplyRoom 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.robotConfigSvc.ApplyRoomConfig(ctx, robotconfigsvc.RoomConfigReq{
		IsListening:   req.IsListening,
		MaxNameLength: req.MaxNameLength,
		NameTrimMode:  req.NameTrimMode,
	})
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.ApplyRoomConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 获取签到模块配置
// @Description 获取签到模块的触发条件、奖励、回复内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.SignConfigResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/sign/get [post]
func (h *Handler) GetSign(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.robotConfigSvc.GetSignConfig(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.GetSignConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, toSignConfigResp(svcResp))
}

// @Summary 更新签到模块配置
// @Description 更新签到模块的触发条件、奖励、回复内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.SignConfigReq true "签到配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/sign/apply [post]
func (h *Handler) ApplySign(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	var req input.SignConfigReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.RobotConfigLogger, "ApplySign 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.robotConfigSvc.ApplySignConfig(ctx, robotconfigsvc.SignConfigReq{
		Enabled:      req.Enabled,
		Scene:        req.Scene,
		Requirement:  req.Requirement,
		RewardType:   req.RewardType,
		RewardAmount: req.RewardAmount,
		Keyword:      req.Keyword,
		QueryKeyword: req.QueryKeyword,
		SuccessReply: req.SuccessReply,
		FailReply:    req.FailReply,
		RepeatReply:  req.RepeatReply,
		QueryReply:   req.QueryReply,
	})
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.ApplySignConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 获取定时广告模块配置
// @Description 获取定时广告模块的发送间隔、内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.AdConfigResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/ad/get [post]
func (h *Handler) GetAd(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.robotConfigSvc.GetAdConfig(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.GetAdConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, toAdConfigResp(svcResp))
}

// @Summary 更新定时广告模块配置
// @Description 更新定时广告模块的发送间隔、内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AdConfigReq true "定时广告配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/ad/apply [post]
func (h *Handler) ApplyAd(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	var req input.AdConfigReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.RobotConfigLogger, "ApplyAd 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.robotConfigSvc.ApplyAdConfig(ctx, robotconfigsvc.AdConfigReq{
		Enabled:  req.Enabled,
		Scene:    req.Scene,
		Interval: req.Interval,
		SendMode: req.SendMode,
		Content:  req.Content,
	})
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.ApplyAdConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 获取礼物答谢模块配置
// @Description 获取礼物答谢模块的触发门槛、礼物合并、答谢内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.GiftConfigResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/gift/get [post]
func (h *Handler) GetGift(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.robotConfigSvc.GetGiftConfig(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.GetGiftConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, toGiftConfigResp(svcResp))
}

// @Summary 更新礼物答谢模块配置
// @Description 更新礼物答谢模块的触发门槛、礼物合并、答谢内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.GiftConfigReq true "礼物答谢配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/gift/apply [post]
func (h *Handler) ApplyGift(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	var req input.GiftConfigReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.RobotConfigLogger, "ApplyGift 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.robotConfigSvc.ApplyGiftConfig(ctx, robotconfigsvc.GiftConfigReq{
		Enabled:         req.Enabled,
		Scene:           req.Scene,
		Requirement:     req.Requirement,
		ShowCount:       req.ShowCount,
		MergeGift:       req.MergeGift,
		IncludeBlindbox: req.IncludeBlindbox,
		MinBattery:      req.MinBattery,
		Content:         req.Content,
	})
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.ApplyGiftConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 获取PK播报模块配置
// @Description 获取PK播报模块的启用状态、播报内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.PkConfigResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/pk/get [post]
func (h *Handler) GetPk(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.robotConfigSvc.GetPkConfig(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.GetPkConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, toPkConfigResp(svcResp))
}

// @Summary 更新PK播报模块配置
// @Description 更新PK播报模块的启用状态、播报内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.PkConfigReq true "PK播报配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/pk/apply [post]
func (h *Handler) ApplyPk(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)

	var req input.PkConfigReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.RobotConfigLogger, "ApplyPk 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.robotConfigSvc.ApplyPkConfig(ctx, robotconfigsvc.PkConfigReq{
		Enabled: req.Enabled,
		Content: req.Content,
	})
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.ApplyPkConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 获取进房欢迎模块配置
// @Description 获取进房欢迎模块的触发门槛、欢迎内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.WelcomeConfigResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/welcome/get [post]
func (h *Handler) GetWelcome(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.robotConfigSvc.GetWelcomeConfig(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.GetWelcomeConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, toWelcomeConfigResp(svcResp))
}

// @Summary 更新进房欢迎模块配置
// @Description 更新进房欢迎模块的触发门槛、欢迎内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.WelcomeConfigReq true "进房欢迎配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/welcome/apply [post]
func (h *Handler) ApplyWelcome(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	var req input.WelcomeConfigReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.RobotConfigLogger, "ApplyWelcome 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.robotConfigSvc.ApplyWelcomeConfig(ctx, robotconfigsvc.WelcomeConfigReq{
		Enabled:     req.Enabled,
		Scene:       req.Scene,
		Requirement: req.Requirement,
		Content:     req.Content,
	})
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.ApplyWelcomeConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 获取感谢关注模块配置
// @Description 获取感谢关注模块的触发门槛、感谢内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.FollowConfigResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/follow/get [post]
func (h *Handler) GetFollow(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.robotConfigSvc.GetFollowConfig(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.GetFollowConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, toFollowConfigResp(svcResp))
}

// @Summary 更新感谢关注模块配置
// @Description 更新感谢关注模块的触发门槛、感谢内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.FollowConfigReq true "感谢关注配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/follow/apply [post]
func (h *Handler) ApplyFollow(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	var req input.FollowConfigReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.RobotConfigLogger, "ApplyFollow 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.robotConfigSvc.ApplyFollowConfig(ctx, robotconfigsvc.FollowConfigReq{
		Enabled:     req.Enabled,
		Scene:       req.Scene,
		Requirement: req.Requirement,
		Content:     req.Content,
	})
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.ApplyFollowConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 获取感谢分享模块配置
// @Description 获取感谢分享模块的触发门槛、感谢内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.ShareConfigResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/share/get [post]
func (h *Handler) GetShare(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.robotConfigSvc.GetShareConfig(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.GetShareConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, toShareConfigResp(svcResp))
}

// @Summary 更新感谢分享模块配置
// @Description 更新感谢分享模块的触发门槛、感谢内容等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.ShareConfigReq true "感谢分享配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/share/apply [post]
func (h *Handler) ApplyShare(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	var req input.ShareConfigReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.RobotConfigLogger, "ApplyShare 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.robotConfigSvc.ApplyShareConfig(ctx, robotconfigsvc.ShareConfigReq{
		Enabled:     req.Enabled,
		Scene:       req.Scene,
		Requirement: req.Requirement,
		Content:     req.Content,
	})
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.ApplyShareConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 获取自动回复模块配置
// @Description 获取自动回复模块的触发门槛、回复规则等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.ReplyConfigResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/reply/get [post]
func (h *Handler) GetReply(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.robotConfigSvc.GetReplyConfig(ctx)
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.GetReplyConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, toReplyConfigResp(svcResp))
}

// @Summary 更新自动回复模块配置
// @Description 更新自动回复模块的触发门槛、回复规则等配置信息
// @Tags 机器人配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.ReplyConfigReq true "自动回复配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/robot/reply/apply [post]
func (h *Handler) ApplyReply(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	var req input.ReplyConfigReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.RobotConfigLogger, "ApplyReply 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.robotConfigSvc.ApplyReplyConfig(ctx, robotconfigsvc.ReplyConfigReq{
		Enabled:     req.Enabled,
		Scene:       req.Scene,
		Requirement: req.Requirement,
		Content:     fromReplyItems(req.Content),
	})
	if errCode != 0 {
		handler.ErrorLog(logger.RobotConfigLogger, "robotConfigSvc.ApplyReplyConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}
