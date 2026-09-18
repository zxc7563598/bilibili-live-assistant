package appconfig

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/appconfig"
)

// @Summary 获取 App 配置
// @Description 获取后台可编辑的全部 App 配置（站点信息、颜色、图标、注册开关、登录页配置），用于回填管理页表单
// @Tags App配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.AppConfigDataResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/appconfig/data [post]
func (h *Handler) GetConfig(c *gin.Context) {
	lang := i18n.GetLang(c.Request.Context())
	svcResp, errCode, err := h.appConfigSvc.GetConfigData()
	if errCode != 0 {
		handler.ErrorLog(logger.AppConfigLogger, "appConfigSvc.GetConfigData 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, resp.AppConfigDataResp{
		SiteName:            svcResp.SiteName,
		SiteDescription:     svcResp.SiteDescription,
		SiteBackgroundColor: svcResp.SiteBackgroundColor,
		SiteThemeColor:      svcResp.SiteThemeColor,
		SiteIcon:            svcResp.SiteIcon,
		Register:            svcResp.Register,
		Logo:                svcResp.Logo,
		LoginBg:             svcResp.LoginBg,
		LoginTitle:          svcResp.LoginTitle,
		LoginSlogan:         svcResp.LoginSlogan,
		OssEndpoint:         svcResp.OssEndpoint,
		OssAccessKeyId:      svcResp.OssAccessKeyId,
		OssAccessKeySecret:  svcResp.OssAccessKeySecret,
		OssBucket:           svcResp.OssBucket,
	})
}

// @Summary 保存 App 基础配置
// @Description 整体覆盖保存基本 App 配置并刷新缓存立即生效，可为空表示清除对应配置
// @Tags App配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AppConfigSaveConfigReq true "App 配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/appconfig/save [post]
func (h *Handler) SaveConfig(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)

	var req input.AppConfigSaveConfigReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.AppConfigLogger, "SaveConfig 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.appConfigSvc.SaveConfig(ctx, appconfig.SaveConfigReq{
		SiteName:            req.SiteName,
		SiteDescription:     req.SiteDescription,
		SiteBackgroundColor: req.SiteBackgroundColor,
		SiteThemeColor:      req.SiteThemeColor,
		SiteIcon:            req.SiteIcon,
		Register:            req.Register,
		Logo:                req.Logo,
		LoginBg:             req.LoginBg,
		LoginTitle:          req.LoginTitle,
		LoginSlogan:         req.LoginSlogan,
	})
	if errCode != 0 {
		handler.ErrorLog(logger.AppConfigLogger, "appConfigSvc.SaveConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}

// @Summary 保存 OSS 配置
// @Description 整体覆盖保存 OSS 相关配置并刷新缓存立即生效，可为空表示清除对应配置
// @Tags App配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AppConfigSaveOssConfigReq true "App 配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/appconfig/oss_save [post]
func (h *Handler) SaveOssConfig(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)

	var req input.AppConfigSaveOssConfigReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.AppConfigLogger, "SaveOssConfig 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.appConfigSvc.SaveOssConfig(ctx, appconfig.SaveOssConfigReq{
		OssEndpoint:        req.OssEndpoint,
		OssAccessKeyId:     req.OssAccessKeyId,
		OssAccessKeySecret: req.OssAccessKeySecret,
		OssBucket:          req.OssBucket,
	})
	if errCode != 0 {
		handler.ErrorLog(logger.AppConfigLogger, "appConfigSvc.SaveConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, nil)
}
