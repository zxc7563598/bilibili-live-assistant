package version

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
)

// @Summary 获取版本信息
// @Description 返回当前运行版本与远端最新版本，供后台首页提示管理员更新；远端检查不通时只把 checkFailed 置为 true
// @Tags 版本
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.VersionInfoResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/version/info [post]
func (h *Handler) GetVersionInfo(c *gin.Context) {
	lang := i18n.GetLang(c.Request.Context())
	svcResp := h.versionSvc.GetVersionInfo()
	response.Success(c, lang, resp.VersionInfoResp{
		Version:     svcResp.Version,
		Commit:      svcResp.Commit,
		Latest:      svcResp.Latest,
		NeedUpdate:  svcResp.NeedUpdate,
		CheckFailed: svcResp.CheckFailed,
	})
}
