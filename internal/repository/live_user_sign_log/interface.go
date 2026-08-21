package live_user_sign_log

import (
	"context"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

// Repository 接口定义
type Repository interface {
	base.Repository[model.LiveUserSignLog]

	// CountByUID 统计指定UID的总签到次数
	CountByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error)
	// StreakByUID 计算指定UID的连续签到天数
	StreakByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error)
}

// CountByUID 统计指定UID的总签到次数
//
// (uid, sign_date) 上有唯一索引，一天至多一条，所以行数即总签到天数
func (r *gormRepo) CountByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error) {
	db := r.getDB(ctx, tx)
	var count int64
	err := db.Model(&model.LiveUserSignLog{}).Where("uid = ?", uid).Count(&count).Error
	return count, err
}

// StreakByUID 计算指定UID的连续签到天数
//
// 以 sign_date 为准逐日回溯。sign_date 是写入时按服务器本地时区生成的 YYYY-MM-DD
func (r *gormRepo) StreakByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error) {
	db := r.getDB(ctx, tx)
	var dates []string
	if err := db.Model(&model.LiveUserSignLog{}).
		Where("uid = ? AND sign_date <> ''", uid).
		Order("sign_date DESC").
		Limit(365).
		Pluck("sign_date", &dates).Error; err != nil {
		return 0, err
	}
	if len(dates) == 0 {
		return 0, nil
	}
	// 从今天开始往前逐日比对，出现断档立即结束（今天未签到则为 0）
	today := time.Now()
	var streak int64
	for _, d := range dates {
		if d != today.AddDate(0, 0, -int(streak)).Format(time.DateOnly) {
			break
		}
		streak++
	}
	return streak, nil
}
