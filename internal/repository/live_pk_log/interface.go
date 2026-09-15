package live_pk_log

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

// Repository PK 对战记录数据访问接口
type Repository interface {
	base.Repository[model.LivePkLog]

	// GetByPkID 按 PK ID 查询记录，不存在返回 (nil, nil)
	GetByPkID(ctx context.Context, tx *gorm.DB, pkID int64) (*model.LivePkLog, error)
}

// GetByPkID 按 PK ID 查询记录，不存在返回 (nil, nil)
func (r *gormRepo) GetByPkID(ctx context.Context, tx *gorm.DB, pkID int64) (*model.LivePkLog, error) {
	return r.FindOneByField(ctx, tx, "pk_id", pkID)
}
