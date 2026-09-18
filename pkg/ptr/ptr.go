package ptr

import (
	"strconv"
	"strings"
)

// Deref 安全地将指针解引用为值，如果指针为 nil 则返回零值
func Deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

// TrimStr 安全地解引用字符串指针并去除首尾空白，指针为 nil 时返回空字符串
func TrimStr(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

// ParseBool 将配置字符串解析为 bool，"1" 为 true
func ParseBool(s string) bool {
	return s == "1"
}

// ParseEnumInt 将配置字符串解析为 int 枚举值，解析失败返回 0
func ParseEnumInt[T ~int](s string) T {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return T(v)
}
