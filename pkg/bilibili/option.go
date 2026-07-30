package bilibili

import (
	"io"
	"net/http"
	"net/url"
)

// Option 是 Client 的函数式配置项
type Option func(*Client)

// WithHTTPClient 设置自定义的 http.Client（如配置代理、TLS 等）
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithSession 注入用户身份信息，并预填充 Buvid Cookie
func WithSession(session *Session) Option {
	return func(c *Client) {
		c.session = session
	}
}

// WithCookies 预加载 Cookie 到 CookieJar 中
// 典型用途：从配置文件或数据库恢复已保存的登录态
func WithCookies(baseURL string, cookies []*http.Cookie) Option {
	return func(c *Client) {
		u, _ := url.Parse(baseURL)
		c.cookieJar.ImportCookies(u, cookies)
	}
}

// WithDebug 启用 HTTP 调试输出，将所有请求和响应的原始报文写入 w
//
// 输出包含完整的 URL、Header（含 CookieJar 自动注入的 Cookie）、Body
// 典型用途：调试 API 请求，查看实际发送了什么参数和 Cookie
//
//	client := bilibili.NewClient(
//	    bilibili.WithStateFile("bilibili_state.json"),
//	    bilibili.WithDebug(os.Stderr),
//	)
func WithDebug(w io.Writer) Option {
	return func(c *Client) {
		c.debugWriter = w
	}
}

// WithStateFile 从 JSON 文件加载 Cookie 和 Session 状态
// 文件通常由 Client.SaveState() 创建
//
// 如果文件不存在（首次运行），该选项静默跳过，不会报错
// 这样可以写死路径，首次启动手动登录后调用 SaveState，重启自动恢复
//
//	// 通用模式：写死状态文件路径，首次登录后自动保存
//	client := bilibili.NewClient(bilibili.WithStateFile("bilibili_state.json"))
//	if client.Session() == nil {
//	    // 未登录，走扫码登录流程
//	    qr, _ := client.Auth.GetQRCode(ctx)
//	    // ... 轮询扫码状态 ...
//	    client.SaveState("bilibili_state.json")
//	}
func WithStateFile(path string) Option {
	return func(c *Client) {
		_ = c.LoadState(path)
	}
}
