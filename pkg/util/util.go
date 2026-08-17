package util

import "unicode/utf8"

// TrimFromFront 从前面裁剪，保留后 maxLen 个字符
func TrimFromFront(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	end := len(s)
	count := 0
	for i := len(s); i > 0 && count < maxLen; {
		_, size := utf8.DecodeLastRuneInString(s[:i])
		count++
		i -= size
		if count == maxLen {
			end = i
			break
		}
	}
	return s[end:]
}

// TrimFromBack 从后面裁剪，保留前 maxLen 个字符
func TrimFromBack(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	start := 0
	count := 0
	for i := 0; i < len(s) && count < maxLen; {
		_, size := utf8.DecodeRuneInString(s[i:])
		count++
		i += size
		if count == maxLen {
			start = i
			break
		}
	}
	return s[:start]
}
