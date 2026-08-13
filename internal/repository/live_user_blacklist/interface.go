package live_user_blacklist

import (
	"context"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

// Repository 接口定义
type Repository interface {
	base.Repository[model.LiveUserBlacklist]
	// GetActiveByRoomUID 获取用户在指定房间内禁言中且未过期的黑名单记录
	GetActiveByRoomUID(ctx context.Context, tx *gorm.DB, roomID, uid int64) (*model.LiveUserBlacklist, error)
	// UpdateUnmuteResult 根据黑名单ID更新解禁结果
	UpdateUnmuteResult(ctx context.Context, tx *gorm.DB, id int64, status enum.MuteStatus, unmuteFailCount int64) error
}

// GetActiveByRoomUID 获取用户在指定房间内禁言中且未过期的黑名单记录
// 多条匹配时取 CreatedAt 最新的一条，不存在返回 nil
func (r *gormRepo) GetActiveByRoomUID(ctx context.Context, tx *gorm.DB, roomID, uid int64) (*model.LiveUserBlacklist, error) {
	db := r.getDB(ctx, tx)
	var entity model.LiveUserBlacklist
	err := db.Where("room_id = ?", roomID).
		Where("uid = ?", uid).
		Where("status = ?", enum.MuteStatusMuted).
		Where("mute_expires_at > ?", time.Now().Unix()).
		Order("created_at desc").
		First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// UpdateUnmuteResult 根据黑名单ID更新解禁结果
func (r *gormRepo) UpdateUnmuteResult(ctx context.Context, tx *gorm.DB, id int64, status enum.MuteStatus, unmuteFailCount int64) error {
	db := r.getDB(ctx, tx)
	return db.Model(&model.LiveUserBlacklist{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":            status,
			"unmute_fail_count": unmuteFailCount,
		}).Error
}
