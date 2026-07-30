package bilibili

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// resolveURL 处理 URL：若 path 已是完整 URL 则直接返回，否则拼接 baseURL
func (c *Client) resolveURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return c.baseURL + path
}

// Get 发送 GET 请求，响应体自动 JSON 解码到 result
// path 可以是相对路径（拼接到 baseURL）或完整 URL
// Cookie 由 http.Client.Jar 自动管理
func (c *Client) Get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.resolveURL(path), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	return c.doRequest(req, result)
}

// Post 发送 POST 请求，body 自动 JSON 序列化，响应体自动 JSON 解码到 result
func (c *Client) Post(ctx context.Context, path string, body any, result any) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.resolveURL(path), bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req, result)
}

// PostForm 发送 POST 请求，body 以 application/x-www-form-urlencoded 编码
// B站 绝大多数 POST 接口（发送弹幕、禁言管理等）使用此格式
func (c *Client) PostForm(ctx context.Context, path string, form url.Values, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.resolveURL(path), strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.doRequest(req, result)
}

// Do 执行原始 HTTP 请求，同时注入通用 Headers（User-Agent、Referer）
// Cookie 仍由 http.Client.Jar 自动管理
// 典型用途：WebSocket 握手等需要完全控制请求的场景
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.setHeaders(req)
	return c.httpClient.Do(req)
}

// doRequest 是内部通用请求方法，负责设置 Headers、执行请求、状态码检查和 JSON 解码
func (c *Client) doRequest(req *http.Request, result any) error {
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// setHeaders 设置所有请求的通用 Headers
// Cookie 不在此处手动设置——由 http.Client.Jar 自动注入
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Origin", "https://live.bilibili.com")
	req.Header.Set("Referer", "https://live.bilibili.com")
}
