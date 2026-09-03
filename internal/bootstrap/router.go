package bootstrap

import (
	"io"
	"io/fs"
	"net/http"
	"net/http/pprof"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/middleware"
	"github.com/zxc7563598/bilibili-live-assistant/internal/webui"
)

func RouteRegister(r *gin.Engine, rdb *redis.Client, handlers *Handlers, corsCfg config.CORSConfig, cryptoCfg config.CryptoConfig) *gin.Engine {
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	// 日志注册
	if gin.Mode() != gin.ReleaseMode {
		registerApiDoc(r)
		registerPprof(r)
	}
	// 中间件注册
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORSMiddleware(middleware.CORSConfig{
		AllowedOrigins: corsCfg.AllowedOrigins,
	}), middleware.LocaleMiddleware())
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	// altcha 验证码（独立路由，不受分组中间件影响）
	r.GET("/auth/altcha/challenge", handlers.Altcha.Challenge)
	// web路由
	admin := r.Group("/admin")
	registerWeb(admin)
	// shop路由
	shop := r.Group("/shop")
	registerShop(shop)
	// shop api路由
	shopApi := r.Group("/api/shop")
	// 请求体解密中间件：验证/解密前端 encryptRequest 加密的请求体，明文请求按策略放行或拒绝
	shopApi.Use(middleware.ShopEncrypt(cryptoCfg.RequireEncryption))
	shopApi.GET("/manifest", handlers.AppConfig.GetManifest)
	shopApi.GET("/theme-color", handlers.AppConfig.GetThemeColor)
	shopApi.GET("/login", handlers.AppConfig.GetLoginConfig)
	shopApi.GET("/public-key", handlers.AppConfig.GetPublicKey)
	// 按账号(UID)固定窗口限流：防止对同一账号暴力撞库 / 频繁探测。
	// account 探测与 login 共用同一预算，避免交替请求绕过单接口上限。
	accountLoginLimiter := middleware.NewAccountRateLimiter(10, time.Minute)
	shopApi.POST("/liveuser/account", accountLoginLimiter, handlers.LiveUser.ExistsAccount)
	shopApi.POST("/liveuser/login", accountLoginLimiter, handlers.LiveUser.Login)
	shopApi.POST("/liveuser/refresh", handlers.LiveUser.Refresh)
	shopApi.POST("/liveuser/logout", middleware.UserAuth(rdb), handlers.LiveUser.Logout)
	shopApi.POST("/liveuser/info", middleware.UserAuth(rdb), handlers.LiveUser.GetUserInfo)
	shopApi.POST("/liveuser/room-id", middleware.UserAuth(rdb), handlers.LiveUser.GetRoomID)
	shopApi.POST("/product/list", middleware.UserAuth(rdb), handlers.Product.ShopListPage)
	shopApi.POST("/product/detail", middleware.UserAuth(rdb), handlers.Product.ShopDetail)
	// api路由
	adminApi := r.Group("/api/admin")
	// 登录接口：如需限流可参考 shop 端的按账号限流（middleware.NewAccountRateLimiter(10, time.Minute)）
	adminApi.POST("/auth/login", handlers.Admin.Login)
	adminApi.POST("/auth/captcha", handlers.Admin.CaptchaStatus)
	// 认证路由（所有登录用户可访问）
	adminApi.POST("/auth/switch-role", middleware.AdminAuth(rdb), handlers.Admin.SwitchRole)
	adminApi.POST("/auth/logout", middleware.AdminAuth(rdb), handlers.Admin.Logout)
	adminApi.POST("/auth/refresh", handlers.Admin.Refresh)
	adminApi.POST("/auth/detail", middleware.AdminAuth(rdb), handlers.Admin.Details)
	adminApi.POST("/auth/change-password", middleware.AdminAuth(rdb), handlers.Admin.ChangePassword)
	adminApi.POST("/admin/update-profile", middleware.AdminAuth(rdb), handlers.Admin.UpdateProfile)
	adminApi.POST("/roles/permissions", middleware.AdminAuth(rdb), handlers.Role.Permissions)
	adminApi.POST("/menu/list", middleware.AdminAuth(rdb), handlers.Menu.List)
	adminApi.POST("/menu/validate", middleware.AdminAuth(rdb), handlers.Menu.Validate)
	adminApi.POST("/menu/buttons", middleware.AdminAuth(rdb), handlers.Menu.Buttons)
	adminApi.POST("/roles/all", middleware.AdminAuth(rdb), handlers.Role.ListAll)
	// 管理员管理路由（仅超级管理员可访问）
	adminApi.POST("/admin/list", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Admin.ListPage)
	adminApi.POST("/admin/delete", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Admin.Delete)
	adminApi.POST("/admin/save", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Admin.Save)
	adminApi.POST("/admin/update-password", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Admin.UpdatePassword)
	// 角色管理路由（仅超级管理员可访问）
	adminApi.POST("/roles/list", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Role.ListPage)
	adminApi.POST("/roles/save", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Role.Save)
	adminApi.POST("/roles/delete", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Role.Delete)
	adminApi.POST("/roles/add-role-users", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Role.AddRoleUsers)
	adminApi.POST("/roles/remove-role-users", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Role.RemoveRoleUsers)
	// 菜单管理路由（仅超级管理员可访问）
	adminApi.POST("/menu/save", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Menu.Save)
	adminApi.POST("/menu/toggle", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Menu.Toggle)
	adminApi.POST("/menu/delete", middleware.AdminAuth(rdb), middleware.RequireRole("SUPER_ADMIN"), handlers.Menu.Delete)
	// 直播控制路由（所有登录用户可访问）
	adminApi.POST("/live/login/qrcode", middleware.AdminAuth(rdb), handlers.Live.GetQRCode)
	adminApi.POST("/live/login/poll", middleware.AdminAuth(rdb), handlers.Live.PollQRCode)
	adminApi.POST("/live/login/status", middleware.AdminAuth(rdb), handlers.Live.GetLoginStatus)
	adminApi.POST("/live/login/logout", middleware.AdminAuth(rdb), handlers.Live.Logout)
	adminApi.POST("/live/room/update", middleware.AdminAuth(rdb), handlers.Live.UpdateRoom)
	adminApi.POST("/live/room/send-danmu", middleware.AdminAuth(rdb), handlers.Live.SendDanmu)
	adminApi.POST("/live/listener/start", middleware.AdminAuth(rdb), handlers.Live.StartListener)
	adminApi.POST("/live/listener/stop", middleware.AdminAuth(rdb), handlers.Live.StopListener)
	adminApi.POST("/live/listener/status", middleware.AdminAuth(rdb), handlers.Live.GetListenerStatus)
	// 直播消息 WebSocket 推送
	// 认证在 Handler 内部通过 query param（?token=xxx）处理，
	// 因为浏览器 WebSocket API 不支持自定义请求头，无法使用 AdminAuth 中间件。
	// 因此该路由注册在 adminApi 上而非 live Group 内，绕过 AdminAuth
	adminApi.GET("/live/messages/stream", handlers.Live.MessageStream)
	// 机器人配置路由（所有登录用户可访问）
	adminApi.POST("/robot/room/get", middleware.AdminAuth(rdb), handlers.RobotConfig.GetRoom)
	adminApi.POST("/robot/room/apply", middleware.AdminAuth(rdb), handlers.RobotConfig.ApplyRoom)
	adminApi.POST("/robot/sign/get", middleware.AdminAuth(rdb), handlers.RobotConfig.GetSign)
	adminApi.POST("/robot/sign/apply", middleware.AdminAuth(rdb), handlers.RobotConfig.ApplySign)
	adminApi.POST("/robot/ad/get", middleware.AdminAuth(rdb), handlers.RobotConfig.GetAd)
	adminApi.POST("/robot/ad/apply", middleware.AdminAuth(rdb), handlers.RobotConfig.ApplyAd)
	adminApi.POST("/robot/gift/get", middleware.AdminAuth(rdb), handlers.RobotConfig.GetGift)
	adminApi.POST("/robot/gift/apply", middleware.AdminAuth(rdb), handlers.RobotConfig.ApplyGift)
	adminApi.POST("/robot/pk/get", middleware.AdminAuth(rdb), handlers.RobotConfig.GetPk)
	adminApi.POST("/robot/pk/apply", middleware.AdminAuth(rdb), handlers.RobotConfig.ApplyPk)
	adminApi.POST("/robot/welcome/get", middleware.AdminAuth(rdb), handlers.RobotConfig.GetWelcome)
	adminApi.POST("/robot/welcome/apply", middleware.AdminAuth(rdb), handlers.RobotConfig.ApplyWelcome)
	adminApi.POST("/robot/follow/get", middleware.AdminAuth(rdb), handlers.RobotConfig.GetFollow)
	adminApi.POST("/robot/follow/apply", middleware.AdminAuth(rdb), handlers.RobotConfig.ApplyFollow)
	adminApi.POST("/robot/share/get", middleware.AdminAuth(rdb), handlers.RobotConfig.GetShare)
	adminApi.POST("/robot/share/apply", middleware.AdminAuth(rdb), handlers.RobotConfig.ApplyShare)
	adminApi.POST("/robot/reply/get", middleware.AdminAuth(rdb), handlers.RobotConfig.GetReply)
	adminApi.POST("/robot/reply/apply", middleware.AdminAuth(rdb), handlers.RobotConfig.ApplyReply)
	// 弹幕列表路由
	adminApi.POST("/livedanmu/room", middleware.AdminAuth(rdb), handlers.LiveDanmu.FetchRoomGroups)
	adminApi.POST("/livedanmu/list", middleware.AdminAuth(rdb), handlers.LiveDanmu.ListPage)
	// 礼物列表路由
	adminApi.POST("/livegift/room", middleware.AdminAuth(rdb), handlers.LiveGift.FetchRoomGroups)
	adminApi.POST("/livegift/list", middleware.AdminAuth(rdb), handlers.LiveGift.ListPage)
	adminApi.POST("/livegift/blindbox", middleware.AdminAuth(rdb), handlers.LiveGift.BlindBoxListPage)
	// 用户列表路由
	adminApi.POST("/liveuser/list", middleware.AdminAuth(rdb), handlers.LiveUser.ListPage)
	adminApi.POST("/liveuser/monthly", middleware.AdminAuth(rdb), handlers.LiveUser.UserMonthlyAnalysis)
	adminApi.POST("/liveuser/danmu", middleware.AdminAuth(rdb), handlers.LiveUser.UserDanmuAnalysis)
	return r
}

// staticCacheControl 为嵌入的静态资源设置合理的缓存策略。
//
// 必须在 http.StripPrefix 之前、基于完整请求路径调用（strip 后的路径不带
// 前导斜杠，无法可靠识别 /assets/ 前缀）：
//
//   - sw.js / workbox-*.js / index.html 等入口与更新类文件返回 no-cache，
//     要求每次回源校验。若它们被 HTTP 层缓存，PWA 的 Service Worker 将无法
//     检测到新版本，老浏览器会一直展示旧内容（禁用浏览器缓存也绕不过 SW）；
//   - /assets/ 下的资源文件名带内容 hash、内容不可变，可交给浏览器/CDN 永久缓存。
func staticCacheControl(c *gin.Context) {
	p := c.Request.URL.Path
	switch name := path.Base(p); {
	case strings.HasSuffix(p, "/"), name == "sw.js", strings.HasPrefix(name, "workbox-"), name == "index.html":
		c.Header("Cache-Control", "no-cache")
	case strings.Contains(p, "/assets/"):
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	}
}

func registerWeb(route *gin.RouterGroup) {
	sub, err := fs.Sub(webui.Dist, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))
	route.GET("", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/admin/")
	})
	route.GET("/*filepath", func(c *gin.Context) {
		staticCacheControl(c)
		path := c.Param("filepath")
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		// index.html 不走 FileServer：它会对任何以 /index.html 结尾的路径 301 到
		// "./"，该重定向被 PWA Service Worker 预缓存阶段跟随后，缓存里会存下
		// redirected response，导航请求(redirect mode=manual)用到时浏览器报
		// "a redirected response was used for a request whose redirect mode is not follow"。
		// 显式请求 index.html 时直接内联 200 返回（与下方 SPA fallback 一致）。
		if path != "index.html" {
			if _, err := sub.Open(path); err == nil {
				http.StripPrefix("/admin/", fileServer).ServeHTTP(c.Writer, c.Request)
				return
			}
		}
		index, err := sub.Open("index.html")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		defer index.Close()
		c.Header("Cache-Control", "no-cache")
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)
		io.Copy(c.Writer, index)
	})
}

func registerShop(route *gin.RouterGroup) {
	sub, err := fs.Sub(webui.Shop, "shop")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))
	route.GET("", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/shop/")
	})
	route.GET("/*filepath", func(c *gin.Context) {
		staticCacheControl(c)
		path := c.Param("filepath")
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		// 见 registerWeb：显式请求 index.html 时绕过 FileServer 的 301 规范化，
		// 避免 PWA 预缓存得到 redirected response。
		if path != "index.html" {
			if _, err := sub.Open(path); err == nil {
				http.StripPrefix("/shop/", fileServer).ServeHTTP(c.Writer, c.Request)
				return
			}
		}
		index, err := sub.Open("index.html")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		defer index.Close()
		c.Header("Cache-Control", "no-cache")
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)
		io.Copy(c.Writer, index)
	})
}

// registerPprof 在开发模式下注册 net/http/pprof 性能分析接口
func registerPprof(r *gin.Engine) {
	r.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	r.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	r.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	r.POST("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	r.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	r.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
	for _, name := range []string{"allocs", "block", "goroutine", "heap", "mutex", "threadcreate"} {
		r.GET("/debug/pprof/"+name, gin.WrapF(pprof.Handler(name).ServeHTTP))
	}
}

func registerApiDoc(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/redoc", func(c *gin.Context) {
		html := `<!DOCTYPE html>
			<html>
				<head>
					<title>API Documentation - ReDoc</title>
					<meta charset="utf-8"/>
					<meta name="viewport" content="width=device-width, initial-scale=1">
					<style>
						body { margin: 0; padding: 0; }
					</style>
				</head>
				<body>
					<redoc spec-url='/swagger/doc.json'></redoc>
					<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
				</body>
			</html>`
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, html)
	})
}
