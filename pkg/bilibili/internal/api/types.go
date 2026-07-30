// Package api 提供 B站 API 通用类型定义和错误处理工具。
package api

import "fmt"

// Response 是 B站 API 的通用响应信封。
//
// 所有 B站 API 响应的顶层结构均为：
//
//	{"code": 0, "message": "0", "data": ...}
//
//   - Code: 业务状态码，0 表示成功，非 0 表示业务错误
//   - Message: 状态信息，成功时通常为 "0" 或空字符串
//
// 各子包定义的具体 API 响应结构体应嵌入此类型以复用 JSON 解码和错误检查。
type Response struct {
	Code    int    `json:"code"`    // 业务状态码，0 表示成功，非 0 表示错误
	Message string `json:"message"` // 状态信息，成功时通常为 "0"
}

// CheckError 检查 API 响应码，非 0 时返回包含错误码和错误信息的 error。
func CheckError(code int, message string) error {
	if code != 0 {
		return fmt.Errorf("B站 API 错误：code=%d, message=%s", code, message)
	}
	return nil
}
