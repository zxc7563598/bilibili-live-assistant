package imagetype

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// mimeByExt 常见 Web 图片扩展名 → MIME 类型映射。
// 取值与 http.DetectContentType 的嗅探结果保持一致（如 .ico → image/x-icon），
// 使扩展名快路径与内容检测两层结果不背离。
var mimeByExt = map[string]string{
	"png":  "image/png",
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"jfif": "image/jpeg",
	"gif":  "image/gif",
	"svg":  "image/svg+xml",
	"webp": "image/webp",
	"avif": "image/avif",
	"ico":  "image/x-icon",
	"bmp":  "image/bmp",
	"tif":  "image/tiff",
	"tiff": "image/tiff",
}

// stripQueryFragment 截掉 URL 的 query 与 fragment（取最早出现的 ? 或 #）。
func stripQueryFragment(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '?' || s[i] == '#' {
			return s[:i]
		}
	}
	return s
}

// MimeFromExtension 仅依据扩展名推断 MIME 类型，零 I/O。
// 命中返回 (mime, true)；无法识别返回 ("", false)。
func MimeFromExtension(pathOrURL string) (string, bool) {
	ext := strings.ToLower(path.Ext(stripQueryFragment(pathOrURL)))
	mime, ok := mimeByExt[strings.TrimPrefix(ext, ".")]
	return mime, ok
}

// Detector 图片类型检测器
type Detector struct {
	HTTPTimeout time.Duration
	UserAgent   string
	MaxFileSize int64

	cache *mimeCache // 检测结果缓存；nil 表示禁用
}

// Option 配置函数类型
type Option func(*Detector)

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(d *Detector) {
		d.HTTPTimeout = timeout
	}
}

// WithUserAgent 设置 User-Agent
func WithUserAgent(ua string) Option {
	return func(d *Detector) {
		d.UserAgent = ua
	}
}

// WithCache 启用检测结果缓存（maxEntries 条，ttl 过期）。
// 仅对无扩展名的 URL 检测生效；参数非正数时保持缓存禁用。
func WithCache(maxEntries int, ttl time.Duration) Option {
	return func(d *Detector) {
		if maxEntries > 0 && ttl > 0 {
			d.cache = newMimeCache(maxEntries, ttl)
		}
	}
}

// NewDetector 创建检测器（带默认配置）
func NewDetector(opts ...Option) *Detector {
	// 默认配置
	d := &Detector{
		HTTPTimeout: 10 * time.Second,
		UserAgent:   "ImageDetector/1.0",
		MaxFileSize: 512,
	}
	// 应用可选配置
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// GetMimeType 获取图片的 MIME 类型，支持本地文件路径或网络 URL。
// 检测顺序：扩展名快路径 → 结果缓存（仅 URL）→ 真实内容检测 → 写缓存。
// 注意：带已知图片扩展名的路径/URL 会跳过字节嗅探，直接按扩展名返回。
func (d *Detector) GetMimeType(pathOrURL string) (string, error) {
	if mimeType, ok := MimeFromExtension(pathOrURL); ok {
		return mimeType, nil
	}
	// 仅缓存 URL 检测结果：本地文件换盘有变旧风险，且内容嗅探本身很廉价
	if d.cache != nil && isURL(pathOrURL) {
		key := stripQueryFragment(pathOrURL)
		if mimeType, err, ok := d.cache.get(key); ok {
			return mimeType, err
		}
	}
	var (
		mimeType string
		err      error
	)
	if isURL(pathOrURL) {
		mimeType, err = d.getMimeFromURL(pathOrURL)
	} else {
		mimeType, err = d.getMimeFromLocalFile(pathOrURL)
	}
	if d.cache != nil && isURL(pathOrURL) {
		d.cache.set(stripQueryFragment(pathOrURL), mimeType, err)
	}
	return mimeType, err
}

// IsImage 判断是否为图片（基于 MIME 类型）
func (d *Detector) IsImage(pathOrURL string) (bool, error) {
	mimeType, err := d.GetMimeType(pathOrURL)
	if err != nil {
		return false, err
	}
	return strings.HasPrefix(mimeType, "image/"), nil
}

// defaultDetector 包级共享检测器（带缓存），对标 http.DefaultClient。
// 调用方如需要独立配置/隔离，请自行 NewDetector(WithCache(...))，不要重新赋值本变量。
var defaultDetector = NewDetector(WithCache(16, time.Hour))

// GetMimeTypeSimple 简易版本（复用包级共享检测器及其缓存）
func GetMimeTypeSimple(pathOrURL string) (string, error) {
	return defaultDetector.GetMimeType(pathOrURL)
}

// isURL 判断字符串是否为 URL
func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

// getMimeFromLocalFile 从本地文件获取 MIME 类型
func (d *Detector) getMimeFromLocalFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer file.Close()
	// 读取前 MaxFileSize 字节用于检测
	buffer := make([]byte, d.MaxFileSize)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	if n == 0 {
		return "", fmt.Errorf("文件为空")
	}
	// 使用 http.DetectContentType 检测
	mimeType := http.DetectContentType(buffer[:n])
	// 如果检测为通用二进制流，可能不是图片
	if mimeType == "application/octet-stream" {
		return "", fmt.Errorf("无法识别该文件类型，可能不是图片")
	}
	return mimeType, nil
}

// getMimeFromURL 从网络 URL 获取 MIME 类型
func (d *Detector) getMimeFromURL(url string) (string, error) {
	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: d.HTTPTimeout,
	}
	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", d.UserAgent)
	// 发起请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 URL 失败: %w", err)
	}
	defer resp.Body.Close()
	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP 请求失败，状态码: %d", resp.StatusCode)
	}
	// 读取前 MaxFileSize 字节
	buffer := make([]byte, d.MaxFileSize)
	n, err := resp.Body.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("读取响应体失败: %w", err)
	}
	if n == 0 {
		return "", fmt.Errorf("响应内容为空")
	}
	// 检测 MIME 类型
	mimeType := http.DetectContentType(buffer[:n])
	if mimeType == "application/octet-stream" {
		return "", fmt.Errorf("无法识别该 URL 内容类型，可能不是图片")
	}
	return mimeType, nil
}
