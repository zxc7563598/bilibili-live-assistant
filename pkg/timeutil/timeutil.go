package timeutil

import "time"

// Format 将时间戳格式化为标准格式
func Format(ts int64) string {
	return time.Unix(ts, 0).Format(time.DateTime)
}

// FormatWithLayout 使用自定义格式格式化时间
func FormatWithLayout(ts int64, layout string) string {
	return time.Unix(ts, 0).Format(layout)
}

// Parse 解析标准格式的时间字符串
func Parse(timeStr string) (time.Time, error) {
	return time.Parse(time.DateTime, timeStr)
}
