package live_interact_word

import (
	"context"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

// Repository 直播间互动记录数据访问接口
type Repository interface {
	base.Repository[model.LiveInteractWord]

	// CountEnterByUIDAndRoomID 统计指定 UID 进入指定房间的累计次数
	CountEnterByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error)
	// EnterStreakDaysByUIDAndRoomID 计算指定 UID 进入指定房间的连续天数
	EnterStreakDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error)
	// EnterTotalDaysByUIDAndRoomID 统计指定 UID 进入指定房间的累计天数
	EnterTotalDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error)

	// CountFollowByUIDAndRoomID 统计指定 UID 关注指定房间的累计次数
	CountFollowByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error)
	// FollowStreakDaysByUIDAndRoomID 计算指定 UID 关注指定房间的连续天数
	FollowStreakDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error)
	// FollowTotalDaysByUIDAndRoomID 统计指定 UID 关注指定房间的累计天数
	FollowTotalDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error)

	// CountShareByUIDAndRoomID 统计指定 UID 分享指定房间的累计次数
	CountShareByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error)
	// ShareStreakDaysByUIDAndRoomID 计算指定 UID 分享指定房间的连续天数
	ShareStreakDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error)
	// ShareTotalDaysByUIDAndRoomID 统计指定 UID 分享指定房间的累计天数
	ShareTotalDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error)
}

// CountEnterByUIDAndRoomID 统计指定 UID 进入指定房间的累计次数
func (r *gormRepo) CountEnterByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error) {
	return r.countByType(ctx, tx, uid, roomID, enum.InteractTypeEnter)
}

// EnterStreakDaysByUIDAndRoomID 计算指定 UID 进入指定房间的连续天数
func (r *gormRepo) EnterStreakDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error) {
	days, err := r.listDistinctDaysByType(ctx, tx, uid, roomID, enum.InteractTypeEnter)
	if err != nil {
		return 0, err
	}
	return streakDays(days), nil
}

// EnterTotalDaysByUIDAndRoomID 统计指定 UID 进入指定房间的累计天数
func (r *gormRepo) EnterTotalDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error) {
	days, err := r.listDistinctDaysByType(ctx, tx, uid, roomID, enum.InteractTypeEnter)
	if err != nil {
		return 0, err
	}
	return int64(len(days)), nil
}

// CountFollowByUIDAndRoomID 统计指定 UID 关注指定房间的累计次数
func (r *gormRepo) CountFollowByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error) {
	return r.countByType(ctx, tx, uid, roomID, enum.InteractTypeFollow)
}

// FollowStreakDaysByUIDAndRoomID 计算指定 UID 关注指定房间的连续天数
func (r *gormRepo) FollowStreakDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error) {
	days, err := r.listDistinctDaysByType(ctx, tx, uid, roomID, enum.InteractTypeFollow)
	if err != nil {
		return 0, err
	}
	return streakDays(days), nil
}

// FollowTotalDaysByUIDAndRoomID 统计指定 UID 关注指定房间的累计天数
func (r *gormRepo) FollowTotalDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error) {
	days, err := r.listDistinctDaysByType(ctx, tx, uid, roomID, enum.InteractTypeFollow)
	if err != nil {
		return 0, err
	}
	return int64(len(days)), nil
}

// CountShareByUIDAndRoomID 统计指定 UID 分享指定房间的累计次数
func (r *gormRepo) CountShareByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error) {
	return r.countByType(ctx, tx, uid, roomID, enum.InteractTypeShare)
}

// ShareStreakDaysByUIDAndRoomID 计算指定 UID 分享指定房间的连续天数
func (r *gormRepo) ShareStreakDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error) {
	days, err := r.listDistinctDaysByType(ctx, tx, uid, roomID, enum.InteractTypeShare)
	if err != nil {
		return 0, err
	}
	return streakDays(days), nil
}

// ShareTotalDaysByUIDAndRoomID 统计指定 UID 分享指定房间的累计天数
func (r *gormRepo) ShareTotalDaysByUIDAndRoomID(ctx context.Context, tx *gorm.DB, uid, roomID int64) (int64, error) {
	days, err := r.listDistinctDaysByType(ctx, tx, uid, roomID, enum.InteractTypeShare)
	if err != nil {
		return 0, err
	}
	return int64(len(days)), nil
}

// countByType 统计指定 UID 在指定房间内指定互动类型的累计次数
func (r *gormRepo) countByType(ctx context.Context, tx *gorm.DB, uid, roomID int64, msgType enum.InteractType) (int64, error) {
	db := r.ResolveDB(ctx, tx)
	var count int64
	err := db.Model(&model.LiveInteractWord{}).
		Where("uid = ? AND room_id = ? AND msg_type = ?", uid, roomID, msgType).
		Count(&count).Error
	return count, err
}

// listDistinctDaysByType 返回指定 UID 在指定房间内指定互动类型的去重日期，按时间倒序。
// 日期以本地自然日零点的时间戳表示，天数在 Go 侧去重，避免依赖 MySQL/PostgreSQL 各自的日期函数。
func (r *gormRepo) listDistinctDaysByType(ctx context.Context, tx *gorm.DB, uid, roomID int64, msgType enum.InteractType) ([]int64, error) {
	db := r.ResolveDB(ctx, tx)
	var timestamps []int64
	if err := db.Model(&model.LiveInteractWord{}).
		Where("uid = ? AND room_id = ? AND msg_type = ?", uid, roomID, msgType).
		Order("timestamp desc").
		Pluck("timestamp", &timestamps).Error; err != nil {
		return nil, err
	}
	return distinctDays(timestamps), nil
}

// localDayStart 返回该时间戳所在「本地自然日」的零点时间戳
func localDayStart(ts int64) int64 {
	t := time.Unix(ts, 0).In(time.Local)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Unix()
}

// distinctDays 将时间戳列表去重为「本地自然日零点」的日期序列，保持时间倒序
func distinctDays(timestamps []int64) []int64 {
	days := make([]int64, 0, len(timestamps))
	seen := make(map[int64]struct{}, len(timestamps))
	for _, ts := range timestamps {
		day := localDayStart(ts)
		if _, ok := seen[day]; ok {
			continue
		}
		seen[day] = struct{}{}
		days = append(days, day)
	}
	return days
}

// streakDays 根据倒序的去重日期列表计算截至今天的连续天数。
// 按本地自然日比较，与 live_danmu.CountDailyByUID、live_session.DistinctLiveDays、
// live_user_sign_log.StreakDaysByUID 的口径保持一致。
func streakDays(days []int64) int64 {
	if len(days) == 0 {
		return 0
	}
	today := localDayStart(time.Now().Unix())
	var streak int64
	for i, d := range days {
		// 今天往前推 i 天的本地零点；再归一化一次，避免夏令时切换带来的小时偏移
		expected := localDayStart(time.Unix(today, 0).In(time.Local).AddDate(0, 0, -i).Unix())
		if d == expected {
			streak++
		} else if d < expected {
			break
		}
	}
	return streak
}
