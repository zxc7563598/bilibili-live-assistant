package migrate

import (
	"fmt"
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"gorm.io/gorm"
)

// explicitIDSeedModels 种子数据中以固定主键 ID 写入的表所对应的模型
//
// 这些表的初始数据带写死的 ID（角色、菜单、管理员、管理员角色），
// 用于建立固定的角色-菜单-管理员关联关系。
var explicitIDSeedModels = []any{
	&model.Role{},
	&model.Menu{},
	&model.Admin{},
	&model.AdminRole{},
}

// explicitIDSeedTables 返回上述模型对应的表名
func explicitIDSeedTables(db *gorm.DB) ([]string, error) {
	tables := make([]string, 0, len(explicitIDSeedModels))
	for _, m := range explicitIDSeedModels {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(m); err != nil {
			return nil, fmt.Errorf("解析 %T 表名失败: %w", m, err)
		}
		tables = append(tables, stmt.Schema.Table)
	}
	return tables, nil
}

// syncPostgresSequences 把 PostgreSQL 的自增序列推进到各表当前最大 ID 之后，其余数据库为空操作。
//
// 背景：模型主键是裸的 ID int64 `gorm:"primaryKey"`，GORM 在 PostgreSQL 下会建成
// bigserial（等价于 bigint DEFAULT nextval('xxx_id_seq')）。serial 序列只在 INSERT
// 省略该列时才取 nextval，而种子数据是显式写入 ID 的，序列因此仍停在起点；之后
// 业务侧新增（不带 ID）取到的就是已被占用的 1，直接报主键冲突。
// MySQL 的 AUTO_INCREMENT 与 SQLite 的 rowid 在显式插入时都会自动抬到 MAX(id)+1，
// 不存在这个问题，所以只在 PostgreSQL 上做补偿。
//
// setval 第三个参数传 false 表示设定值本身即下一个 nextval 的返回值，
// 因此取 MAX(id)+1 后下一个自增 ID 恰好是当前最大值加一；表为空时回到 1。
func syncPostgresSequences(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	tables, err := explicitIDSeedTables(db)
	if err != nil {
		return err
	}
	for _, table := range tables {
		// 先取出该列绑定的序列名：手工建表等情况下 id 可能没有序列，此时跳过而不是让启动失败
		var sequence string
		if err := db.Raw("SELECT COALESCE(pg_get_serial_sequence(?, 'id'), '')", table).Scan(&sequence).Error; err != nil {
			return fmt.Errorf("查询 %s 自增序列失败: %w", table, err)
		}
		if sequence == "" {
			log.Printf("[migrate] 表 %s 的 id 未绑定自增序列，跳过序列同步", table)
			continue
		}
		// sequence 是数据库返回的标识符，走参数绑定并显式转 regclass：
		// setval 只接受 regclass，直接传 text 参数会因类型推断失败而找不到函数。
		// 表名来自模型定义，非外部输入，可安全内联。
		sql := fmt.Sprintf(
			"SELECT setval(?::regclass, (SELECT COALESCE(MAX(id), 0) + 1 FROM %s), false)",
			table,
		)
		if err := db.Exec(sql, sequence).Error; err != nil {
			return fmt.Errorf("同步 %s 自增序列失败: %w", table, err)
		}
	}
	return nil
}
