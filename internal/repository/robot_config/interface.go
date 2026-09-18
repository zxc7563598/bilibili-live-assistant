package robot_config

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.RobotConfig]

	// UpdateValueByID 根据 ID 更新配置值（只写 config_value 一列）
	UpdateValueByID(ctx context.Context, tx *gorm.DB, id int64, configValue string) error
}

// UpdateValueByID 根据 ID 更新配置值（只写 config_value 一列）
func (r *gormRepo) UpdateValueByID(ctx context.Context, tx *gorm.DB, id int64, configValue string) error {
	return r.UpdateField(ctx, tx, id, "config_value", configValue)
}
