package admin_role

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.AdminRole]
	ExistsByAdminIDAndRoleID(ctx context.Context, tx *gorm.DB, adminID, roleID int64) (bool, error)
	DeleteByAdminID(ctx context.Context, tx *gorm.DB, adminID int64) error
	ListByAdminIDs(ctx context.Context, tx *gorm.DB, adminIDs []int64) ([]model.AdminRole, error)
	BindRoles(ctx context.Context, tx *gorm.DB, adminID []int64, roleID int64) error
	UnbindRoles(ctx context.Context, tx *gorm.DB, adminIDs []int64, roleID int64) error
}

// ExistsByAdminIDAndRoleID 判断管理员是否拥有指定角色
func (r *gormRepo) ExistsByAdminIDAndRoleID(ctx context.Context, tx *gorm.DB, adminID, roleID int64) (bool, error) {
	db := r.getDB(ctx, tx)
	var count int64
	if err := db.Model(&model.AdminRole{}).Where("admin_id = ? AND role_id = ?", adminID, roleID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteByAdminID 根据管理员ID删除绑定记录
func (r *gormRepo) DeleteByAdminID(ctx context.Context, tx *gorm.DB, adminID int64) error {
	db := r.getDB(ctx, tx)
	return db.Where("admin_id = ?", adminID).Delete(&model.AdminRole{}).Error
}

// ListByAdminIDs 根据多个管理员ID获取全部相关角色
func (r *gormRepo) ListByAdminIDs(ctx context.Context, tx *gorm.DB, adminIDs []int64) ([]model.AdminRole, error) {
	db := r.getDB(ctx, tx)
	var list []model.AdminRole
	if err := db.Where("admin_id IN ?", adminIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// BindRoles 绑定管理员/角色
func (r *gormRepo) BindRoles(ctx context.Context, tx *gorm.DB, adminID []int64, roleID int64) error {
	entities := make([]model.AdminRole, 0, len(adminID))
	for _, v := range adminID {
		entities = append(entities, model.AdminRole{
			AdminID: v,
			RoleID:  roleID,
		})
	}
	return r.CreateBatch(ctx, tx, entities)
}

// UnbindRoles 取消绑定管理员/角色
func (r *gormRepo) UnbindRoles(ctx context.Context, tx *gorm.DB, adminIDs []int64, roleID int64) error {
	if len(adminIDs) == 0 {
		return nil
	}
	db := r.getDB(ctx, tx)
	return db.Where("admin_id IN ? AND role_id = ?", adminIDs, roleID).Delete(&model.AdminRole{}).Error
}
