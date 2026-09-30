package dump

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// timeLayouts 导入时尝试的时间格式，覆盖三种驱动返回的写法
var timeLayouts = []string{
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999-07:00",
	"2006-01-02 15:04:05-07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05.999999999Z07:00",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02",
}

// timeLayout 渲染时间用的格式。
//
// MySQL 的 datetime 不接受时区偏移，只能写裸的本地时间，但必须带上小数秒——
// 软删除列建的是 datetime(3)，截掉毫秒就是静默丢精度；
// SQLite 把时间存成文本、PostgreSQL 的 timestamptz 自带偏移，两者都要带偏移量，
// 否则导入后时间会整体偏移（曾被丢掉 +08:00 导致差 8 小时）。
func timeLayout(d Dialect) string {
	if d.Name == DialectMySQL {
		return "2006-01-02 15:04:05.999999"
	}
	return "2006-01-02 15:04:05.999999999-07:00"
}

// renderValue 把驱动返回的值渲染成当前方言的 SQL 字面量
func renderValue(v any, kind string, d Dialect) (string, error) {
	if v == nil {
		return "NULL", nil
	}
	switch kind {
	case KindTime:
		at, err := toTime(v)
		if err != nil {
			return "", err
		}
		return "'" + at.Format(timeLayout(d)) + "'", nil
	case KindBytes:
		data, err := toBytes(v)
		if err != nil {
			return "", err
		}
		// PostgreSQL 的 bytea 十六进制输入法，另外两种方言是标准的 X'..'
		if d.Name == DialectPostgres {
			return `'\x` + hex.EncodeToString(data) + `'::bytea`, nil
		}
		return "X'" + hex.EncodeToString(data) + "'", nil
	case KindBool:
		value, err := toBool(v)
		if err != nil {
			return "", err
		}
		if d.Name == DialectPostgres {
			// PostgreSQL 的 boolean 列不接受裸的 1/0（那是整数字面量），必须写 TRUE/FALSE
			if value {
				return "TRUE", nil
			}
			return "FALSE", nil
		}
		if value {
			return "1", nil
		}
		return "0", nil
	case KindInt:
		value, err := toInt(v)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(value, 10), nil
	case KindFloat:
		value, err := toFloat(v)
		if err != nil {
			return "", err
		}
		return strconv.FormatFloat(value, 'f', -1, 64), nil
	default:
		return renderString(toString(v), d)
	}
}

// renderString 渲染字符串字面量。
//
// 含 NUL 字节的值必须特殊处理：NUL 是 C 系 SQL 客户端的字符串终止符，
// 原样写进文件会让 sqlite3 / mysql 客户端从这里读串位（弹幕文本里出现过这种脏数据）。
func renderString(s string, d Dialect) (string, error) {
	if !strings.ContainsRune(s, 0) {
		return quoteString(s, d), nil
	}
	switch d.Name {
	case DialectMySQL:
		// MySQL 用 \0 表示 NUL，与 mysqldump 一致
		return quoteString(s, d), nil
	case DialectSQLite:
		// SQLite 没有表示 NUL 的转义序列，只能整串写成十六进制再转回文本
		return "CAST(X'" + hex.EncodeToString([]byte(s)) + "' AS TEXT)", nil
	default:
		return "", fmt.Errorf("值包含 NUL 字节，PostgreSQL 的文本列无法存储，请先清理该列数据")
	}
}

// quoteString 转义字符串字面量。
//
// MySQL 默认把反斜杠当转义符，不转义的话 "C:\path" 会被读成 "C:path"；
// 同理 NUL 与 0x1A 也要转义，否则客户端可能提前结束读取。
// PostgreSQL（standard_conforming_strings 默认开启）与 SQLite 按标准处理，原样输出即可。
func quoteString(s string, d Dialect) string {
	if d.Name != DialectMySQL {
		return "'" + strings.ReplaceAll(s, "'", "''") + "'"
	}
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('\'')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString("''")
		case 0:
			b.WriteString(`\0`)
		case 0x1a:
			b.WriteString(`\Z`)
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('\'')
	return b.String()
}

// parseLiteral 把导出文件里的字面量还原成 Go 值
func parseLiteral(token, kind string, d Dialect) (any, error) {
	token = strings.TrimSpace(token)
	if strings.EqualFold(token, "NULL") {
		return nil, nil
	}
	// 含 NUL 的字符串导出时写成了十六进制转换函数，先还原回来
	if data, ok, err := parseHexCast(token); err != nil {
		return nil, err
	} else if ok {
		if kind == KindBytes {
			return data, nil
		}
		return string(data), nil
	}
	switch kind {
	case KindBytes:
		return parseBytes(token, d)
	case KindBool:
		return parseBool(token)
	case KindInt:
		return parseNumbered(token, true)
	case KindFloat:
		return parseNumbered(token, false)
	case KindTime:
		text, err := parseString(token, d)
		if err != nil {
			return nil, err
		}
		return parseTime(text)
	default:
		return parseString(token, d)
	}
}

// parseString 还原字符串字面量
func parseString(token string, d Dialect) (string, error) {
	if len(token) < 2 || token[0] != '\'' || token[len(token)-1] != '\'' {
		return "", fmt.Errorf("不是合法的字符串字面量: %s", preview(token))
	}
	return unescapeString(token[1:len(token)-1], d), nil
}

// unescapeString 还原导出时的转义。只处理导出侧会写出的几种：” 表示单引号、
// MySQL 下 \\ 表示反斜杠、\0 表示 NUL、\Z 表示 0x1A；
// 单趟扫描即可，不会出现多次解转义的问题。
func unescapeString(s string, d Dialect) string {
	escapeBackslash := d.Name == DialectMySQL
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if escapeBackslash && c == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case '0':
				b.WriteByte(0)
			case 'Z', 'z':
				b.WriteByte(0x1a)
			default:
				b.WriteByte(s[i])
			}
			continue
		}
		if c == '\'' && i+1 < len(s) && s[i+1] == '\'' {
			b.WriteByte('\'')
			i++
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

// parseBytes 还原字节串字面量，接受 X'..' 与 PostgreSQL 的 '\x..'::bytea 两种写法
func parseBytes(token string, d Dialect) ([]byte, error) {
	if strings.HasPrefix(token, "X'") || strings.HasPrefix(token, "x'") {
		return hex.DecodeString(strings.TrimSuffix(token[2:], "'"))
	}
	text, err := parseString(strings.TrimSuffix(token, "::bytea"), d)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(strings.TrimPrefix(text, `\x`))
}

// parseHexCast 解析 CAST(X'..' AS TEXT) 与 convert_from(X'..','UTF8') 这类十六进制转文本的写法
func parseHexCast(token string) ([]byte, bool, error) {
	upper := strings.ToUpper(token)
	if !strings.HasPrefix(upper, "CAST(") && !strings.HasPrefix(upper, "CONVERT_FROM(") {
		return nil, false, nil
	}
	index := strings.Index(upper, "X'")
	if index < 0 {
		return nil, false, fmt.Errorf("无法解析字面量: %s", preview(token))
	}
	rest := token[index+2:]
	end := strings.Index(rest, "'")
	if end < 0 {
		return nil, false, fmt.Errorf("无法解析字面量: %s", preview(token))
	}
	data, err := hex.DecodeString(rest[:end])
	if err != nil {
		return nil, false, fmt.Errorf("十六进制字面量解析失败: %w", err)
	}
	return data, true, nil
}

// parseBool 还原布尔字面量
func parseBool(token string) (bool, error) {
	switch strings.ToUpper(token) {
	case "TRUE", "1":
		return true, nil
	case "FALSE", "0":
		return false, nil
	}
	return false, fmt.Errorf("不是合法的布尔字面量: %s", preview(token))
}

// parseNumbered 还原数值字面量，asInt 为真时按整数解析
func parseNumbered(token string, asInt bool) (any, error) {
	text := token
	// 容错：数值被引号包起来时先脱引号
	if len(text) >= 2 && text[0] == '\'' && text[len(text)-1] == '\'' {
		text = text[1 : len(text)-1]
	}
	if asInt {
		value, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("不是合法的整数字面量: %s", preview(token))
		}
		return value, nil
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil, fmt.Errorf("不是合法的浮点字面量: %s", preview(token))
	}
	return value, nil
}

// toTime 归一化驱动返回的时间值
func toTime(v any) (time.Time, error) {
	switch value := v.(type) {
	case time.Time:
		return value, nil
	case string:
		return parseTime(value)
	case []byte:
		return parseTime(string(value))
	}
	return time.Time{}, fmt.Errorf("无法转换为时间: %v (%T)", v, v)
}

// parseTime 按候选格式逐个尝试解析，全部失败时报错而不是写零值——
// 静默写坏时间比直接失败难排查得多。
func parseTime(text string) (time.Time, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return time.Time{}, fmt.Errorf("时间字面量为空")
	}
	for _, layout := range timeLayouts {
		if at, err := time.Parse(layout, text); err == nil {
			return at, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法识别的时间格式: %s", preview(text))
}

func toInt(v any) (int64, error) {
	switch value := v.(type) {
	case int64:
		return value, nil
	case int:
		return int64(value), nil
	case int8:
		return int64(value), nil
	case int16:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case uint:
		return int64(value), nil
	case uint8:
		return int64(value), nil
	case uint16:
		return int64(value), nil
	case uint32:
		return int64(value), nil
	case uint64:
		return int64(value), nil
	case float64:
		if value == math.Trunc(value) {
			return int64(value), nil
		}
	case bool:
		if value {
			return 1, nil
		}
		return 0, nil
	case string:
		if n, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64); err == nil {
			return n, nil
		}
	case []byte:
		if n, err := strconv.ParseInt(strings.TrimSpace(string(value)), 10, 64); err == nil {
			return n, nil
		}
	}
	return 0, fmt.Errorf("无法转换为整数: %v (%T)", v, v)
}

func toFloat(v any) (float64, error) {
	switch value := v.(type) {
	case float64:
		return value, nil
	case float32:
		return float64(value), nil
	case int64:
		return float64(value), nil
	case int:
		return float64(value), nil
	case int32:
		return float64(value), nil
	case uint64:
		return float64(value), nil
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err == nil {
			return f, nil
		}
	case []byte:
		if f, err := strconv.ParseFloat(strings.TrimSpace(string(value)), 64); err == nil {
			return f, nil
		}
	}
	return 0, fmt.Errorf("无法转换为浮点数: %v (%T)", v, v)
}

func toBool(v any) (bool, error) {
	switch value := v.(type) {
	case bool:
		return value, nil
	case int64:
		return value != 0, nil
	case int:
		return value != 0, nil
	case []byte:
		return strconv.ParseBool(strings.TrimSpace(string(value)))
	case string:
		return strconv.ParseBool(strings.TrimSpace(value))
	}
	return false, fmt.Errorf("无法转换为布尔值: %v (%T)", v, v)
}

func toBytes(v any) ([]byte, error) {
	switch value := v.(type) {
	case []byte:
		return value, nil
	case string:
		return []byte(value), nil
	}
	return nil, fmt.Errorf("无法转换为字节串: %v (%T)", v, v)
}

func toString(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	}
	return fmt.Sprint(v)
}

// preview 截断过长的字面量，避免报错信息刷屏
func preview(text string) string {
	const maxLength = 60
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}
