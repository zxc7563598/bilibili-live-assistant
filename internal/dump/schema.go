package dump

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/migrate"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Column 一列的元信息，全部取自模型定义
type Column struct {
	Name     string
	Kind     string
	Nullable bool
}

// Table 一张表的元信息
type Table struct {
	Name    string
	Schema  *schema.Schema
	Columns []Column
}

// PrimaryColumn 主键列名，导出时作为分块游标
func (t Table) PrimaryColumn() (string, bool) {
	if len(t.Schema.PrimaryFields) == 0 {
		return "", false
	}
	return t.Schema.PrimaryFields[0].DBName, true
}

// tableOrder 导出与导入的表顺序，按 xxx_id 的引用关系排列，被引用的表在前。
//
// 模型里没有任何关联声明，数据库层也就没有外键约束，所以这个顺序不影响成败；
// 它保证的是导入过程中不会出现"引用了还不存在的行"的中间状态。
var tableOrder = []string{
	"roles",
	"menus",
	"admins",
	"role_menus",
	"admin_roles",
	"app_configs",
	"robot_configs",
	"live_users",
	"live_sessions",
	"live_pk_logs",
	"live_danmus",
	"live_gifts",
	"live_interact_words",
	"live_user_sign_logs",
	"live_user_blacklists",
	"live_user_credit_logs",
	"live_user_addresses",
	"feedbacks",
	"products",
	"product_skus",
	"product_specs",
	"product_spec_values",
	"product_images",
	"product_sku_stock_logs",
	"live_user_orders",
	"live_user_order_drafts",
}

// Tables 解析全部模型的表结构，按 tableOrder 返回。
//
// 表顺序清单写死在 tableOrder，新增模型若忘了登记会直接报错——
// 漏一张表的表现是"导出文件悄悄少了数据"，比启动失败难发现得多。
func Tables(db *gorm.DB) ([]Table, error) {
	models := migrate.Models()
	byName := make(map[string]Table, len(models))
	for _, m := range models {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(m); err != nil {
			return nil, fmt.Errorf("解析模型 %T 失败: %w", m, err)
		}
		byName[stmt.Schema.Table] = newTable(stmt.Schema)
	}
	tables := make([]Table, 0, len(models))
	for _, name := range tableOrder {
		t, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("表顺序清单中的 %s 没有对应模型，请检查 internal/dump/schema.go 的 tableOrder", name)
		}
		tables = append(tables, t)
		delete(byName, name)
	}
	if len(byName) > 0 {
		missing := make([]string, 0, len(byName))
		for name := range byName {
			missing = append(missing, name)
		}
		sort.Strings(missing)
		return nil, fmt.Errorf("模型 %s 未登记表顺序，请补进 internal/dump/schema.go 的 tableOrder", strings.Join(missing, "、"))
	}
	return tables, nil
}

// newTable 从 GORM 的 schema 提取表结构
func newTable(s *schema.Schema) Table {
	columns := make([]Column, 0, len(s.DBNames))
	for _, name := range s.DBNames {
		field := s.FieldsByDBName[name]
		columns = append(columns, Column{
			Name:     name,
			Kind:     kindOf(field),
			Nullable: !field.NotNull,
		})
	}
	return Table{Name: s.Table, Schema: s, Columns: columns}
}

// kindOf 判定列的中性类型。
//
// 优先看 GORM 的数据类型：带 type 标签的字段（如 type:smallint 的枚举）DataType 会是
// 原始 SQL 类型名而非标准类型，此时退回按 Go 类型判断。
func kindOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool:
		return KindBool
	case schema.Int, schema.Uint:
		return KindInt
	case schema.Float:
		return KindFloat
	case schema.String:
		return KindString
	case schema.Bytes:
		return KindBytes
	case schema.Time:
		return KindTime
	}

	switch field.IndirectFieldType.Kind() {
	case reflect.Bool:
		return KindBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return KindInt
	case reflect.Float32, reflect.Float64:
		return KindFloat
	case reflect.String:
		return KindString
	case reflect.Slice:
		if field.IndirectFieldType.Elem().Kind() == reflect.Uint8 {
			return KindBytes
		}
	case reflect.Struct:
		if field.IndirectFieldType == reflect.TypeOf(time.Time{}) {
			return KindTime
		}
	}
	return KindString
}
