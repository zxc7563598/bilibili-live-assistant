package export

import "fmt"

// ColumnSpec 一列的完整声明
type ColumnSpec[T any] struct {
	// Key 列标识，与前端表格列的 key 对应；仅用于索引本模块的字面量，绝不进 SQL
	Key string
	// TitleKey i18n 表头文案键
	TitleKey string
	// Value 从一行里取该列的值；返回值交给 formatCell 渲染
	Value func(T) any
}

// ColumnsOf 把列声明映射为框架使用的列清单
func ColumnsOf[T any](specs []ColumnSpec[T]) []Column {
	cols := make([]Column, 0, len(specs))
	for _, s := range specs {
		cols = append(cols, Column{Key: s.Key, TitleKey: s.TitleKey})
	}
	return cols
}

// ValueMapOf 把列声明映射为「列 key → 取值函数」，供按调用方给的列序取数
func ValueMapOf[T any](specs []ColumnSpec[T]) map[string]func(T) any {
	m := make(map[string]func(T) any, len(specs))
	for _, s := range specs {
		m[s.Key] = s.Value
	}
	return m
}

// BuildRecords 按调用方给的列序从每行取值，组装成一块导出数据
func BuildRecords[T any](rows []T, keys []string, values map[string]func(T) any, idOf func(T) int64) ([][]any, int64, error) {
	out := make([][]any, 0, len(rows))
	for _, row := range rows {
		rec := make([]any, 0, len(keys))
		for _, key := range keys {
			// keys 已由 resolveColumns 按 Columns() 校验过，这里的兜底只为防两处声明分叉
			value, ok := values[key]
			if !ok {
				return nil, 0, fmt.Errorf("未支持的导出列: %q", key)
			}
			rec = append(rec, value(row))
		}
		out = append(out, rec)
	}
	// 空块时游标用不到，由调用方按「不足一块」结束循环
	var nextID int64
	if len(rows) > 0 {
		nextID = idOf(rows[len(rows)-1])
	}
	return out, nextID, nil
}
