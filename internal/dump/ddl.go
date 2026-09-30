package dump

import (
	"strings"

	"gorm.io/gorm"
)

// createTableSQL 按模型定义生成当前方言的建表语句。
//
// GORM 不对外暴露建表 SQL（Migrator.CreateTable 内部拼完就执行，DryRun 也被它自己关掉），
// 所以这里自己拼；但列定义直接复用 Migrator.FullDataTypeOf —— 它是方言相关的，
// MySQL 给 bigint AUTO_INCREMENT、PostgreSQL 给 bigserial、SQLite 给 integer PRIMARY KEY AUTOINCREMENT。
// SQLite 那种把主键写进列类型的情况，靠 migrator 同款的「类型里是否已含 PRIMARY KEY」判断来避开重复声明。
func createTableSQL(db *gorm.DB, t Table, d Dialect) string {
	migrator := db.Migrator()
	defs := make([]string, 0, len(t.Columns)+1)
	hasPrimaryKeyInDataType := false
	for _, col := range t.Columns {
		field, ok := t.Schema.FieldsByDBName[col.Name]
		if !ok || field.IgnoreMigration {
			continue
		}
		def := migrator.FullDataTypeOf(field).SQL
		hasPrimaryKeyInDataType = hasPrimaryKeyInDataType || strings.Contains(strings.ToUpper(def), "PRIMARY KEY")
		defs = append(defs, "  "+d.quote(col.Name)+" "+def)
	}
	if !hasPrimaryKeyInDataType && len(t.Schema.PrimaryFields) > 0 {
		keys := make([]string, 0, len(t.Schema.PrimaryFields))
		for _, field := range t.Schema.PrimaryFields {
			keys = append(keys, d.quote(field.DBName))
		}
		defs = append(defs, "  PRIMARY KEY ("+strings.Join(keys, ", ")+")")
	}
	// MySQL 要显式指定字符集：库的默认字符集若是 latin1，原生导入后中文会变乱码
	tail := "\n);"
	if d.Name == DialectMySQL {
		tail = "\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"
	}
	return "CREATE TABLE " + d.quote(t.Name) + " (\n" + strings.Join(defs, ",\n") + tail
}

// createIndexSQLs 生成索引语句。
//
// 索引必须一起导出：唯一索引丢了不只是性能问题——migrate 里靠 HasIndex("uk_uid_sign_date")
// 判断历史补丁是否已执行，索引缺失会让下次启动的去重补丁再删一遍数据。
func createIndexSQLs(t Table, d Dialect) []string {
	indexes := t.Schema.ParseIndexes()
	sqls := make([]string, 0, len(indexes))
	for _, idx := range indexes {
		if len(idx.Fields) == 0 {
			continue
		}
		columns := make([]string, 0, len(idx.Fields))
		for _, field := range idx.Fields {
			columns = append(columns, d.quote(field.DBName))
		}
		var b strings.Builder
		b.WriteString("CREATE ")
		if strings.EqualFold(idx.Class, "UNIQUE") {
			b.WriteString("UNIQUE ")
		}
		b.WriteString("INDEX " + d.quote(idx.Name) + " ON " + d.quote(t.Name) +
			" (" + strings.Join(columns, ", ") + ");")
		sqls = append(sqls, b.String())
	}
	return sqls
}
