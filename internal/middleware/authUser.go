package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/jwt"
)

func UserAuth(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, "", 10004)
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, "", 10005)
			c.Abort()
			return
		}
		tokenString := parts[1]
		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			response.Error(c, "", 10002)
			c.Abort()
			return
		}
		if claims.Type != "access" {
			response.Error(c, "", 10003)
			c.Abort()
			return
		}
		// Redis 单点登录校验
		// 如果未启动 Redis 则不进行单点校正，用户会在 AccessToken 过期后，重新申请 RefreshToken 时才会被踢出
		if rdb != nil {
			ctx := c.Request.Context()
			key := jwt.UserTokenKey(claims.ID)
			redisToken, err := rdb.Get(ctx, key).Result()
			if err == redis.Nil {
				response.Error(c, "", 10006)
				c.Abort()
				return
			}
			if err != nil {
				response.Error(c, "", 10007)
				c.Abort()
				return
			}
			if redisToken != tokenString {
				response.Error(c, "", 10008)
				c.Abort()
				return
			}
		}
		// 写入上下文
		c.Set("userID", claims.ID)
		c.Next()
	}
}
