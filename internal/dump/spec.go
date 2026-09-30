// Package dump 提供数据库导出（生成方言原生 SQL 文件）与导入（按配置写入目标库）能力。
//
// 导出文件同时服务两类使用者：外部数据库客户端（sqlite3 / mysql / psql）可直接导入，
// 本系统的 db import 也能读它并跨方言写入。文件里的元信息以行注释承载，
// 既不影响原生客户端执行，又让导入侧不必猜列类型。
package dump

import (
	"encoding/json"
	"fmt"
	"strings"
)

// 支持的数据库方言，取值与 GORM 驱动名一致
const (
	DialectMySQL    = "mysql"
	DialectPostgres = "postgres"
	DialectSQLite   = "sqlite"
)

// 值的类型。导出文件不写具体 SQL 类型，只写这几种中性类型，导入侧据此还原成 Go 值。
const (
	KindInt    = "int"
	KindFloat  = "float"
	KindString = "string"
	KindBytes  = "bytes"
	KindTime   = "time"
	KindBool   = "bool"
)

// 文件内的标记，均以 SQL 行注释写出
const (
	headerMarker = "BLA-DUMP"
	tableMarker  = "BLA-TABLE"
	endMarker    = "BLA-END"
)

// FileVersion 导出文件格式版本，格式变更时递增，导入侧据此拒绝不认识的文件
const FileVersion = 1

// headerNote 写在文件头、给用户看的说明
const headerNote = "本文件可直接用 sqlite3 / mysql / psql 客户端导入，导入前请确保目标库为空。" +
	"也可以执行 `BiliLiveAssistant db import -in <本文件>` 跨数据库导入。"

// Header 文件头
type Header struct {
	Version   int    `json:"version"`
	Dialect   string `json:"dialect"`
	CreatedAt int64  `json:"created_at"`
}

// ColumnMeta 一列的元信息
type ColumnMeta struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Nullable bool   `json:"nullable"`
}

// TableMeta 一张表的元信息，写在建表语句之后、数据之前
type TableMeta struct {
	Name    string       `json:"name"`
	Rows    int64        `json:"rows"`
	Columns []ColumnMeta `json:"columns"`
}

// Footer 文件尾
type Footer struct {
	Tables int   `json:"tables"`
	Rows   int64 `json:"rows"`
}

// Dialect 一种方言的语法差异
type Dialect struct {
	Name string
}

// Valid 是否受支持的方言
func (d Dialect) Valid() bool {
	switch d.Name {
	case DialectMySQL, DialectPostgres, DialectSQLite:
		return true
	}
	return false
}

// quote 引用标识符（表名 / 列名 / 索引名）
func (d Dialect) quote(name string) string {
	if d.Name == DialectMySQL {
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// marker 生成一行标记注释
func marker(name string, payload any) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("序列化 %s 标记失败: %w", name, err)
	}
	return "-- " + name + " " + string(data), nil
}
