package role_menu

import (
	"context"

	"github.com/zxc7563598/GoAdminKit/internal/model"
	"github.com/zxc7563598/GoAdminKit/internal/repository/base"
	"gorm.io/gorm"
)

type gormRepo struct {
	db *gorm.DB
	base.Repository[model.RoleMenu]
}

func New(db *gorm.DB) Repository {
	return &gormRepo{
		db:         db,
		Repository: base.New[model.RoleMenu](db),
	}
}

// getDB 封装 ctx + tx
func (r *gormRepo) getDB(ctx context.Context, tx *gorm.DB) *gorm.DB {
	db := r.db
	if tx != nil {
		db = tx
	}
	if ctx != nil {
		db = db.WithContext(ctx)
	}
	return db
}
