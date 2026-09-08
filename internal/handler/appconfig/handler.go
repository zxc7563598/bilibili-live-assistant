package appconfig

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/crypto"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/fileutil"
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

// uploadSceneAllow 图片上传场景白名单
var uploadSceneAllow = map[string]string{
	"login_bg":  "login_bg",  // 登录页背景图（建议 20:9）
	"site_icon": "site_icon", // 网站图标（建议 1:1，≤512x512）
	"logo":      "logo",      // 网站 logo（建议 1:1，≤512x512）
}

// @Summary 上传 App 配置图片
// @Description 接收登录页背景图 / 网站图标 / logo 图片，按 scene 白名单落盘到 uploads/ 目录，返回可直接访问的图片路径
// @Tags App配置
// @Security BearerAuth
// @Accept multipart/form-data
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param scene formData string true "图片用途（决定落盘子目录）" enums(login_bg,site_icon,logo)
// @Param file formData file true "图片文件"
// @Success 200 {object} response.Response{data=resp.AppUploadPathResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/appconfig/upload [post]
func (h *Handler) UploadImage(c *gin.Context) {
	lang := i18n.GetLang(c.Request.Context())
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, fileutil.MaxUploadSize+1<<20)
	if _, formErr := c.MultipartForm(); formErr != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(formErr, &maxBytesErr) {
			handler.ErrorLog(logger.AppConfigLogger, "UploadImage 上传文件超过大小上限", 10903, formErr)
			response.Error(c, lang, 10903)
			return
		}
		handler.ErrorLog(logger.AppConfigLogger, "UploadImage multipart 表单解析失败", 10901, formErr)
		response.Error(c, lang, 10901)
		return
	}
	scene := c.PostForm("scene")
	subDir, ok := uploadSceneAllow[scene]
	if !ok {
		handler.ErrorLog(logger.AppConfigLogger, "UploadImage 上传场景不合法", 10904, errors.New("未知场景: "+scene))
		response.Error(c, lang, 10904)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		// MultipartForm 已解析成功，走到这里只可能是没带 file 字段
		handler.ErrorLog(logger.AppConfigLogger, "UploadImage 未接收到上传文件", 10902, err)
		response.Error(c, lang, 10902)
		return
	}
	// 空文件直接拒绝
	if file.Size == 0 {
		handler.ErrorLog(logger.AppConfigLogger, "UploadImage 上传文件为空", 10902, nil)
		response.Error(c, lang, 10902)
		return
	}
	// 非图片直接拒绝
	if !sniffUploadImage(file) {
		handler.ErrorLog(logger.AppConfigLogger, "UploadImage 上传内容不是有效图片", 10901, nil)
		response.Error(c, lang, 10901)
		return
	}
	uploadPath, err := fileutil.SaveUploadedFile(file, subDir)
	switch {
	case err == nil:
		response.Success(c, lang, resp.AppUploadPathResp{Path: uploadPath})
	case errors.Is(err, fileutil.ErrEmpty):
		handler.ErrorLog(logger.AppConfigLogger, "UploadImage 上传文件为空", 10902, err)
		response.Error(c, lang, 10902)
	case errors.Is(err, fileutil.ErrTooLarge):
		handler.ErrorLog(logger.AppConfigLogger, "UploadImage 上传文件超过大小上限", 10903, err)
		response.Error(c, lang, 10903)
	case errors.Is(err, fileutil.ErrInvalidDir):
		// scene 白名单已保证子目录合法，走到这里属开发期传参问题
		handler.ErrorLog(logger.AppConfigLogger, "UploadImage 上传子目录不合法", 10904, err)
		response.Error(c, lang, 10904)
	default:
		handler.ErrorLog(logger.AppConfigLogger, "fileutil.SaveUploadedFile 调用失败", 60904, err)
		response.Error(c, lang, 60904)
	}
}

// @Summary 获取 App 全部配置
// @Description 获取后台可编辑的全部 App 配置（站点信息、颜色、图标、注册开关、登录页配置），用于回填管理页表单
// @Tags App配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.AppConfigDataResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/appconfig/data [post]
func (h *Handler) GetConfig(c *gin.Context) {
	lang := i18n.GetLang(c.Request.Context())
	svcResp, errCode, err := h.appConfigSvc.ConfigData()
	if errCode != 0 {
		handler.ErrorLog(logger.AppConfigLogger, "appConfigSvc.ConfigData 调用失败", errCode, err)
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
	})
}

// @Summary 保存 App 全部配置
// @Description 整体覆盖保存全部 App 配置并刷新缓存立即生效；入参与「获取配置」返回一一对应，可为空表示清除对应配置
// @Tags App配置
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AppConfigSaveReq true "App 配置参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/appconfig/save [post]
func (h *Handler) SaveConfig(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)

	var req input.AppConfigSaveReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.AppConfigLogger, "SaveConfig 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	errCode, err := h.appConfigSvc.SaveConfig(ctx, appconfig.ConfigData{
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

// sniffUploadImage 读取上传文件头部做内容嗅探，确认其真实类型为图片
func sniffUploadImage(file *multipart.FileHeader) bool {
	src, err := file.Open()
	if err != nil {
		return false
	}
	defer src.Close()
	buf := make([]byte, 512)
	n, _ := src.Read(buf)
	if n == 0 {
		return false
	}
	return strings.HasPrefix(http.DetectContentType(buf[:n]), "image/")
}
