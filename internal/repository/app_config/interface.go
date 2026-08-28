package app_config

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.AppConfig]
	// GetAll 获取全部配置，按 ID 升序
	GetAll(ctx context.Context, tx *gorm.DB) ([]model.AppConfig, error)
	// GetByKey 根据配置键获取单条配置，不存在返回 nil
	GetByKey(ctx context.Context, tx *gorm.DB, key string) (*model.AppConfig, error)
}

// GetAll 获取全部配置
func (r *gormRepo) GetAll(ctx context.Context, tx *gorm.DB) ([]model.AppConfig, error) {
	db := r.getDB(ctx, tx)
	var list []model.AppConfig
	err := db.Order("id asc").Find(&list).Error
	return list, err
}

// GetByKey 根据配置键获取单条配置
func (r *gormRepo) GetByKey(ctx context.Context, tx *gorm.DB, key string) (*model.AppConfig, error) {
	return r.FindOneByField(ctx, tx, "config_key", key)
}
