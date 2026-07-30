package bilibili

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/auth"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/room"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/user"
)

// Client 是 B站 API 客户端它持有 HTTP 传输层、CookieJar、用户身份信息，
// 以及所有业务子服务（Auth、Room、Live、User）
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	cookieJar  *CookieJar
	session    *Session

	// debug
	debugWriter io.Writer

	// 业务子服务
	Auth *auth.Service
	Room *room.Service
	User *user.Service
}

// NewClient 创建 B站 API 客户端
//
// 默认配置：
//   - 30 秒请求超时
//   - 自动 Cookie 管理（CookieJar）
//   - User-Agent: "Mozilla/5.0"
//
// 通过 Option 函数可覆盖默认值
func NewClient(options ...Option) *Client {
	cj := NewCookieJar()
	c := &Client{
		httpClient: &http.Client{
			Jar:     cj,
			Timeout: 30 * time.Second,
		},
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
		cookieJar: cj,
	}
	for _, o := range options {
		o(c)
	}
	// 如果设置了 debug writer，包装 Transport 以捕获完整 HTTP 报文
	// 放在 options 之后确保即使 WithHTTPClient 替换了 httpClient 也能生效
	if c.debugWriter != nil {
		transport := c.httpClient.Transport
		if transport == nil {
			transport = http.DefaultTransport
		}
		c.httpClient.Transport = &debugTransport{
			transport: transport,
			writer:    c.debugWriter,
		}
	}
	// 如果注入了 Session 且包含 Buvid，预填充 CookieJar
	if c.session != nil && c.session.Buvid != "" {
		u, _ := url.Parse("https://api.bilibili.com")
		c.cookieJar.SetCookies(u, []*http.Cookie{
			{Name: "buvid3", Value: c.session.Buvid},
		})
	}
	// 创建子服务*Client 由于实现了 Get/Post/PostForm/Do 方法，
	// 自动满足各子包的 HttpClient 接口
	c.Auth = auth.NewService(c)
	c.Room = room.NewService(c)
	c.User = user.NewService(c)
	return c
}

// Session 返回当前用户身份信息（只读）
func (c *Client) Session() *Session {
	return c.session
}

// SetSession 更新用户身份信息登录成功后调用，以便 SaveState 将 Session 一并持久化
func (c *Client) SetSession(s *Session) {
	c.session = s
}

// CookieJar 返回 CookieJar 实例，用于导出/导入 Cookie 以实现持久化
func (c *Client) CookieJar() *CookieJar {
	return c.cookieJar
}

// CSRF 返回 bili_jct（CSRF Token），B站 所有 POST 接口都需要该值
// CookieJar 中必须已包含 bili_jct（登录成功后由 Set-Cookie 自动写入）
// ImportEntries 会将 cookie 写入 .bilibili.com 根域，因此直接查根域即可
func (c *Client) CSRF() (string, error) {
	u, _ := url.Parse("https://bilibili.com")
	for _, ck := range c.cookieJar.Cookies(u) {
		if ck.Name == "bili_jct" {
			return ck.Value, nil
		}
	}
	return "", fmt.Errorf("bilibili: bili_jct not found in cookie jar, please login first")
}
