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

	// ExistsByUIDToday 判断指定UID今日是否已有签到记录
	ExistsByUIDToday(ctx context.Context, tx *gorm.DB, uid int64) (bool, error)
	// CountByUID 统计指定UID的总签到次数
	CountByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error)
	// StreakByUID 计算指定UID的连续签到天数
	StreakByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error)
}

// ExistsByUIDToday 判断指定UID今日是否已有签到记录
func (r *gormRepo) ExistsByUIDToday(ctx context.Context, tx *gorm.DB, uid int64) (bool, error) {
	db := r.getDB(ctx, tx)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).Unix()
	var count int64
	err := db.Model(&model.LiveUserSignLog{}).
		Where("uid = ? AND created_at >= ? AND created_at <= ?", uid, todayStart, todayEnd).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountByUID 统计指定UID的总签到次数
func (r *gormRepo) CountByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error) {
	db := r.getDB(ctx, tx)
	var count int64
	err := db.Model(&model.LiveUserSignLog{}).Where("uid = ?", uid).Count(&count).Error
	return count, err
}

// StreakByUID 计算指定UID的连续签到天数
func (r *gormRepo) StreakByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error) {
	db := r.getDB(ctx, tx)
	var logs []model.LiveUserSignLog
	if err := db.Model(&model.LiveUserSignLog{}).
		Where("uid = ?", uid).
		Order("created_at DESC").
		Limit(365).
		Find(&logs).Error; err != nil {
		return 0, err
	}
	if len(logs) == 0 {
		return 0, nil
	}
	var streak int64
	today := time.Now().Truncate(24 * time.Hour)
	for _, l := range logs {
		logDate := time.Unix(l.CreatedAt, 0).Truncate(24 * time.Hour)
		expectedDate := today.AddDate(0, 0, -int(streak))
		if logDate.Equal(expectedDate) {
			streak++
		} else if logDate.Before(expectedDate) {
			break
		}
	}
	return streak, nil
}
