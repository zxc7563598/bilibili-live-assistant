// Package room 提供 B站 直播间相关的 API 调用。
// 本包不导入父包 bilibili，通过自有的 HttpClient 接口实现依赖反转。
package room

import (
	"context"
	"net/url"
)

// HttpClient 是本包需要的 HTTP 客户端接口。
// 比 auth 包多了 PostForm 方法（发送弹幕、禁言管理等 B站 POST 接口使用 form-urlencoded）。
type HttpClient interface {
	Get(ctx context.Context, path string, result any) error
	Post(ctx context.Context, path string, body any, result any) error
	PostForm(ctx context.Context, path string, form url.Values, result any) error
}
