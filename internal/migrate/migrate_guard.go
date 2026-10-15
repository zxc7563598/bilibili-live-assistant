package migrate

import (
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"gorm.io/gorm"
)

// dropLegacyLiveGuardRecords 升级暂存期跑过的库：删掉还是「有效期窗口」结构的 live_guard_records
//
// 必须在 Run() 的 AutoMigrate 之前执行。台账早期存的是 30 天有效期窗口（start_at /
// end_at），现在存的是名单月份区间（cover_from / cover_to，两端都是月首），而 GORM
// 不做列改名，只会对已存在的表执行
//
//	ALTER TABLE live_guard_records ADD COLUMN cover_from integer NOT NULL
//
// SQLite 与 PostgreSQL 都拒绝这句（SQLite 报 Cannot add a NOT NULL column with default
// value NULL），迁移失败会让进程起不来；MySQL 倒是会给存量行塞 0，留下一批没有月份含义的行。
//
// 删表无损：台账每一行都能从 live_gifts 重放（补录同样会写一条礼物记录），而
// seedLiveGuardRecords 恰好在表为空时执行，当次启动就会按新口径重建。
func dropLegacyLiveGuardRecords(db *gorm.DB) error {
	m := db.Migrator()
	if !m.HasTable(&model.LiveGuardRecord{}) {
		// 全新库：交给 AutoMigrate 建表
		return nil
	}
	// 旧字段已从模型里删掉，HasColumn 查不到字段时按裸列名处理
	if !m.HasColumn(&model.LiveGuardRecord{}, "start_at") || m.HasColumn(&model.LiveGuardRecord{}, "cover_from") {
		return nil
	}
	table := model.LiveGuardRecord{}.TableName()
	log.Printf("[migrate] 大航海台账为旧结构，删除后按名单月份区间重建: %s", table)
	return m.DropTable(table)
}
