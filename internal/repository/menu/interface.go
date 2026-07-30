package menu

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.Menu]
	ListEnabled(ctx context.Context, tx *gorm.DB) ([]model.Menu, error)
	ListEnabledByIDs(ctx context.Context, tx *gorm.DB, ids []uint64) ([]model.Menu, error)
	ExistsByPath(ctx context.Context, tx *gorm.DB, path string) (bool, error)
	ListButtonsByParentID(ctx context.Context, tx *gorm.DB, parentID uint64) ([]model.Menu, error)
	UpdateByID(ctx context.Context, tx *gorm.DB, id uint64, form model.MenuUpdateByIdForm) error
	UpdateEnableByID(ctx context.Context, tx *gorm.DB, id uint64) error
}

// ListEnabled 获取全部启用菜单
func (r *gormRepo) ListEnabled(ctx context.Context, tx *gorm.DB) ([]model.Menu, error) {
	return r.FindByField(ctx, tx, "enable", enum.EnableEnable)
}

// ListEnabledByIDs 根据ID获取全部菜单
func (r *gormRepo) ListEnabledByIDs(ctx context.Context, tx *gorm.DB, ids []uint64) ([]model.Menu, error) {
	db := r.getDB(ctx, tx)
	var list []model.Menu
	if err := db.Where("enable = ?", enum.EnableEnable).Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ExistsByPath 根据路径获取菜单是否存在
func (r *gormRepo) ExistsByPath(ctx context.Context, tx *gorm.DB, path string) (bool, error) {
	return r.Exists(ctx, tx, "path", path)
}

// ListButtonsByParentID 获取菜单下的按钮
func (r *gormRepo) ListButtonsByParentID(ctx context.Context, tx *gorm.DB, parentID uint64) ([]model.Menu, error) {
	db := r.getDB(ctx, tx)
	var list []model.Menu
	if err := db.Where("parent_id = ?", parentID).Where("type = ?", enum.MenuTypeButton).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// UpdateByID 变更菜单基本信息
func (r *gormRepo) UpdateByID(ctx context.Context, tx *gorm.DB, id uint64, form model.MenuUpdateByIdForm) error {
	return r.UpdateMap(ctx, tx, "id", id, map[string]any{
		"code":       form.Code,
		"enable":     form.Enable,
		"show":       form.Show,
		"keep_alive": form.KeepAlive,
		"layout":     form.Layout,
		"type":       form.Type,
		"parent_id":  form.ParentID,
		"name":       form.Name,
		"icon":       form.Icon,
		"path":       form.Path,
		"component":  form.Component,
		"order":      form.Order,
	})
}

// UpdateEnableByID 切换菜单启动状态
func (r *gormRepo) UpdateEnableByID(ctx context.Context, tx *gorm.DB, id uint64) error {
	db := r.getDB(ctx, tx)
	err := db.WithContext(ctx).
		Model(&model.Menu{}).
		Where("id = ?", id).
		Update("enable", gorm.Expr("CASE WHEN enable = ? THEN ? ELSE ? END",
			enum.EnableEnable,
			enum.EnableDisable,
			enum.EnableEnable,
		)).Error
	if err != nil {
		return err
	}
	return nil
}
