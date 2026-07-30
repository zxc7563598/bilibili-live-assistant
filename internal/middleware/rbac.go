package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/GoAdminKit/internal/handler"
	"github.com/zxc7563598/GoAdminKit/internal/response"
)

// RequireRole 角色权限校验中间件
// allowedCodes: 允许访问的角色 code 列表（如 "SUPER_ADMIN"）
func RequireRole(allowedCodes ...string) gin.HandlerFunc {
	allowedSet := make(map[string]bool, len(allowedCodes))
	for _, c := range allowedCodes {
		allowedSet[c] = true
	}
	return func(c *gin.Context) {
		adminInfo, ok := handler.GetAdminInfo(c)
		if !ok {
			response.Error(c, "", 20001)
			c.Abort()
			return
		}
		if !allowedSet[adminInfo.RoleCode] {
			response.Error(c, "", 30001)
			c.Abort()
			return
		}
		c.Next()
	}
}
