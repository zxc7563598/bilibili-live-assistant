package robot_config

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.RobotConfig]

	// UpdateByID 根据 ID 部分更新配置
	UpdateByID(ctx context.Context, tx *gorm.DB, id int64, configValue string) error
}

// UpdateByID 根据 ID 部分更新配置
func (r *gormRepo) UpdateByID(ctx context.Context, tx *gorm.DB, id int64, configValue string) error {
	return r.UpdateField(ctx, tx, id, "config_value", configValue)
}
