package app_config

import (
	"context"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	base.Repository[model.AppConfig]
	// GetAll 获取全部配置，按 ID 升序
	GetAll(ctx context.Context, tx *gorm.DB) ([]model.AppConfig, error)
	// GetByKey 根据配置键获取单条配置，不存在返回 nil
	GetByKey(ctx context.Context, tx *gorm.DB, key string) (*model.AppConfig, error)
	// SaveValues 按 config_key 幂等批量保存配置值：已存在则更新 config_value，不存在则插入
	SaveValues(ctx context.Context, tx *gorm.DB, values map[string]string) error
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

// SaveValues 按 config_key 幂等批量保存配置值
func (r *gormRepo) SaveValues(ctx context.Context, tx *gorm.DB, values map[string]string) error {
	if len(values) == 0 {
		return nil
	}
	rows := make([]model.AppConfig, 0, len(values))
	for key, value := range values {
		rows = append(rows, model.AppConfig{
			ConfigKey:   key,
			ConfigValue: value,
		})
	}
	db := r.getDB(ctx, tx)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "config_key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"config_value": gorm.Expr("excluded.config_value"),
			"updated_at":   time.Now().Unix(),
		}),
	}).Create(&rows).Error
}
