package base

import (
	"context"

	"gorm.io/gorm"
)

// Repository 定义通用接口
type Repository[T any] interface {
	GetByID(ctx context.Context, tx *gorm.DB, id uint64) (*T, error)
	GetByIDs(ctx context.Context, tx *gorm.DB, ids []uint64) ([]T, error)
	FindAll(ctx context.Context, tx *gorm.DB) ([]T, error)
	FindByField(ctx context.Context, tx *gorm.DB, field string, value any) ([]T, error)
	FindOneByField(ctx context.Context, tx *gorm.DB, field string, value any) (*T, error)
	FindByCondition(ctx context.Context, tx *gorm.DB, cond map[string]any) ([]T, error)
	Create(ctx context.Context, tx *gorm.DB, entity *T) (*T, error)
	CreateBatch(ctx context.Context, tx *gorm.DB, entities []T) error
	Update(ctx context.Context, tx *gorm.DB, entity *T) error
	UpdateMap(ctx context.Context, tx *gorm.DB, field string, value any, updates map[string]any) error
	UpdateField(ctx context.Context, tx *gorm.DB, id uint64, field string, value any) error
	Delete(ctx context.Context, tx *gorm.DB, id uint64) error
	DeleteByIDs(ctx context.Context, tx *gorm.DB, ids []uint64) error
	Count(ctx context.Context, tx *gorm.DB) (int64, error)
	Exists(ctx context.Context, tx *gorm.DB, field string, value any) (bool, error)
}

// GetByID 根据主键查询记录。
// 如果记录不存在，返回 (nil, nil)。
func (r *gormRepo[T]) GetByID(ctx context.Context, tx *gorm.DB, id uint64) (*T, error) {
	db := r.getDB(ctx, tx)
	var entity T
	if err := db.First(&entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// GetByIDs 根据主键批量查询记录。
// 返回匹配的记录列表，如果没有匹配记录则返回空切片。
func (r *gormRepo[T]) GetByIDs(ctx context.Context, tx *gorm.DB, ids []uint64) ([]T, error) {
	db := r.getDB(ctx, tx)
	var list []T
	if err := db.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindAll 查询所有记录。
func (r *gormRepo[T]) FindAll(ctx context.Context, tx *gorm.DB) ([]T, error) {
	db := r.getDB(ctx, tx)
	var list []T
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindByField 根据指定字段查询记录列表。
func (r *gormRepo[T]) FindByField(ctx context.Context, tx *gorm.DB, field string, value any) ([]T, error) {
	db := r.getDB(ctx, tx)
	var list []T
	if err := db.Where(field+" = ?", value).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// FindOneByField 根据指定字段查询单条记录。
// 如果记录不存在，返回 (nil, nil)。
func (r *gormRepo[T]) FindOneByField(ctx context.Context, tx *gorm.DB, field string, value any) (*T, error) {
	db := r.getDB(ctx, tx)
	var entity T
	if err := db.Where(field+" = ?", value).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// FindByCondition 根据条件映射查询记录列表。
func (r *gormRepo[T]) FindByCondition(ctx context.Context, tx *gorm.DB, cond map[string]any) ([]T, error) {
	db := r.getDB(ctx, tx)
	var list []T
	if err := db.Where(cond).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// Create 创建一条记录，返回创建后的模型（包含ID等数据库生成的值）
func (r *gormRepo[T]) Create(ctx context.Context, tx *gorm.DB, entity *T) (*T, error) {
	db := r.getDB(ctx, tx)
	err := db.Create(entity).Error
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// CreateBatch 批量创建记录。
// 如果实体列表为空，则不执行任何操作。
func (r *gormRepo[T]) CreateBatch(ctx context.Context, tx *gorm.DB, entities []T) error {
	db := r.getDB(ctx, tx)
	if len(entities) == 0 {
		return nil
	}
	return db.Create(&entities).Error
}

// Update 更新一条记录（根据主键保存整个实体）。
func (r *gormRepo[T]) Update(ctx context.Context, tx *gorm.DB, entity *T) error {
	db := r.getDB(ctx, tx)
	return db.Save(entity).Error
}

// UpdateMap 使用map更新指定字段
func (r *gormRepo[T]) UpdateMap(ctx context.Context, tx *gorm.DB, field string, value any, updates map[string]any) error {
	db := r.getDB(ctx, tx)
	return db.Model(new(T)).Where(field+" = ?", value).Updates(updates).Error
}

// UpdateField 根据主键更新单个字段。
func (r *gormRepo[T]) UpdateField(ctx context.Context, tx *gorm.DB, id uint64, field string, value any) error {
	db := r.getDB(ctx, tx)
	return db.Model(new(T)).Where("id = ?", id).Update(field, value).Error
}

// Delete 根据主键删除记录。
func (r *gormRepo[T]) Delete(ctx context.Context, tx *gorm.DB, id uint64) error {
	db := r.getDB(ctx, tx)
	return db.Delete(new(T), id).Error
}

// DeleteByIDs 根据主键批量删除记录。
// 如果主键列表为空，则不执行任何操作。
func (r *gormRepo[T]) DeleteByIDs(ctx context.Context, tx *gorm.DB, ids []uint64) error {
	db := r.getDB(ctx, tx)
	if len(ids) == 0 {
		return nil
	}
	return db.Delete(new(T), ids).Error
}

// Count 统计记录总数。
func (r *gormRepo[T]) Count(ctx context.Context, tx *gorm.DB) (int64, error) {
	db := r.getDB(ctx, tx)
	var total int64
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// Exists 判断指定字段的记录是否存在。
func (r *gormRepo[T]) Exists(ctx context.Context, tx *gorm.DB, field string, value any) (bool, error) {
	db := r.getDB(ctx, tx)
	var count int64
	if err := db.Model(new(T)).Where(field+" = ?", value).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
