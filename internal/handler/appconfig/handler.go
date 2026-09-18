package appconfig

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/crypto"
)

// Handler App 配置 HTTP 接口处理器
type Handler struct {
	appConfigSvc *appconfig.Service
}

// New 创建 Handler 实例
func New(appConfigSvc *appconfig.Service) *Handler {
	return &Handler{
		appConfigSvc: appConfigSvc,
	}
}

// @Summary 获取 RSA 公钥（带 HMAC 验签）
// @Description 获取用于前端 RSA-OAEP 加密的 RSA 公钥（SPKI DER 的 base64），并附带 HMAC 签名供前端验签
// @Tags 移动端
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.AppPublicKeyResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/public-key [get]
func (h *Handler) GetPublicKey(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	pubKeyB64, err := crypto.EnsureRSAKeyPair()
	if err != nil {
		response.Error(c, lang, 60001)
		return
	}
	keyID := crypto.PublicKeyID(pubKeyB64)
	ts := time.Now().Unix()
	// 签名消息格式必须与前端一致：pubkey:<key_id><public_key><timestamp>
	msg := "pubkey:" + keyID + pubKeyB64 + strconv.FormatInt(ts, 10)
	response.Success(c, lang, resp.AppPublicKeyResp{
		KeyID:     keyID,
		PublicKey: pubKeyB64,
		Timestamp: ts,
		Sign:      crypto.HMACSHA256(msg, crypto.SignSecret),
	})
}

// @Summary 获取 App 的 Manifest 信息
// @Description 获取 App 的 Manifest 信息，用于前端构建 PWA 应用
// @Tags 移动端
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.AppShopManifestResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/manifest [get]
func (h *Handler) GetManifest(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.appConfigSvc.GetManifest()
	if errCode != 0 {
		handler.ErrorLog(logger.AppConfigLogger, "appConfigSvc.GetManifest 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, resp.AppShopManifestResp{
		Name:            svcResp.Name,
		ShortName:       svcResp.Name,
		Description:     svcResp.Description,
		ThemeColor:      svcResp.BackgroundColor,
		BackgroundColor: svcResp.BackgroundColor,
		Favicon:         svcResp.Icon,
		AppleTouchIcon:  svcResp.Icon,
		StartURL:        "/shop/",
		Scope:           "/shop/",
		Display:         "standalone",
		Icons: []resp.AppShopManifestIcon{
			resp.AppShopManifestIcon{
				Src:     svcResp.Icon,
				Sizes:   "512x512",
				Type:    svcResp.IconType,
				Purpose: "any",
			},
		},
	})
}

// @Summary 获取 App 主题色
// @Description 获取 App 主题色，用于前端构建样式
// @Tags 移动端
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.AppShopThemeColorResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/theme-color [get]
func (h *Handler) GetThemeColor(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.appConfigSvc.GetThemeColor()
	if errCode != 0 {
		handler.ErrorLog(logger.AppConfigLogger, "appConfigSvc.GetThemeColor 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, resp.AppShopThemeColorResp{
		Color: svcResp,
	})
}

// @Summary 获取登录页面配置信息
// @Description 获取登录页配置（注册开关、Logo、背景图、标题、Slogan）
// @Tags 移动端
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.AppShopLoginConfigResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/login [get]
func (h *Handler) GetLoginConfig(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	svcResp, errCode, err := h.appConfigSvc.GetLoginConfig()
	if errCode != 0 {
		handler.ErrorLog(logger.AppConfigLogger, "appConfigSvc.GetLoginConfig 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, resp.AppShopLoginConfigResp{
		Register: svcResp.Register,
		Logo:     svcResp.Logo,
		LoginBg:  svcResp.LoginBg,
		Title:    svcResp.Title,
		Slogan:   svcResp.Slogan,
	})
}

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
		handler.ErrorLog(logger.AppConfigLogger, "SaveConfig 参数异常", code, err)
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
