package bilibili

import (
	"encoding/json"
	"net/url"
	"os"
)

// bilibiliDomains 是 Cookie 持久化时需要覆盖的所有 B站 相关域名
var bilibiliDomains = []string{
	"https://api.bilibili.com",
	"https://live.bilibili.com",
	"https://passport.bilibili.com",
	"https://www.bilibili.com",
	"https://space.bilibili.com",
}

// State 是 Client 的完整可持久化状态，包含用户身份信息和所有 Cookie
//
// 典型用法：
//
//	// 首次登录后保存
//	client.SaveState("bilibili_state.json")
//
//	// 重启时加载
//	client := NewClient(WithStateFile("bilibili_state.json"))
type State struct {
	Session *Session      `json:"session,omitempty"`
	Cookies []CookieEntry `json:"cookies,omitempty"`
}

// SaveState 将当前客户端状态（Session + CookieJar 中所有 B站 域名的 Cookie）
// 序列化为 JSON 并写入指定文件
//
// 文件权限为 0600（仅 owner 可读写），因为 Cookie 中包含敏感信息（SESSDATA 等）
func (c *Client) SaveState(filePath string) error {
	// 构造 URL 列表
	urls := make([]*url.URL, len(bilibiliDomains))
	for i, d := range bilibiliDomains {
		u, err := url.Parse(d)
		if err != nil {
			return err
		}
		urls[i] = u
	}

	state := State{
		Session: c.session,
		Cookies: c.cookieJar.ExportAll(urls...),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0600)
}

// LoadState 从 JSON 文件加载客户端状态，恢复 Session 和 Cookie
//
// 加载后的 Cookie 会被写入 CookieJar，后续所有请求自动携带
// 如果文件不存在或格式错误，返回 error（调用方可选择忽略）
func (c *Client) LoadState(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	// 恢复 Session
	if state.Session != nil {
		c.session = state.Session

		// 如果 Session 包含 Buvid，写入 CookieJar
		if state.Session.Buvid != "" {
			u, _ := url.Parse("https://api.bilibili.com")
			c.cookieJar.SetCookieString(u, "buvid3="+state.Session.Buvid)
		}
	}

	// 恢复 Cookie
	if len(state.Cookies) > 0 {
		c.cookieJar.ImportEntries(state.Cookies)
	}

	return nil
}
