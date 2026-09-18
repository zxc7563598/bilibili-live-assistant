package sqlutil

import "strings"

// likeEscape 是 EscapeLike 使用的转义字符，SQL 里 ESCAPE 子句必须写同一个字符。
//
// 这里没有沿用反斜杠：反斜杠是 MySQL 字符串字面量自身的转义符，写 ESCAPE '\' 会被解析成
// 未闭合的字符串（得写成 '\\'），而 PostgreSQL 在 standard_conforming_strings 打开时
// '\\' 又是两个字符、报 invalid escape string —— 没有一种写法能同时喂给三家数据库。
// 换成 '!' 之后 MySQL / PostgreSQL / SQLite 的语义完全一致。
const likeEscape = '!'

// EscapeLike 转义 LIKE 模式里的通配符 % 与 _，以及转义字符本身，让它们按字面意思参与匹配。
//
// 必须同时给 SQL 补上 ESCAPE 子句，转义才会生效：
//
//	db.Where("uname LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(name)+"%")
//
// 只加转义符、不写 ESCAPE 是不够的：SQLite 没有默认转义字符，会把反斜杠当成普通字符
// 去匹配，导致搜索词里含 _ % \ 时一条都查不到。MySQL / PostgreSQL 的默认转义字符恰好
// 是反斜杠，所以漏写 ESCAPE 在那两家上碰巧是对的，问题只在 SQLite 上暴露。
func EscapeLike(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '%' || r == '_' || r == likeEscape {
			b.WriteRune(likeEscape)
		}
		b.WriteRune(r)
	}
	return b.String()
}
