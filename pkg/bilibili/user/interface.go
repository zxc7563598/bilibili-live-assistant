// Package user 提供 B站 用户信息相关的 API 调用。
// 本包不导入父包 bilibili，通过自有的 HttpClient 接口实现依赖反转。
package user

import (
	"context"
)

// HttpClient 是本包需要的 HTTP 客户端接口。
type HttpClient interface {
	Get(ctx context.Context, path string, result any) error
	Post(ctx context.Context, path string, body any, result any) error
}
