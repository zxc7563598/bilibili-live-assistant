package live

import (
	"math/rand/v2"
	"regexp"
)

// 匹配 @变量名@ 占位符
var tmplVarPattern = regexp.MustCompile(`@(\w+)@`)

// CollectVars 扫描模板列表，收集所有被引用的变量名。
// 各业务模块在解析变量前先调用此函数收集变量名，再按需查询数据。
func CollectVars(templates ...[]string) map[string]bool {
	used := make(map[string]bool)
	for _, group := range templates {
		for _, tmpl := range group {
			for _, match := range tmplVarPattern.FindAllStringSubmatch(tmpl, -1) {
				used[match[1]] = true
			}
		}
	}
	return used
}

// RenderTemplate 替换模板中的 @var@ 占位符为实际值。
// 未找到对应变量的占位符将保留原样。
func RenderTemplate(tmpl string, vars map[string]string) string {
	return tmplVarPattern.ReplaceAllStringFunc(tmpl, func(match string) string {
		key := match[1 : len(match)-1] // 去掉首尾的 @
		if v, ok := vars[key]; ok {
			return v
		}
		return match
	})
}

// PickRandom 从切片中随机选取一个元素。
func PickRandom[T any](items []T) T {
	var zero T
	if len(items) == 0 {
		return zero
	}
	return items[rand.IntN(len(items))]
}
