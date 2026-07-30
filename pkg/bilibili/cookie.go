package bilibili

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// CookieJar 封装了标准库 cookiejar.Jar，提供 Cookie 持久化能力
// 它实现了 http.CookieJar 接口，可直接设置为 http.Client.Jar
type CookieJar struct {
	jar *cookiejar.Jar
}

// NewCookieJar 创建一个新的 CookieJar
func NewCookieJar() *CookieJar {
	jar, _ := cookiejar.New(nil)
	return &CookieJar{jar: jar}
}

// SetCookies 实现 http.CookieJar 接口
func (cj *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	cj.jar.SetCookies(u, cookies)
}

// Cookies 实现 http.CookieJar 接口
func (cj *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	return cj.jar.Cookies(u)
}

// ExportCookies 导出指定 URL 的所有 Cookie，可用于持久化到文件或数据库
func (cj *CookieJar) ExportCookies(u *url.URL) []*http.Cookie {
	return cj.jar.Cookies(u)
}

// ImportCookies 批量导入 Cookie，用于从持久化存储恢复
func (cj *CookieJar) ImportCookies(u *url.URL, cookies []*http.Cookie) {
	cj.jar.SetCookies(u, cookies)
}

// CookieString 将指定 URL 的所有 Cookie 导出为 HTTP Cookie Header 格式的字符串
// 格式: "name1=value1; name2=value2"
func (cj *CookieJar) CookieString(u *url.URL) string {
	cookies := cj.jar.Cookies(u)
	if len(cookies) == 0 {
		return ""
	}
	pairs := make([]string, len(cookies))
	for i, c := range cookies {
		pairs[i] = c.Name + "=" + c.Value
	}
	return strings.Join(pairs, "; ")
}

// SetCookieString 从原始 Cookie 字符串解析并导入 Cookie
// 典型用法：加载之前保存的 B站 Cookie 字符串（如 "SESSDATA=xxx; bili_jct=yyy"）
func (cj *CookieJar) SetCookieString(u *url.URL, raw string) {
	header := http.Header{}
	header.Add("Cookie", raw)
	req := http.Request{Header: header}
	cookies := req.Cookies()
	cj.jar.SetCookies(u, cookies)
}

// =========================================================================
// Cookie 序列化（JSON 可持久化）
// =========================================================================

// CookieEntry 是可序列化的 Cookie 条目，用于持久化到文件或数据库
type CookieEntry struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Expires  int64  `json:"expires,omitempty"` // Unix 时间戳，0 表示会话 Cookie
	Secure   bool   `json:"secure"`
	HttpOnly bool   `json:"http_only"`
}

// entryKey 为去重用的唯一键
func (e CookieEntry) entryKey() string {
	return e.Name + "\x00" + e.Domain + "\x00" + e.Path
}

// toHTTPCookie 将 CookieEntry 转为标准库 *http.Cookie
func (e CookieEntry) toHTTPCookie() *http.Cookie {
	c := &http.Cookie{
		Name:     e.Name,
		Value:    e.Value,
		Path:     e.Path,
		Domain:   e.Domain,
		Secure:   e.Secure,
		HttpOnly: e.HttpOnly,
	}
	if e.Expires > 0 {
		c.Expires = time.Unix(e.Expires, 0)
	}
	return c
}

// fromHTTPCookie 从标准库 *http.Cookie 创建 CookieEntry
func fromHTTPCookie(c *http.Cookie) CookieEntry {
	e := CookieEntry{
		Name:     c.Name,
		Value:    c.Value,
		Domain:   c.Domain,
		Path:     c.Path,
		Secure:   c.Secure,
		HttpOnly: c.HttpOnly,
	}
	if !c.Expires.IsZero() {
		e.Expires = c.Expires.Unix()
	}
	if e.Path == "" {
		e.Path = "/"
	}
	return e
}

// isBilibiliHost 判断 host 是否属于 bilibili.com 及其子域
func isBilibiliHost(host string) bool {
	return host == "bilibili.com" || strings.HasSuffix(host, ".bilibili.com")
}

// ExportAll 从多个 URL 导出所有 Cookie，按 (Name, Domain, Path) 去重
// 空 Domain 的 cookie 自动推导域名：bilibili 相关 host → ".bilibili.com"，否则用 host
func (cj *CookieJar) ExportAll(urls ...*url.URL) []CookieEntry {
	seen := make(map[string]bool)
	var entries []CookieEntry
	for _, u := range urls {
		for _, c := range cj.jar.Cookies(u) {
			e := fromHTTPCookie(c)
			if e.Domain == "" {
				if isBilibiliHost(u.Host) {
					e.Domain = ".bilibili.com"
				} else {
					e.Domain = u.Host
				}
			}
			key := e.entryKey()
			if seen[key] {
				continue
			}
			seen[key] = true
			entries = append(entries, e)
		}
	}
	return entries
}

// ImportEntries 批量导入序列化的 Cookie 条目到 CookieJar
// bilibili 域名的 cookie 会同时写入其自身 domain 和 .bilibili.com 根域，
// 确保 api.bilibili.com、api.live.bilibili.com 等所有子域请求都能携带
func (cj *CookieJar) ImportEntries(entries []CookieEntry) {
	for _, e := range entries {
		domain := e.Domain
		if domain == "" {
			domain = ".bilibili.com" // 修复旧格式空 Domain 数据
		}

		scheme := "https"
		if !e.Secure {
			scheme = "http"
		}

		// 写入自身 domain
		cj.setCookieEntry(e, domain, scheme)

		// bilibili 子域额外写入 .bilibili.com 根域，覆盖所有 *.bilibili.com
		if isBilibiliHost(domain) && domain != ".bilibili.com" {
			cj.setCookieEntry(e, ".bilibili.com", scheme)
		}
	}
}

// setCookieEntry 将单条 CookieEntry 写入指定 domain
func (cj *CookieJar) setCookieEntry(e CookieEntry, domain, scheme string) {
	c := e.toHTTPCookie()
	c.Domain = domain
	u := &url.URL{Scheme: scheme, Host: domain, Path: e.Path}
	cj.jar.SetCookies(u, []*http.Cookie{c})
}
