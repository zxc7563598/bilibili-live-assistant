package admin

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.Admin]
	// GetByUsername 根据 username 获取账号信息
	GetByUsername(ctx context.Context, tx *gorm.DB, username string) (*model.Admin, error)
	// UpdateTokenByID 根据 id 更换管理员 refreshToken
	UpdateTokenByID(ctx context.Context, tx *gorm.DB, id int64, token *string) error
	// UpdatePasswordByID 更新管理员密码
	UpdatePasswordByID(ctx context.Context, tx *gorm.DB, adminID int64, password string) error
	// UpdateBasicInfoByID 更新管理员基本信息
	UpdateBasicInfoByID(ctx context.Context, tx *gorm.DB, adminID int64, form model.AdminUpdateBasicInfoByIdForm) error
	// UpdateRoleIDByID 更新管理员角色信息
	UpdateRoleIDByID(ctx context.Context, tx *gorm.DB, adminID int64, roleID int64) error
	// ListPage 获取分页列表数据
	ListPage(ctx context.Context, tx *gorm.DB, query model.AdminListPageQuery) ([]model.AdminListItem, int64, error)
	// UpdateProfileByID 更新管理员个人资料
	UpdateProfileByID(ctx context.Context, tx *gorm.DB, adminID int64, form model.AdminUpdateProfileByIdForm) error
}

// GetByUsername 根据 username 获取账号信息
func (r *gormRepo) GetByUsername(ctx context.Context, tx *gorm.DB, username string) (*model.Admin, error) {
	return r.FindOneByField(ctx, tx, "username", username)
}

// UpdateTokenByID 根据 id 更换管理员 refreshToken
func (r *gormRepo) UpdateTokenByID(ctx context.Context, tx *gorm.DB, id int64, token *string) error {
	return r.UpdateField(ctx, tx, id, "token", token)
}

// UpdatePasswordByID 更新管理员密码
func (r *gormRepo) UpdatePasswordByID(ctx context.Context, tx *gorm.DB, id int64, password string) error {
	return r.UpdateField(ctx, tx, id, "password", password)
}

// UpdateRoleIDByID 更新管理员角色信息
func (r *gormRepo) UpdateRoleIDByID(ctx context.Context, tx *gorm.DB, id int64, roleID int64) error {
	return r.UpdateField(ctx, tx, id, "role_id", roleID)
}

// UpdateBasicInfoByID 更新管理员基本信息
func (r *gormRepo) UpdateBasicInfoByID(ctx context.Context, tx *gorm.DB, id int64, form model.AdminUpdateBasicInfoByIdForm) error {
	updateMap := make(map[string]any, 3)
	if v := form.Username; v != nil && *v != "" {
		updateMap["username"] = *v
	}
	if v := form.RoleID; v != nil && *v != 0 {
		updateMap["role_id"] = *v
	}
	if v := form.Enable; v != nil {
		e := enum.Enable(*v)
		if e.IsValid() {
			updateMap["enable"] = e
		}
	}
	if len(updateMap) == 0 {
		return nil
	}
	return r.UpdateMap(ctx, tx, "id", id, updateMap)
}

// ListPage 获取分页列表数据
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.AdminListPageQuery) ([]model.AdminListItem, int64, error) {
	var list []model.AdminListItem
	var total int64
	db := r.getDB(ctx, tx)
	db = db.Model(&model.Admin{})
	if v := query.Username; v != nil && *v != "" {
		escaped := escapeLike(*v)
		db = db.Where("username LIKE ?", "%"+escaped+"%")
	}
	if v := query.Gender; v != nil {
		g := enum.Gender(*v)
		if g.IsValid() {
			db = db.Where("gender = ?", g)
		}
	}
	if v := query.Enable; v != nil {
		e := enum.Enable(*v)
		if e.IsValid() {
			db = db.Where("enable = ?", e)
		}
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("created_at desc").Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UpdateProfileByID 更新管理员个人资料
func (r *gormRepo) UpdateProfileByID(ctx context.Context, tx *gorm.DB, id int64, form model.AdminUpdateProfileByIdForm) error {
	updateMap := make(map[string]any, 5)
	if v := form.Nickname; v != nil && *v != "" {
		updateMap["nickname"] = *v
	}
	if v := form.Email; v != nil && *v != "" {
		updateMap["email"] = *v
	}
	if v := form.Address; v != nil && *v != "" {
		updateMap["address"] = *v
	}
	if v := form.Avatar; v != nil && *v != "" {
		updateMap["avatar"] = *v
	}
	if v := form.Gender; v != nil {
		g := enum.Gender(*v)
		if g.IsValid() {
			updateMap["gender"] = g
		}
	}
	if len(updateMap) == 0 {
		return nil
	}
	return r.UpdateMap(ctx, tx, "id", id, updateMap)
}

// escapeLike 转义 LIKE 查询中的特殊字符 _ %
func escapeLike(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '%' || r == '_' || r == '\\' {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
