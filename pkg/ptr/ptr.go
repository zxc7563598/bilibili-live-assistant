package ptr

// Uint64 将 uint64 转换为 *uint64
func Uint64(v uint64) *uint64 {
	return &v
}

// Int 将 int 转换为 *int
func Int(v int) *int {
	return &v
}

// String 将 string 转换为 *string
func String(v string) *string {
	return &v
}

// Bool 将 bool 转换为 *bool
func Bool(v bool) *bool {
	return &v
}

// Deref 安全地将指针解引用为值，如果指针为 nil 则返回零值
func Deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
