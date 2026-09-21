package export

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
)

// Service 导出流程编排。
type Service struct {
	cfg      Config
	registry *Registry
	signer   *signer
	limiter  *limiter
}

// New 构建导出服务。sources 为各业务模块实现的导出数据源。
func New(cfg Config, sources ...Source) *Service {
	return &Service{
		cfg:      cfg,
		registry: NewRegistry(sources...),
		signer:   newSigner(cfg.Secret),
		limiter:  newLimiter(cfg.MaxConcurrent),
	}
}

// Modules 已注册的导出模块名
func (s *Service) Modules() []string {
	return s.registry.Modules()
}

// WriteIdleTimeout 客户端写空闲上限。handler 拿它做写超时的滚动续期，避免为此在 handler 层再注入一份配置。
func (s *Service) WriteIdleTimeout() time.Duration {
	return s.cfg.WriteIdleTimeout
}

// CreateTicket 获取凭证
func (s *Service) CreateTicket(ctx context.Context, req CreateTicketReq, adminID int64, lang string) (CreateTicketResp, int, error) {
	src, ok := s.registry.Get(req.Module)
	if !ok {
		return CreateTicketResp{}, CodeModuleNotFound, fmt.Errorf("未注册的导出模块: %q", req.Module)
	}
	cols, code, err := resolveColumns(src, req.Columns)
	if code != 0 {
		return CreateTicketResp{}, code, err
	}
	filters, code, err := src.Normalize(req.Filters)
	if code != 0 {
		return CreateTicketResp{}, code, err
	}
	// 并发预检
	if !s.limiter.hasFree() {
		return CreateTicketResp{}, CodeExportBusy, nil
	}
	// 有界计数：limit 传上限+1，命中超限即可提前停止，避免在大表上做全量 COUNT
	total, err := src.Count(ctx, filters, s.cfg.MaxRows+1)
	if err != nil {
		return CreateTicketResp{}, CodeExportFailed, fmt.Errorf("统计导出行数失败: %w", err)
	}
	if total > int64(s.cfg.MaxRows) {
		return CreateTicketResp{}, CodeRowLimit, nil
	}
	if total == 0 {
		return CreateTicketResp{}, CodeRowEmpty, nil
	}
	expiresAt := time.Now().Add(s.cfg.TicketTTL)
	keys := make([]string, len(cols))
	for i, c := range cols {
		keys[i] = c.Key
	}
	token, err := s.signer.sign(ticketPayload{
		Typ:  ticketType,
		Sub:  adminID,
		Mod:  req.Module,
		Cols: keys,
		Filt: filters,
		Lang: lang,
		Exp:  expiresAt.Unix(),
	})
	if err != nil {
		return CreateTicketResp{}, CodeExportFailed, fmt.Errorf("签发导出凭证失败: %w", err)
	}
	filename := filenameFor(req.Module, lang)
	return CreateTicketResp{
		URL:       "/api/admin/export/download?ticket=" + url.QueryEscape(token),
		Filename:  filename,
		Total:     total,
		ExpiresAt: expiresAt,
	}, 0, nil
}

// StreamResult 一次导出的结果。
type StreamResult struct {
	// Written 已写出的数据行数
	Written int64
	// Started 是否已经开始写出响应体。为 false 时调用方还能改用错误状态码表达失败
	Started bool
}

// Stream 校验凭证并流式导出
func (s *Service) Stream(ctx context.Context, token string, sink Sink) (StreamResult, int, error) {
	payload, err := s.signer.verify(token)
	if err != nil {
		return StreamResult{}, CodeTicketInvalid, err
	}
	src, ok := s.registry.Get(payload.Mod)
	if !ok {
		return StreamResult{}, CodeModuleNotFound, fmt.Errorf("未注册的导出模块: %q", payload.Mod)
	}
	cols, code, err := resolveColumns(src, payload.Cols)
	if code != 0 {
		return StreamResult{}, code, err
	}
	if !s.limiter.tryAcquire() {
		return StreamResult{}, CodeExportBusy, nil
	}
	defer s.limiter.release()
	keys := make([]string, len(cols))
	titles := make([]string, len(cols))
	for i, c := range cols {
		keys[i] = c.Key
		titles[i] = i18n.T(payload.Lang, c.TitleKey)
	}
	return s.stream(ctx, src, payload, keys, titles, sink)
}

// stream 分块拉取并写出的主循环
func (s *Service) stream(ctx context.Context, src Source, payload ticketPayload, keys, titles []string, sink Sink) (StreamResult, int, error) {
	var res StreamResult
	deadline := time.Now().Add(s.cfg.MaxDuration)
	batch := s.cfg.BatchSize
	var afterID int64
	// 未支持类型的汇总：逐格刷日志会淹掉文件，只记前几个并去重
	var unknown []string
	unknownSeen := map[string]bool{}
	disposition := buildDisposition(payload.Mod, payload.Lang)
	for {
		// 客户端断开时 Request.Context 会被取消，先把这一步挡掉，避免白查一块
		if err := ctx.Err(); err != nil {
			return res, 0, err
		}
		rows, nextID, err := src.FetchChunk(ctx, payload.Filt, afterID, keys, batch)
		if err != nil {
			// 已写出字节时无法再用错误状态码表达
			if res.Started {
				sink.Marker(i18n.E(payload.Lang, CodeExportFailed))
			}
			return res, CodeExportFailed, fmt.Errorf("读取导出数据失败: %w", err)
		}
		// 第一块成功取到之前不写任何字节：这样上面那些错误还能用正常状态码返回
		if !res.Started {
			if err := sink.Start(disposition, titles); err != nil {
				return res, CodeExportFailed, err
			}
			res.Started = true
		}
		// 行数上限：获取凭证时已按总量判定过，这里是统计之后又有并发写入的兜底。
		truncated := int64(len(rows)) > int64(s.cfg.MaxRows)-res.Written
		if truncated {
			rows = rows[:int64(s.cfg.MaxRows)-res.Written]
		}
		for _, row := range rows {
			if len(row) != len(keys) {
				if res.Started {
					sink.Marker(i18n.E(payload.Lang, CodeExportFailed))
				}
				return res, CodeExportFailed, fmt.Errorf("导出列数与数据列数不一致: %d != %d", len(row), len(keys))
			}
			rec := make([]string, len(row))
			for i, v := range row {
				cell, ok := formatCell(v, payload.Lang)
				if !ok && len(unknown) < 5 {
					sig := fmt.Sprintf("%s(%T)", keys[i], v)
					if !unknownSeen[sig] {
						unknownSeen[sig] = true
						unknown = append(unknown, sig)
					}
				}
				rec[i] = cell
			}
			if err := sink.Row(rec); err != nil {
				// 写失败通常就是客户端断开，不再追加标记行（对面已经没人了）
				return res, 0, err
			}
			res.Written++
		}
		if err := sink.Flush(); err != nil {
			return res, 0, err
		}
		if len(unknown) > 0 {
			sink.Marker(fmt.Sprintf("#存在未支持的数据类型，已按字符串输出: %s", strings.Join(unknown, ", ")))
			unknown = unknown[:0]
		}
		// 截断了才提示：这里是真的丢了行，必须让用户知道这份文件不完整
		if truncated {
			sink.Marker(i18n.E(payload.Lang, CodeRowLimit))
			return res, 0, nil
		}
		if len(rows) < batch {
			return res, 0, nil // 已取完
		}
		afterID = nextID
		if time.Now().After(deadline) {
			sink.Marker(i18n.E(payload.Lang, CodeExportFailed))
			return res, 0, nil
		}
	}
}

// buildDisposition 生成 Content-Disposition 取值
func buildDisposition(module, lang string) string {
	stamp := time.Now().Format("20060102_1504")
	asciiName := fmt.Sprintf("%s_%s.csv", module, stamp)
	return fmt.Sprintf(
		`attachment; filename="%s"; filename*=UTF-8''%s`,
		asciiName, url.PathEscape(filenameFor(module, lang)),
	)
}

// filenameFor 按语言取模块文案拼出的文件名
func filenameFor(module, lang string) string {
	label := i18n.T(lang, "export.module."+module)
	// 文案缺失时 T 会原样返回 key，此时退回模块名
	if label == "" || strings.HasPrefix(label, "export.module.") {
		label = module
	}
	return fmt.Sprintf("%s_%s.csv", label, time.Now().Format("20060102_1504"))
}
