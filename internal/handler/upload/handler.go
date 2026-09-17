package upload

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	uploadSvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/upload"
)

// Handler 图片上传 / OSS 同步 HTTP 接口处理器
type Handler struct {
	uploadSvc *uploadSvc.Service
}

// New 创建 Handler 实例
func New(uploadSvc *uploadSvc.Service) *Handler {
	return &Handler{
		uploadSvc: uploadSvc,
	}
}

// @Summary 上传图片
// @Description 接收图片文件，按 scene 白名单落盘到 uploads/ 目录，返回可直接访问的图片路径
// @Tags 上传
// @Security BearerAuth
// @Accept multipart/form-data
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param scene formData string true "图片用途（决定落盘子目录）" enums(login_bg,site_icon,logo,cover,carousel,details)
// @Param file formData file true "图片文件"
// @Success 200 {object} response.Response{data=resp.UploadPathResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/upload/image [post]
func (h *Handler) UploadImage(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 请求体限长（单文件上限 + multipart 开销），必须在 MultipartForm 之前设置，
	// 否则超限内容会被读进内存直到 OOM
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, uploadSvc.MaxRequestSize)
	if _, formErr := c.MultipartForm(); formErr != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(formErr, &maxBytesErr) {
			handler.ErrorLog(logger.UploadLogger, "UploadImage 上传文件超过大小上限", 11404, formErr)
			response.Error(c, lang, 11404)
			return
		}
		handler.ErrorLog(logger.UploadLogger, "UploadImage multipart 表单解析失败", 11401, formErr)
		response.Error(c, lang, 11401)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		// MultipartForm 已解析成功，走到这里只可能是没带 file 字段
		handler.ErrorLog(logger.UploadLogger, "UploadImage 未接收到上传文件", 11403, err)
		response.Error(c, lang, 11403)
		return
	}
	svcResp, errCode, err := h.uploadSvc.UploadImage(ctx, uploadSvc.UploadImageReq{
		Scene: c.PostForm("scene"),
		File:  file,
	})
	if errCode != 0 {
		handler.ErrorLog(logger.UploadLogger, "uploadSvc.UploadImage 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, resp.UploadPathResp{Path: svcResp.Path})
}

// @Summary 同步图片到阿里云OSS
// @Description 接收图片路径，同步到阿里云OSS，返回可直接访问的图片路径
// @Tags 上传
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.UploadSyncOSSReq true "图片路径参数"
// @Success 200 {object} response.Response{data=resp.UploadPathResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/upload/oss-sync [post]
func (h *Handler) SyncOSS(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	var req input.UploadSyncOSSReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.UploadLogger, "SyncOSS 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	svcResp, errCode, err := h.uploadSvc.SyncOSS(ctx, req.Path)
	if errCode != 0 {
		handler.ErrorLog(logger.UploadLogger, "uploadSvc.SyncOSS 调用失败", errCode, err)
		response.Error(c, lang, errCode)
		return
	}
	response.Success(c, lang, resp.UploadPathResp{Path: svcResp.Path})
}
