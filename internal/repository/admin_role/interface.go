package admin_role

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.AdminRole]
	// ExistsByAdminIDAndRoleID 判断管理员是否拥有指定角色
	ExistsByAdminIDAndRoleID(ctx context.Context, tx *gorm.DB, adminID, roleID int64) (bool, error)
	// DeleteByAdminID 物理删除该管理员的全部角色绑定（可命中多行）
	DeleteByAdminID(ctx context.Context, tx *gorm.DB, adminID int64) error
	// ListByAdminIDs 根据多个管理员ID获取全部相关角色
	ListByAdminIDs(ctx context.Context, tx *gorm.DB, adminIDs []int64) ([]model.AdminRole, error)
	// BindRoles 绑定管理员/角色
	BindRoles(ctx context.Context, tx *gorm.DB, adminIDs []int64, roleID int64) error
	// UnbindRoles 取消绑定管理员/角色（物理删除，同 DeleteByAdminID）
	UnbindRoles(ctx context.Context, tx *gorm.DB, adminIDs []int64, roleID int64) error
}

// ExistsByAdminIDAndRoleID 判断管理员是否拥有指定角色
func (r *gormRepo) ExistsByAdminIDAndRoleID(ctx context.Context, tx *gorm.DB, adminID, roleID int64) (bool, error) {
	db := r.ResolveDB(ctx, tx)
	var count int64
	if err := db.Model(&model.AdminRole{}).Where("admin_id = ? AND role_id = ?", adminID, roleID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteByAdminID 物理删除该管理员的全部角色绑定（可命中多行）
//
// 关联表不做软删除。软删除留下的行仍然占着 (admin_id, role_id) 唯一索引
// uk_admin_role，而绑定走的是「先删后插」（见 service/admin 的 bindRoles），
// 于是重新绑定同一组合时会撞唯一键报 gorm.ErrDuplicatedKey。故这里用
// Unscoped 真删。row 本身只是「某管理员拥有某角色」这个事实的载体，
// 删掉它不丢失任何需要保留的信息。
func (r *gormRepo) DeleteByAdminID(ctx context.Context, tx *gorm.DB, adminID int64) error {
	db := r.ResolveDB(ctx, tx)
	return db.Unscoped().Where("admin_id = ?", adminID).Delete(&model.AdminRole{}).Error
}

// ListByAdminIDs 根据多个管理员ID获取全部相关角色
func (r *gormRepo) ListByAdminIDs(ctx context.Context, tx *gorm.DB, adminIDs []int64) ([]model.AdminRole, error) {
	db := r.ResolveDB(ctx, tx)
	var list []model.AdminRole
	if err := db.Where("admin_id IN ?", adminIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// BindRoles 绑定管理员/角色
func (r *gormRepo) BindRoles(ctx context.Context, tx *gorm.DB, adminIDs []int64, roleID int64) error {
	entities := make([]model.AdminRole, 0, len(adminIDs))
	for _, v := range adminIDs {
		entities = append(entities, model.AdminRole{
			AdminID: v,
			RoleID:  roleID,
		})
	}
	return r.CreateBatch(ctx, tx, entities)
}

// UnbindRoles 取消绑定管理员/角色
//
// 同样物理删除，理由见 DeleteByAdminID：若留着软删除的行，
// 被取消过的组合就再也加不回来——service/role 的 AddRoleUsers 用
// ExistsByAdminIDAndRoleID 判断，查不到软删除的行，于是照样去 BindRoles 插入。
func (r *gormRepo) UnbindRoles(ctx context.Context, tx *gorm.DB, adminIDs []int64, roleID int64) error {
	if len(adminIDs) == 0 {
		return nil
	}
	db := r.ResolveDB(ctx, tx)
	return db.Unscoped().Where("admin_id IN ? AND role_id = ?", adminIDs, roleID).Delete(&model.AdminRole{}).Error
}
