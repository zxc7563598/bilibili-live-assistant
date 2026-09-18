package live_user_blacklist

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

// Repository 接口定义
type Repository interface {
	base.Repository[model.LiveUserBlacklist]
	// GetActiveByRoomUID 获取用户在指定房间内禁言中且未过期的黑名单记录
	// now 由调用方传入当前时间戳（秒级），与 ListExpiredMuted 保持同一约定，便于测试
	GetActiveByRoomUID(ctx context.Context, tx *gorm.DB, roomID, uid, now int64) (*model.LiveUserBlacklist, error)
	// UpdateUnmuteResult 根据黑名单ID更新解禁结果
	UpdateUnmuteResult(ctx context.Context, tx *gorm.DB, id int64, status enum.MuteStatus, unmuteFailCount int64) error
	// ListExpiredMuted 查询已到解禁时间但仍处于禁言状态的黑名单记录
	ListExpiredMuted(ctx context.Context, tx *gorm.DB, now int64) ([]model.LiveUserBlacklist, error)
}

// GetActiveByRoomUID 获取用户在指定房间内禁言中且未过期的黑名单记录
// 多条匹配时取 CreatedAt 最新的一条，不存在返回 nil
func (r *gormRepo) GetActiveByRoomUID(ctx context.Context, tx *gorm.DB, roomID, uid, now int64) (*model.LiveUserBlacklist, error) {
	db := r.ResolveDB(ctx, tx)
	var entity model.LiveUserBlacklist
	err := db.Where("room_id = ?", roomID).
		Where("uid = ?", uid).
		Where("status = ?", enum.MuteStatusMuted).
		Where("mute_expires_at > ?", now).
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
	db := r.ResolveDB(ctx, tx)
	return db.Model(&model.LiveUserBlacklist{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":            status,
			"unmute_fail_count": unmuteFailCount,
		}).Error
}

// ListExpiredMuted 查询已到解禁时间但仍处于禁言状态的黑名单记录
// 按解禁时间升序排列，最早到期的优先处理
func (r *gormRepo) ListExpiredMuted(ctx context.Context, tx *gorm.DB, now int64) ([]model.LiveUserBlacklist, error) {
	db := r.ResolveDB(ctx, tx)
	var list []model.LiveUserBlacklist
	err := db.Where("status = ?", enum.MuteStatusMuted).
		Where("mute_expires_at <= ?", now).
		Order("mute_expires_at asc").
		Find(&list).Error
	return list, err
}
