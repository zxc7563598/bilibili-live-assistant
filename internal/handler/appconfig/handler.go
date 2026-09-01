package appconfig

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/crypto"
)

// Handler 直播控制 HTTP 接口处理器
type Handler struct {
	appConfigSvc *appconfig.Service
	rdb          *redis.Client
}

// New 创建 Handler 实例
func New(appConfigSvc *appconfig.Service, rdb *redis.Client) *Handler {
	return &Handler{
		appConfigSvc: appConfigSvc,
		rdb:          rdb,
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
	svcResp, errCode, err := h.appConfigSvc.Manifest()
	if errCode != 0 {
		handler.ErrorLog(logger.AppConfigLogger, "appConfigSvc.Manifest 调用失败", errCode, err)
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
	svcResp, errCode, err := h.appConfigSvc.ThemeColor()
	if errCode != 0 {
		handler.ErrorLog(logger.AppConfigLogger, "appConfigSvc.ThemeColor 调用失败", errCode, err)
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
	svcResp, errCode, err := h.appConfigSvc.LoginConfig()
	if errCode != 0 {
		handler.ErrorLog(logger.AppConfigLogger, "appConfigSvc.LoginConfig 调用失败", errCode, err)
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
