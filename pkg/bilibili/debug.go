package bilibili

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

// debugTransport 包装 http.RoundTripper，在每次请求/响应时输出完整的 HTTP 原始报文
//
// http.Client.Do() 在调用 RoundTrip 之前已将 CookieJar 中的 Cookie 注入 Header，
// 所以 dump 出来的请求会包含完整的 Cookie（SESSDATA、bili_jct 等）
// 如果在 doRequest 层用 httputil.DumpRequestOut，会看不到 Jar 自动注入的 Cookie
type debugTransport struct {
	transport http.RoundTripper
	writer    io.Writer
}

func (t *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// dump 请求（含 Cookie、Body 等完整信息）
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		fmt.Fprintf(t.writer, "[DEBUG] dump request error: %v\n", err)
	} else {
		fmt.Fprintf(t.writer, "\n========== REQUEST ==========\n%s\n", reqDump)
	}
	// 实际执行请求
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(t.writer, "[DEBUG] roundtrip error: %v\n", err)
		return resp, err
	}
	// dump 响应（先保存 body，dump 完再恢复，否则调用方读不到）
	bodyBytes, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		fmt.Fprintf(t.writer, "[DEBUG] read body error: %v\n", readErr)
		resp.Body = io.NopCloser(bytes.NewReader(nil))
	} else {
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}
	// dump 响应状态行 + header
	respDump, err := httputil.DumpResponse(resp, false)
	if err != nil {
		fmt.Fprintf(t.writer, "[DEBUG] dump response error: %v\n", err)
	} else {
		fmt.Fprintf(t.writer, "========== RESPONSE ==========\n%s\n", respDump)
		// 继续输出 body（最多 8KB，避免日志爆炸）
		if len(bodyBytes) > 0 {
			maxBody := bodyBytes
			if len(maxBody) > 8192 {
				maxBody = maxBody[:8192]
				fmt.Fprintf(t.writer, "%s\n... (截断，共 %d bytes)\n", maxBody, len(bodyBytes))
			} else {
				fmt.Fprintf(t.writer, "%s\n", maxBody)
			}
		}
	}
	// 恢复 body，调用方仍可正常读取
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return resp, nil
}
