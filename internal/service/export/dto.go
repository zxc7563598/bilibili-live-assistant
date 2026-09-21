// Package export 提供通用的数据导出能力：短时效下载凭证 + CSV 流式写出。
//
// 与业务解耦：本包不 import 任何业务模块，各模块实现 Source 接口后由 bootstrap 注册进来。
// 全链路为同步流式 —— 分块查询业务表，边查边把 CSV 字节写给 HTTP 响应，
// 服务器不落盘、不缓存全量数据，因此百万行量级下内存恒定。
//
// 详细设计取舍见 internal/service/export/CLAUDE.md。
package export

import (
	"encoding/json"
	"time"
)

// Column 导出的列声明。
type Column struct {
	// Key 列标识，与前端表格列的 key 对应；仅用于索引本模块的字面量白名单，绝不进 SQL
	Key string
	// TitleKey i18n 文案键，表头文案由此取得（不接受前端传入的表头文本）
	TitleKey string
}

// Labeler 由需要按语言渲染文案的类型实现，internal/enum 下的枚举天然满足。
type Labeler interface {
	Text(lang string) string
}

// UnixTime 秒级时间戳，导出时格式化为 yyyy-MM-dd HH:mm:ss，0 输出空串。
type UnixTime int64

// Money 以「分」为单位的金额，导出时输出两位小数的元。
type Money int64

// CreateTicketReq 获取凭证入参。
type CreateTicketReq struct {
	// Module 导出模块名，须已在注册表中
	Module string
	// Columns 期望导出的列 key，顺序即 CSV 列顺序
	Columns []string
	// Filters 前端筛选条件的原始 JSON，由各模块自行解析与校验
	Filters json.RawMessage
}

// CreateTicketResp 获取凭证出参。
type CreateTicketResp struct {
	// URL 下载地址（含凭证），前端据此触发浏览器原生下载
	URL string
	// Filename 建议的文件名（与下载响应的 Content-Disposition 一致）
	Filename string
	// Total 命中行数，供前端提示
	Total int64
	// ExpiresAt 凭证过期时间
	ExpiresAt time.Time
}

// Config 导出服务的运行参数，来自配置文件的 export 段。
type Config struct {
	// Secret 凭证签名密钥（取 jwt.secret）
	Secret string
	// MaxRows 单次导出行数上限
	MaxRows int
	// MaxConcurrent 同时进行的导出数上限
	MaxConcurrent int
	// BatchSize 分块扫描的每块行数
	BatchSize int
	// TicketTTL 凭证有效期
	TicketTTL time.Duration
	// MaxDuration 单次导出墙钟上限
	MaxDuration time.Duration
	// WriteIdleTimeout 客户端写空闲上限，每块续期
	WriteIdleTimeout time.Duration
}

// Sink 导出数据的落地目标。
type Sink interface {
	// Start 写 BOM 与表头行
	Start(disposition string, titles []string) error
	// Row 写一条数据行
	Row(rec []string) error
	// Flush 把缓冲刷给客户端
	Flush() error
	// Marker 写一条中断标记行
	Marker(text string)
}
