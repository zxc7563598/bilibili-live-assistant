package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/GoAdminKit/internal/i18n"
)

// LocaleMiddleware 多语言中间件
func LocaleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Query("lang")
		if lang == "" {
			lang = c.GetHeader("X-Lang")
		}
		if lang == "" {
			lang = c.GetHeader("Accept-Language")
		}
		if lang == "" {
			lang = "zh"
		}
		// zh-CN -> zh
		if strings.Contains(lang, "-") {
			lang = strings.Split(lang, "-")[0]
		}
		lang = strings.ToLower(lang)
		// 写入标准 context
		ctx := context.WithValue(c.Request.Context(), i18n.LangKey, lang)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
