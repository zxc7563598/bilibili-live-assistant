package export

// ColumnSpec 一列的完整声明：表头文案 + 从行里取值的函数。
//
// 刻意把「声明」和「取值」写在一起，而不是像早期那样分成 exportColumns 与取数 switch
// 两份手工对齐的字面量 —— 后者加一列忘了加分支时不会编译失败，只会在导出时才暴露，
// 需要每个模块再配一个同步单测来兜。合成一份之后这类漂移在结构上就不可能出现。
//
// 泛型参数是该模块的查询行类型（model 或 JOIN 后的 ListItem）。
type ColumnSpec[T any] struct {
	// Key 列标识，与前端表格列的 key 对应；仅用于索引本模块的字面量，绝不进 SQL
	Key string
	// TitleKey i18n 表头文案键
	TitleKey string
	// Value 从一行里取该列的值；返回值交给 formatCell 渲染
	// （export.UnixTime 出时间、export.Money 出金额、实现 Text(lang) 的枚举出文案，其余按类型分发）
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

// ValueMapOf 把列声明映射为「列 key → 取值函数」，供按调用方给的列序取数。
//
// 包级初始化一次即可，导出是只读操作，之后并发调用安全。
func ValueMapOf[T any](specs []ColumnSpec[T]) map[string]func(T) any {
	m := make(map[string]func(T) any, len(specs))
	for _, s := range specs {
		m[s.Key] = s.Value
	}
	return m
}
