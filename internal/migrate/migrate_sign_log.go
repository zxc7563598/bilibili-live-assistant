package migrate

import (
	"log"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"gorm.io/gorm"
)

// signDateIndexName (uid, sign_date) 唯一索引名，与 model.LiveUserSignLog 上的 tag 保持一致
const signDateIndexName = "uk_uid_sign_date"

// signLogBackfillBatchSize 回填时每批读取的行数，避免大表一次性载入内存
const signLogBackfillBatchSize = 500

// backfillLiveUserSignLogSignDate 兼容历史数据：为 live_user_sign_logs 补充 sign_date 字段并按 created_at 的本地日期回填
//
// 表不存在（全新库）或唯一索引已建好（本函数已执行过）时直接跳过。
// 后续 AutoMigrate 会自动补齐 (uid, sign_date) 唯一索引
func backfillLiveUserSignLogSignDate(db *gorm.DB) error {
	m := db.Migrator()
	if !m.HasTable(&model.LiveUserSignLog{}) {
		return nil
	}
	if m.HasIndex(&model.LiveUserSignLog{}, signDateIndexName) {
		// 索引已存在说明 sign_date 早就回填完毕，无需每次启动都全表扫一遍
		return nil
	}
	table := model.LiveUserSignLog{}.TableName()
	if !m.HasColumn(&model.LiveUserSignLog{}, "SignDate") {
		if err := db.Exec("ALTER TABLE " + table + " ADD COLUMN sign_date varchar(10)").Error; err != nil {
			return err
		}
	}
	// 按主键游标分批回填缺失的 sign_date，批内按日期归并成一条 UPDATE 减少往返
	type signLogRow struct {
		ID        int64
		CreatedAt int64
	}
	var lastID, updated int64
	for {
		var rows []signLogRow
		if err := db.Table(table).
			Select("id", "created_at").
			Where("(sign_date = '' OR sign_date IS NULL) AND id > ?", lastID).
			Order("id ASC").
			Limit(signLogBackfillBatchSize).
			Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		idsByDate := make(map[string][]int64)
		for _, row := range rows {
			signDate := time.Unix(row.CreatedAt, 0).Format(time.DateOnly)
			idsByDate[signDate] = append(idsByDate[signDate], row.ID)
			if row.ID > lastID {
				lastID = row.ID
			}
		}
		for signDate, ids := range idsByDate {
			// 走原生 SQL 绕开 model 上禁止更新的 BeforeUpdate 钩子
			if err := db.Exec("UPDATE "+table+" SET sign_date = ? WHERE id IN ?", signDate, ids).Error; err != nil {
				return err
			}
		}
		updated += int64(len(rows))
	}
	if updated > 0 {
		log.Printf("[migrate] live_user_sign_logs 回填 sign_date 完成，共 %d 条", updated)
	}
	return nil
}

// dedupeLiveUserSignLogSignDate 按 (uid, sign_date) 去重：保留 id 最小的一条，删除其余，
// 避免后续 AutoMigrate 创建唯一索引 uk_uid_sign_date 时因重复数据失败
//
// 注意这是破坏性操作，会真实删除历史签到记录，升级前请先备份数据库
func dedupeLiveUserSignLogSignDate(db *gorm.DB) error {
	m := db.Migrator()
	if !m.HasTable(&model.LiveUserSignLog{}) {
		return nil
	}
	if m.HasIndex(&model.LiveUserSignLog{}, signDateIndexName) {
		// 唯一索引已存在，重复数据不可能产生，跳过全表 GROUP BY
		return nil
	}
	table := model.LiveUserSignLog{}.TableName()
	var groups []struct {
		UID      int64
		SignDate string
		KeepID   int64
	}
	if err := db.Raw(
		"SELECT uid, sign_date, MIN(id) AS keep_id FROM " + table +
			" WHERE sign_date <> '' GROUP BY uid, sign_date HAVING COUNT(*) > 1",
	).Scan(&groups).Error; err != nil {
		return err
	}
	var deleted int64
	for _, g := range groups {
		// 走原生 SQL 绕开 model 上禁止删除的 BeforeDelete 钩子
		res := db.Exec(
			"DELETE FROM "+table+" WHERE uid = ? AND sign_date = ? AND id <> ?",
			g.UID, g.SignDate, g.KeepID,
		)
		if res.Error != nil {
			return res.Error
		}
		deleted += res.RowsAffected
	}
	if deleted > 0 {
		log.Printf("[migrate] live_user_sign_logs 按 (uid, sign_date) 去重，涉及 %d 组、删除 %d 条重复记录（每组保留 id 最小的一条）", len(groups), deleted)
	}
	return nil
}
