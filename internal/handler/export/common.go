// Package export 数据导出的 HTTP 入口。
//
// 本包只有两个接口：获取凭证（正常 JSON 信封）与下载（浏览器原生下载的流式 CSV）。
// 业务模块不在这里，各模块的导出数据源见 internal/service/<模块>/export.go。
package export

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
)

// Handler 导出接口
type Handler struct {
	exportSvc *export.Service
}

// New 构造函数
func New(exportSvc *export.Service) *Handler {
	return &Handler{exportSvc: exportSvc}
}

// writeStreamError 下载接口的失败返回。
//
// 不能复用 response.Error：走这里的是浏览器原生下载，响应体会被直接存成文件，
// JSON 信封在用户看来就是一个内容不对的 .csv。所以改用纯文本 + 明确的 HTTP 状态码，
// 并禁掉缓存与内容嗅探。
//
// 只在「还没写出任何字节」时调用；流一旦开始就改不了状态码了。
func writeStreamError(c *gin.Context, lang string, code int) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.String(statusForCode(code), i18n.E(lang, code))
}

// statusForCode 导出错误码 → HTTP 状态码。
//
// 正常路径不会走到这里（凭证几秒前才签发，各项校验都已通过），
// 这些状态码是给「凭证过期后浏览器重试」「手工改 URL」之类的情况用的。
func statusForCode(code int) int {
	switch code {
	case export.CodeTicketInvalid:
		return 401
	case export.CodeModuleNotFound:
		return 404
	case export.CodeColumnInvalid, export.CodeColumnEmpty, export.CodeFilterInvalid:
		return 400
	case export.CodeExportBusy:
		return 429
	case export.CodeRowLimit, export.CodeRowEmpty:
		return 409
	default:
		return 500
	}
}
