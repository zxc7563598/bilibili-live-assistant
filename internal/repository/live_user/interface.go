package live_user

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.LiveUser]
	// GetByUID 根据 B站 UID 查询单条用户记录
	GetByUID(ctx context.Context, tx *gorm.DB, uid int64) (*model.LiveUser, error)
	// ListPage 分页查询用户，UID 精确匹配，Uname 模糊匹配，按 CreatedAt 倒序
	ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserListPageQuery) ([]model.LiveUser, int64, error)
}

// GetByUID 根据 B站 UID 查询单条用户记录
func (r *gormRepo) GetByUID(ctx context.Context, tx *gorm.DB, uid int64) (*model.LiveUser, error) {
	return r.FindOneByField(ctx, tx, "uid", uid)
}

// ListPage 分页查询用户
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserListPageQuery) ([]model.LiveUser, int64, error) {
	var list []model.LiveUser
	var total int64
	db := r.getDB(ctx, tx)
	db = db.Model(&model.LiveUser{})
	if v := query.UID; v != nil {
		db = db.Where("uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("uname LIKE ?", "%"+escapeLike(*v)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("created_at desc").Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	return list, total, err
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
