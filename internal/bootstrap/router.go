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

func RouteRegister(r *gin.Engine, rdb *redis.Client, handlers *Handlers, corsCfg config.CORSConfig) *gin.Engine {
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
	live := adminApi.Group("/live").Use(middleware.AdminAuth(rdb))
	live.POST("/login/qrcode", handlers.Live.GetQRCode)
	live.POST("/login/poll", handlers.Live.PollQRCode)
	live.POST("/login/status", handlers.Live.GetLoginStatus)
	live.POST("/login/logout", handlers.Live.Logout)
	live.POST("/room/update", handlers.Live.UpdateRoom)
	live.POST("/room/send-danmu", handlers.Live.SendDanmu)
	live.POST("/listener/start", handlers.Live.StartListener)
	live.POST("/listener/stop", handlers.Live.StopListener)
	live.POST("/listener/status", handlers.Live.GetListenerStatus)
	return r
}

func registerWeb(admin *gin.RouterGroup) {
	sub, err := fs.Sub(webui.Dist, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))
	admin.GET("", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/admin/")
	})
	admin.GET("/*filepath", func(c *gin.Context) {
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
