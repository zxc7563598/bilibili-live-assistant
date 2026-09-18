package role

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/sqlutil"
	"gorm.io/gorm"
)

// sortColumns 允许参与 ListPage 排序的 DB 列
var sortColumns = map[string]string{
	"id":         "id",
	"code":       "code",
	"name":       "name",
	"enable":     "enable",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

// sortOrder 允许的排序方向 → SQL 方向。
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

type Repository interface {
	base.Repository[model.Role]
	// GetByCode 根据 code 获取单条数据
	GetByCode(ctx context.Context, tx *gorm.DB, code string) (*model.Role, error)
	// ListEnabled 获取全部启用角色，按 ID 升序
	ListEnabled(ctx context.Context, tx *gorm.DB) ([]model.Role, error)
	// ListPage 获取分页列表数据
	ListPage(ctx context.Context, tx *gorm.DB, query model.RoleListPageQuery) ([]model.RoleListItem, int64, error)
	// UpdateByID 变更基本信息
	UpdateByID(ctx context.Context, tx *gorm.DB, id int64, form model.RoleUpdateByIdForm) error
}

// GetByCode 根据 code 获取单条数据
func (r *gormRepo) GetByCode(ctx context.Context, tx *gorm.DB, code string) (*model.Role, error) {
	return r.GetByField(ctx, tx, "code", code)
}

// ListEnabled 获取全部启用角色，按 ID 升序
func (r *gormRepo) ListEnabled(ctx context.Context, tx *gorm.DB) ([]model.Role, error) {
	db := r.ResolveDB(ctx, tx)
	var list []model.Role
	// Role 没有专门的排序列，固定按 ID 升序，避免返回顺序由数据库决定
	err := db.Where("enable = ?", enum.EnableEnable).Order("id asc").Find(&list).Error
	return list, err
}

// ListPage 获取分页列表数据
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.RoleListPageQuery) ([]model.RoleListItem, int64, error) {
	var list []model.RoleListItem
	var total int64
	db := r.ResolveDB(ctx, tx)
	db = db.Model(&model.Role{})
	if v := query.Name; v != nil && *v != "" {
		escaped := sqlutil.EscapeLike(*v)
		db = db.Where("name LIKE ? ESCAPE '!'", "%"+escaped+"%")
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
	// 默认按创建时间倒序（与全层其它 ListPage 一致）；排序参数非法/缺失时静默回退该默认
	orderClause := "created_at desc"
	if query.SortField != nil && query.SortOrder != nil {
		// 方向先小写归一化再查白名单，非法值直接回退默认
		if field, ok := sortColumns[*query.SortField]; ok {
			if dir, ok := sortOrder[strings.ToLower(*query.SortOrder)]; ok {
				// field/dir 均来自字面量白名单，杜绝注入；id asc 保证同键值时翻页稳定
				orderClause = field + " " + dir + ", id asc"
			}
		}
	}
	err := db.Order(orderClause).Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UpdateByID 变更基本信息
func (r *gormRepo) UpdateByID(ctx context.Context, tx *gorm.DB, id int64, form model.RoleUpdateByIdForm) error {
	updateMap := make(map[string]any)
	if v := form.Code; v != nil && *v != "" {
		updateMap["code"] = *v
	}
	if v := form.Name; v != nil && *v != "" {
		updateMap["name"] = *v
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
