package role_menu

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.RoleMenu]
	// ListByRoleID 根据角色ID获取权限信息
	ListByRoleID(ctx context.Context, tx *gorm.DB, roleID int64) ([]model.RoleMenu, error)
	// DeleteByRoleID 物理删除该角色的全部菜单授权（可命中多行）
	DeleteByRoleID(ctx context.Context, tx *gorm.DB, roleID int64) error
	// ListByRoleIDs 根据角色ID批量获取
	ListByRoleIDs(ctx context.Context, tx *gorm.DB, ids []int64) ([]model.RoleMenu, error)
}

// ListByRoleID 根据角色ID获取权限信息
func (r *gormRepo) ListByRoleID(ctx context.Context, tx *gorm.DB, roleID int64) ([]model.RoleMenu, error) {
	return r.ListByField(ctx, tx, "role_id", roleID)
}

// DeleteByRoleID 物理删除该角色的全部菜单授权（可命中多行）
//
// 关联表不做软删除，理由同 admin_role：role_menus 目前没有唯一索引，
// 软删除不会报错，但重新授权走的是「先删后插」（见 service/role 的
// resetRoleMenus），死行只会在表里越堆越多，且没有保留价值。
func (r *gormRepo) DeleteByRoleID(ctx context.Context, tx *gorm.DB, roleID int64) error {
	db := r.ResolveDB(ctx, tx)
	return db.Unscoped().Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error
}

// ListByRoleIDs 根据角色ID批量获取
func (r *gormRepo) ListByRoleIDs(ctx context.Context, tx *gorm.DB, ids []int64) ([]model.RoleMenu, error) {
	db := r.ResolveDB(ctx, tx)
	var list []model.RoleMenu
	if err := db.Where("role_id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
