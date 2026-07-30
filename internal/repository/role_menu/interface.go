package role_menu

import (
	"context"

	"github.com/zxc7563598/GoAdminKit/internal/model"
	"github.com/zxc7563598/GoAdminKit/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.RoleMenu]
	ListByRoleID(ctx context.Context, tx *gorm.DB, roleID uint64) ([]model.RoleMenu, error)
	DeleteByRoleID(ctx context.Context, tx *gorm.DB, roleID uint64) error
	ListByRoleIDs(ctx context.Context, tx *gorm.DB, ids []uint64) ([]model.RoleMenu, error)
}

// ListByRoleID 根据角色ID获取权限信息
func (r *gormRepo) ListByRoleID(ctx context.Context, tx *gorm.DB, roleID uint64) ([]model.RoleMenu, error) {
	return r.FindByField(ctx, tx, "role_id", roleID)
}

// DeleteByRoleID 删除角色ID相关的权限
func (r *gormRepo) DeleteByRoleID(ctx context.Context, tx *gorm.DB, roleID uint64) error {
	db := r.getDB(ctx, tx)
	return db.Where("role_id = ?", roleID).Delete(&model.RoleMenu{}).Error
}

// ListByRoleIDs 根据角色ID批量获取
func (r *gormRepo) ListByRoleIDs(ctx context.Context, tx *gorm.DB, ids []uint64) ([]model.RoleMenu, error) {
	db := r.getDB(ctx, tx)
	var list []model.RoleMenu
	if err := db.Where("role_id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
