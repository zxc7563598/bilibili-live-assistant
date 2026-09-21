package export

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
	"go.uber.org/zap"
)

// @Summary 领取数据导出下载凭证
// @Description 按当前筛选条件校验导出请求（模块、列、行数上限、并发数），通过后返回一次性下载地址。
// @Description 所有校验都在这步完成：后续的下载接口走浏览器原生下载，失败信息无法展示给用户。
// @Tags 数据导出
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.ExportTicketReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.ExportTicketResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/export/ticket [post]
func (h *Handler) CreateTicket(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取请求参数
	var req input.ExportTicketReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.ExportLogger,
			"CreateTicket 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.exportSvc.CreateTicket(ctx, export.CreateTicketReq{
		Module:  req.Module,
		Columns: req.Columns,
		Filters: req.Filters,
	}, adminInfo.AdminID, lang)
	if errCode != 0 {
		handler.ErrorLog(
			logger.ExportLogger,
			"exportSvc.CreateTicket 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.String("req.module", req.Module),
			zap.Strings("req.columns", req.Columns),
			zap.ByteString("req.filters", req.Filters),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.ExportTicketResp{
		URL:       svcResp.URL,
		Filename:  svcResp.Filename,
		Total:     svcResp.Total,
		ExpiresAt: svcResp.ExpiresAt.Unix(),
	})
}

// @Summary 下载导出的 CSV
// @Description 凭凭证下载。浏览器原生下载无法携带 Authorization 头，因此凭证即鉴权，
// @Description 与 /api/admin/live/messages/stream 用 query 传 token 是同一类做法。
// @Description 响应为流式 CSV，不是统一 JSON 信封；失败时返回纯文本与对应状态码。
// @Tags 数据导出
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param ticket query string true "下载凭证"
// @Success 200 {file} file "CSV 文件流"
// @Router /api/admin/export/download [get]
func (h *Handler) Download(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	token := c.Query("ticket")
	if token == "" {
		writeStreamError(c, lang, export.CodeTicketInvalid)
		return
	}
	// 写超时治理：全局 WriteTimeout 是「读完请求头后 +15s」的绝对期限，而一次导出可能持续几分钟。
	// 这里改成滚动续期（每写一块续一次），既不会被切断，也不至于让卡死的客户端无限期占住连接。
	//
	// 只在本次请求上生效，不动全局配置 —— 其它接口的 15 秒保护必须保留。
	rc := http.NewResponseController(c.Writer)
	setDeadline := func() error {
		err := rc.SetWriteDeadline(time.Now().Add(h.exportSvc.WriteIdleTimeout()))
		// 底层不支持设置写期限时不致命：退回全局 15 秒，导出会被截断但不影响其它接口
		if err != nil && !errors.Is(err, errors.ErrUnsupported) {
			return err
		}
		return nil
	}
	// 这里只设期限、绝不 Flush：gin 的 Flush 会立即把 200 状态码写出去，
	// 那样后面所有失败分支都改不了状态码，只能以 200 返回一段错误文本。
	if err := setDeadline(); err != nil {
		handler.ErrorLog(logger.ExportLogger, "设置写超时失败", 0, err)
	}
	// 真正的刷新只发生在已经开始写响应体之后（由 Sink 调用）
	flush := func() error {
		if err := setDeadline(); err != nil {
			return err
		}
		c.Writer.Flush()
		return nil
	}

	sink := &httpSink{c: c, inner: export.NewSink(c.Writer, flush)}
	res, errCode, err := h.exportSvc.Stream(ctx, token, sink)
	if errCode == 0 && err == nil {
		logger.ExportLogger.Info("导出完成",
			zap.Int64("rows", res.Written),
		)
		return
	}
	// 已经写出过字节：状态码改不了，Stream 内部已往 CSV 里补了中断标记行，这里只需记日志
	if res.Started {
		handler.ErrorLog(
			logger.ExportLogger,
			"导出中途失败（响应已开始，已追加中断标记）",
			errCode,
			err,
			zap.Int64("rows", res.Written),
		)
		return
	}
	// 尚未写出任何字节：正常报错
	handler.ErrorLog(logger.ExportLogger, "exportSvc.Stream 调用失败", errCode, err)
	writeStreamError(c, lang, errCode)
}

// httpSink 在通用 CSV 写出器外面补上 HTTP 响应头。
//
// 表头必须在「第一块数据成功取到之后」才写：在那之前任何失败都还能用错误状态码表达，
// 一旦写了响应头就只能中断连接了。
type httpSink struct {
	c     *gin.Context
	inner export.Sink
}

func (s *httpSink) Start(disposition string, titles []string) error {
	s.c.Header("Content-Type", "text/csv; charset=utf-8")
	s.c.Header("Content-Disposition", disposition)
	s.c.Header("Cache-Control", "no-store")
	s.c.Header("X-Content-Type-Options", "nosniff")
	return s.inner.Start(disposition, titles)
}

func (s *httpSink) Row(rec []string) error { return s.inner.Row(rec) }

func (s *httpSink) Flush() error { return s.inner.Flush() }

func (s *httpSink) Marker(text string) { s.inner.Marker(text) }
