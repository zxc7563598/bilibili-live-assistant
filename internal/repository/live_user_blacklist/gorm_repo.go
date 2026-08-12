package live_user_blacklist

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type gormRepo struct {
	db *gorm.DB
	base.Repository[model.LiveUserBlacklist]
}

func New(db *gorm.DB) Repository {
	return &gormRepo{
		db:         db,
		Repository: base.New[model.LiveUserBlacklist](db),
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
