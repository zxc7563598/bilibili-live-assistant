package export

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// bom UTF-8 BOM。缺了它 Windows 版 Excel 双击打开会把中文显示成乱码。
var bom = []byte{0xEF, 0xBB, 0xBF}

// csvBufferSize CSV 输出缓冲大小，约 32KB
const csvBufferSize = 32 * 1024

// csvSink 把导出行写成 CSV
type csvSink struct {
	bw    *bufio.Writer
	cw    *csv.Writer
	flush func() error
	// started 确保 Start 之后才写数据行
	started bool
}

// NewSink 构建 CSV 落地器
func NewSink(w io.Writer, flush func() error) Sink {
	return newCSVSink(w, flush)
}

// newCSVSink 构建 CSV 落地器
func newCSVSink(w io.Writer, flush func() error) *csvSink {
	bw := bufio.NewWriterSize(w, csvBufferSize)
	cw := csv.NewWriter(bw)
	// Excel 对 CRLF 的兼容性最好，其它消费方也都能接受
	cw.UseCRLF = true
	return &csvSink{bw: bw, cw: cw, flush: flush}
}

// Start 写 BOM 与表头行
func (s *csvSink) Start(_ string, titles []string) error {
	if _, err := s.bw.Write(bom); err != nil {
		return err
	}
	if err := s.cw.Write(titles); err != nil {
		return err
	}
	s.started = true
	return s.Flush()
}

// Row 写一条数据行
func (s *csvSink) Row(rec []string) error {
	if !s.started {
		return fmt.Errorf("导出未开始就写入数据行")
	}
	return s.cw.Write(rec)
}

// Flush 把缓冲刷给客户端
func (s *csvSink) Flush() error {
	s.cw.Flush()
	if err := s.cw.Error(); err != nil {
		return err
	}
	if err := s.bw.Flush(); err != nil {
		return err
	}
	if s.flush != nil {
		return s.flush()
	}
	return nil
}

// Marker 写一条中断标记行
func (s *csvSink) Marker(text string) {
	// 错误只能忽略：客户端可能已经断开，写失败正是中断的原因
	_ = s.cw.Write([]string{text})
	_ = s.Flush()
}

// formatCell 把一个单元格的原始值渲染成 CSV 文本
func formatCell(v any, lang string) (string, bool) {
	switch t := v.(type) {
	case nil:
		return "", true
	case string:
		return guardInjection(t), true
	case Labeler:
		return guardInjection(t.Text(lang)), true
	case bool:
		if t {
			return i18n.T(lang, "yes"), true
		}
		return i18n.T(lang, "no"), true
	case UnixTime:
		return timeutil.Format(int64(t)), true
	case Money:
		return strconv.FormatFloat(float64(t)/100, 'f', 2, 64), true
	case time.Time:
		return t.Format(time.DateTime), true
	case int:
		return strconv.Itoa(t), true
	case int8:
		return strconv.FormatInt(int64(t), 10), true
	case int16:
		return strconv.FormatInt(int64(t), 10), true
	case int32:
		return strconv.FormatInt(int64(t), 10), true
	case int64:
		return strconv.FormatInt(t, 10), true
	case uint:
		return strconv.FormatUint(uint64(t), 10), true
	case uint8:
		return strconv.FormatUint(uint64(t), 10), true
	case uint16:
		return strconv.FormatUint(uint64(t), 10), true
	case uint32:
		return strconv.FormatUint(uint64(t), 10), true
	case uint64:
		return strconv.FormatUint(t, 10), true
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32), true
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), true
	default:
		return fmt.Sprintf("%v", v), false
	}
}

// guardInjection 防 CSV 公式注入
func guardInjection(s string) string {
	if s == "" {
		return s
	}
	// 制表符/回车同样是注入向量，一并视为需要前置单引号
	if strings.ContainsAny(s, "\t\r") {
		return "'" + s
	}
	switch s[0] {
	case '=', '+', '@':
		return "'" + s
	case '-':
		// 负数不能被破坏成文本，仅在「减号后面不是数字」时才转义
		if len(s) == 1 {
			return "'" + s
		}
		if _, err := strconv.ParseFloat(s, 64); err != nil {
			return "'" + s
		}
	}
	return s
}
