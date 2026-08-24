package migrate

import (
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"gorm.io/gorm"
)

// liveUserBackfillBatchSize 回填时每批读取的待处理行数，避免大表一次性载入内存
const liveUserBackfillBatchSize = 500

// backfillLiveUserDanmuGift 兼容历史数据：为已存在的 live_users 表回填 total_danmu_count / total_gift_amount
//
// 依赖 Run() 中先执行的 AutoMigrate 创建这两列（不在此手动写 ALTER DDL）。
// 幂等策略：只处理 total_danmu_count = 0 AND total_gift_amount = 0 的行——这两列刚被 AutoMigrate
// 以 DEFAULT 0 创建时存量行全为 0；已被业务追加（IncrementField）或已被回填过的行必然至少一列为非 0，
// 天然不会被触碰。半途中断后下次启动能自动补齐剩余全 0 行；真实历史就是 (0,0) 的用户每次启动会被重扫
// 并判定为无需写回，属于无副作用的幂等操作。
//
// 回填语义与 EnsureUser 完全一致：
//
//	total_danmu_count = COUNT(*) FROM live_danmus WHERE uid = ? AND deleted_at IS NULL
//	total_gift_amount = COALESCE(SUM(num * price), 0) FROM live_gifts WHERE uid = ? AND deleted_at IS NULL
func backfillLiveUserDanmuGift(db *gorm.DB) error {
	m := db.Migrator()
	if !m.HasTable(&model.LiveUser{}) {
		return nil
	}
	if !m.HasColumn(&model.LiveUser{}, "TotalDanmuCount") {
		return nil
	}
	liveUserTable := model.LiveUser{}.TableName()
	danmuTable := model.LiveDanmu{}.TableName()
	giftTable := model.LiveGift{}.TableName()
	// 没有全 0 行就不必跑两条聚合 SQL（全新库、已回填完的库直接返回）
	var pendingCount int64
	if err := db.Model(&model.LiveUser{}).
		Where("total_danmu_count = 0 AND total_gift_amount = 0").
		Count(&pendingCount).Error; err != nil {
		return err
	}
	if pendingCount == 0 {
		return nil
	}
	// 两条 GROUP BY 一次性取出全表聚合
	type aggRow struct {
		UID   int64
		Total int64
	}
	var danmuAggs []aggRow
	if err := db.Raw(
		"SELECT uid, COUNT(*) AS total FROM " + danmuTable + " WHERE deleted_at IS NULL GROUP BY uid",
	).Scan(&danmuAggs).Error; err != nil {
		return err
	}
	danmuByUID := make(map[int64]int64, len(danmuAggs))
	for _, r := range danmuAggs {
		danmuByUID[r.UID] = r.Total
	}
	var giftAggs []aggRow
	if err := db.Raw(
		"SELECT uid, COALESCE(SUM(num * price), 0) AS total FROM " + giftTable + " WHERE deleted_at IS NULL GROUP BY uid",
	).Scan(&giftAggs).Error; err != nil {
		return err
	}
	giftByUID := make(map[int64]int64, len(giftAggs))
	for _, r := range giftAggs {
		giftByUID[r.UID] = r.Total
	}
	// 按主键游标分批取待回填行
	type pendingRow struct {
		ID  int64
		UID int64
	}
	var lastID, updated int64
	for {
		var rows []pendingRow
		if err := db.Model(&model.LiveUser{}).
			Select("id", "uid").
			Where("total_danmu_count = 0 AND total_gift_amount = 0 AND id > ?", lastID).
			Order("id ASC").
			Limit(liveUserBackfillBatchSize).
			Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			if row.ID > lastID {
				lastID = row.ID
			}
			danmu := danmuByUID[row.UID] // map 查不到的 uid 计为 0
			gift := giftByUID[row.UID]
			if danmu == 0 && gift == 0 {
				continue // 真实历史就是 (0,0)，写回无意义（下次启动仍会被重扫，属预期成本）
			}
			// 走原生 SQL 绕开 model 的 BeforeUpdate 钩子，与 sign_date 回填先例一致
			if err := db.Exec(
				"UPDATE "+liveUserTable+" SET total_danmu_count = ?, total_gift_amount = ? WHERE id = ?",
				danmu, gift, row.ID,
			).Error; err != nil {
				return err
			}
			updated++
		}
	}
	if updated > 0 {
		log.Printf("[migrate] live_users 回填累计弹幕数/礼物金额完成，共 %d 条", updated)
	}
	return nil
}
