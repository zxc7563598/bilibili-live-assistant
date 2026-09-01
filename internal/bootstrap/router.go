package bootstrap

import (
	"io"
	"io/fs"
	"net/http"

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
	shopApi.GET("/public-key", handlers.AppConfig.GetPublicKey)
	// api路由
	adminApi := r.Group("/api/admin")
	// 登录接口：如有需要可以自行实现限流器
	// loginLimiter := middleware.NewRateLimiter(10, 1*time.Minute)
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
		path := c.Param("filepath")
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		if _, err := sub.Open(path); err == nil {
			http.StripPrefix("/admin/", fileServer).ServeHTTP(c.Writer, c.Request)
			return
		}
		index, err := sub.Open("index.html")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		defer index.Close()
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
		path := c.Param("filepath")
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		if _, err := sub.Open(path); err == nil {
			http.StripPrefix("/shop/", fileServer).ServeHTTP(c.Writer, c.Request)
			return
		}
		index, err := sub.Open("index.html")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		defer index.Close()
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)
		io.Copy(c.Writer, index)
	})
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
