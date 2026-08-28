package live_user_address

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type Repository interface {
	base.Repository[model.LiveUserAddress]
	// ListByUserID 根据用户ID获取全部收货地址，按 ID 升序
	ListByUserID(ctx context.Context, tx *gorm.DB, userID int64) ([]model.LiveUserAddress, error)
	// GetDefaultByUserID 根据用户ID获取默认收货地址，不存在返回 nil
	GetDefaultByUserID(ctx context.Context, tx *gorm.DB, userID int64) (*model.LiveUserAddress, error)
	// SetDefault 将指定地址置为默认（先清空该用户全部默认，再置当前为默认）
	SetDefault(ctx context.Context, tx *gorm.DB, id, userID int64) error
}

// ListByUserID 根据用户ID获取全部收货地址
func (r *gormRepo) ListByUserID(ctx context.Context, tx *gorm.DB, userID int64) ([]model.LiveUserAddress, error) {
	db := r.getDB(ctx, tx)
	var list []model.LiveUserAddress
	err := db.Where("user_id = ?", userID).Order("id asc").Find(&list).Error
	return list, err
}

// GetDefaultByUserID 根据用户ID获取默认收货地址
func (r *gormRepo) GetDefaultByUserID(ctx context.Context, tx *gorm.DB, userID int64) (*model.LiveUserAddress, error) {
	db := r.getDB(ctx, tx)
	var entity model.LiveUserAddress
	err := db.Where("user_id = ?", userID).Where("is_default = ?", enum.Yes).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// SetDefault 将指定地址置为默认
func (r *gormRepo) SetDefault(ctx context.Context, tx *gorm.DB, id, userID int64) error {
	apply := func(db *gorm.DB) error {
		if err := db.Model(&model.LiveUserAddress{}).Where("user_id = ?", userID).Update("is_default", enum.No).Error; err != nil {
			return err
		}
		return db.Model(&model.LiveUserAddress{}).Where("id = ? AND user_id = ?", id, userID).Update("is_default", enum.Yes).Error
	}
	if tx != nil {
		return apply(r.getDB(ctx, tx))
	}
	db := r.db
	if ctx != nil {
		db = db.WithContext(ctx)
	}
	return db.Transaction(apply)
}
