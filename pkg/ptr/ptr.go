package ptr

import "strconv"

// Deref 安全地将指针解引用为值，如果指针为 nil 则返回零值
func Deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
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
