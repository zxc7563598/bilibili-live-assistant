package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig CORS 中间件配置
type CORSConfig struct {
	// 允许的来源列表，为空时允许所有来源
	AllowedOrigins []string
}

// CORSMiddleware CORS 中间件
// 根据配置动态设置 Access-Control-Allow-Origin，避免 Allow-Origin: * 与 Allow-Credentials 冲突
func CORSMiddleware(cfg CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 判断是否允许该来源
		allowed := false
		if len(cfg.AllowedOrigins) == 0 {
			allowed = true
		} else {
			for _, o := range cfg.AllowedOrigins {
				if strings.EqualFold(o, origin) {
					allowed = true
					break
				}
			}
		}

		if allowed && origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// 允许的请求方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		// 允许的请求头
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Lang, Accept-Language")
		// 暴露的响应头
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		// 允许携带认证信息
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// 预检请求缓存时间（1小时）
		c.Writer.Header().Set("Access-Control-Max-Age", "3600")

		// 处理 OPTIONS 预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
