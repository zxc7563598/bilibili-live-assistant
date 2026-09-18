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

// 本模块的路由是公开的 /auth/altcha/challenge，既不属于 /api/shop 也不属于
// /api/admin（商城端与管理端登录都要过这道验证码），因此本包没有 shop.go /
// admin.go，接口方法直接放在 common.go 里。

// GetChallenge 获取 altcha 验证码挑战，未启用验证码或生成失败时返回空 JSON。
//
// 本接口不用 response.Success 包统一信封：altcha 前端组件解析的是它自己的
// 挑战报文格式，套上信封反而会让组件取不到字段。
//
// @Summary 获取 altcha 验证码挑战
// @Description 生成 altcha 工作量证明挑战，供前端验证码组件使用。未配置 hmacKey（即未启用验证码）或生成失败时返回空对象 {}
// @Tags 通用
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} object "altcha 挑战对象；未启用验证码或生成失败时为空对象 {}"
// @Router /auth/altcha/challenge [get]
func (h *Handler) GetChallenge(c *gin.Context) {
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
