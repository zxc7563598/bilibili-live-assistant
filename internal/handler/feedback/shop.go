package feedback

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/feedback"
	"go.uber.org/zap"
)

// @Summary 提交用户投诉/反馈
// @Description 当前登录用户提交一条投诉/反馈（问题类型、内容、联系方式），提交后由平台侧处理；联系方式仅用于平台回访
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.FeedbackSubmitReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/shop/feedback/submit [post]
func (h *Handler) Submit(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取请求参数
	var req input.FeedbackSubmitReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.FeedbackLogger,
			"Submit 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.feedbackSvc.Add(ctx, userInfo.UserID, feedback.CreateReq{
		Type:    req.Type,
		Content: req.Content,
		Contact: req.Contact,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.FeedbackLogger,
			"feedbackSvc.Add 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
			zap.String("req.type", req.Type),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}
