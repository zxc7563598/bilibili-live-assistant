package live_user

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	base.Repository[model.LiveUser]
	// GetByUID 根据 B站 UID 查询单条用户记录
	GetByUID(ctx context.Context, tx *gorm.DB, uid int64) (*model.LiveUser, error)
	// UpdateName 根据 ID 变更用户昵称
	UpdateName(ctx context.Context, tx *gorm.DB, id int64, uname string) error
	// CreateIfNotExist 若 uid 已存在则忽略创建并返回已有记录，否则创建新记录
	CreateIfNotExist(ctx context.Context, tx *gorm.DB, entity *model.LiveUser) (*model.LiveUser, error)
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

// UpdateName 根据 ID 变更用户昵称
func (r *gormRepo) UpdateName(ctx context.Context, tx *gorm.DB, id int64, uname string) error {
	return r.UpdateField(ctx, tx, id, "uname", uname)
}

// CreateIfNotExist 若 uid 已存在则忽略创建并返回已有记录，否则创建新记录
func (r *gormRepo) CreateIfNotExist(ctx context.Context, tx *gorm.DB, entity *model.LiveUser) (*model.LiveUser, error) {
	db := r.getDB(ctx, tx)
	res := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "uid"}},
		DoNothing: true,
	}).Create(entity)
	if res.Error != nil {
		return nil, res.Error
	}
	// 冲突未插入（RowsAffected == 0），返回已存在的记录
	if res.RowsAffected == 0 {
		return r.GetByUID(ctx, tx, entity.Uid)
	}
	return entity, nil
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
