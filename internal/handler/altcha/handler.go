package altcha

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/altcha"
	"go.uber.org/zap"
)

// Handler 用于处理 altcha 验证码相关 HTTP 请求
type Handler struct {
	altchaSvc *altcha.Service
}

// New 返回一个新的 Altcha Handler 实例
func New(altchaSvc *altcha.Service) *Handler {
	return &Handler{altchaSvc: altchaSvc}
}

// Challenge 获取 altcha 验证码挑战
// 若未配置 hmacKey 或创建失败则返回空 JSON
func (h *Handler) Challenge(c *gin.Context) {
	ctx := c.Request.Context()
	challenge, errCode, err := h.altchaSvc.CreateChallenge(ctx)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AltchaLogger,
			"altchaSvc.CreateChallenge 调用失败",
			errCode,
			err,
			zap.Int("errCode", errCode),
		)
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	if challenge == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	c.JSON(http.StatusOK, challenge)
}
