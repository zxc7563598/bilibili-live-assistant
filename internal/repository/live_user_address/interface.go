package live_user_address

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	base.Repository[model.LiveUserAddress]
	// ListByUserID 根据用户ID获取收货地址；addressType 为 nil 时返回全部类型，否则仅返回该类型。默认地址排前（is_default desc, id asc）
	ListByUserID(ctx context.Context, tx *gorm.DB, userID int64, addressType *enum.AddressType) ([]model.LiveUserAddress, error)
	// GetDefaultByUserID 根据用户ID + 类型获取该类型的默认收货地址，不存在返回 nil
	GetDefaultByUserID(ctx context.Context, tx *gorm.DB, userID int64, addressType enum.AddressType) (*model.LiveUserAddress, error)
	// SetDefault 将指定地址置为默认（先清空该用户该类型下的默认，再置当前为默认）
	SetDefault(ctx context.Context, tx *gorm.DB, id, userID int64, addressType enum.AddressType) error
}

// ListByUserID 根据用户ID获取收货地址（可按类型过滤）
func (r *gormRepo) ListByUserID(ctx context.Context, tx *gorm.DB, userID int64, addressType *enum.AddressType) ([]model.LiveUserAddress, error) {
	db := r.getDB(ctx, tx)
	query := db.Where("user_id = ?", userID)
	if addressType != nil {
		query = query.Where("type = ?", *addressType)
	}
	var list []model.LiveUserAddress
	err := query.Order("is_default desc, id asc").Find(&list).Error
	return list, err
}

// GetDefaultByUserID 根据用户ID + 类型获取默认收货地址
func (r *gormRepo) GetDefaultByUserID(ctx context.Context, tx *gorm.DB, userID int64, addressType enum.AddressType) (*model.LiveUserAddress, error) {
	db := r.getDB(ctx, tx)
	var entity model.LiveUserAddress
	err := db.Where("user_id = ? AND type = ?", userID, addressType).Where("is_default = ?", enum.Yes).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// SetDefault 将指定地址置为默认（先清空该用户该类型下的默认，再置当前为默认）
// 通过 SELECT ... FOR UPDATE 锁定该用户该类型下的现有地址行，串行化并发的默认切换，避免产生两个默认。
// 注：对同一用户+类型完全不存在任何地址时并发首建两条默认地址仍存在极小竞争窗口，若需杜绝需数据库层唯一约束兜底。
func (r *gormRepo) SetDefault(ctx context.Context, tx *gorm.DB, id, userID int64, addressType enum.AddressType) error {
	apply := func(db *gorm.DB) error {
		var locked []int64
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Model(&model.LiveUserAddress{}).
			Where("user_id = ? AND type = ?", userID, addressType).
			Pluck("id", &locked).Error; err != nil {
			return err
		}
		if err := db.Model(&model.LiveUserAddress{}).Where("user_id = ? AND type = ?", userID, addressType).Update("is_default", enum.No).Error; err != nil {
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
