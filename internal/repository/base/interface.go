package base

import (
	"context"

	"gorm.io/gorm"
)

// Repository 定义通用接口
type Repository[T any] interface {
	// GetByID 根据主键查询记录
	GetByID(ctx context.Context, tx *gorm.DB, id int64) (*T, error)
	// GetByIDs 根据主键批量查询记录
	GetByIDs(ctx context.Context, tx *gorm.DB, ids []int64) ([]T, error)
	// ListAll 查询所有记录（无排序、无分页）
	ListAll(ctx context.Context, tx *gorm.DB) ([]T, error)
	// ListByField 根据指定字段查询记录列表
	ListByField(ctx context.Context, tx *gorm.DB, field string, value any) ([]T, error)
	// GetByField 根据指定字段查询单条记录
	GetByField(ctx context.Context, tx *gorm.DB, field string, value any) (*T, error)
	// Create 创建一条记录，返回创建后的模型
	Create(ctx context.Context, tx *gorm.DB, entity *T) (*T, error)
	// CreateBatch 批量创建记录
	CreateBatch(ctx context.Context, tx *gorm.DB, entities []T) error
	// Save 保存整个实体：主键为零时插入，否则按主键整行覆盖
	Save(ctx context.Context, tx *gorm.DB, entity *T) error
	// UpdateMap 使用map更新指定字段（命中多行则全部更新）
	UpdateMap(ctx context.Context, tx *gorm.DB, field string, value any, updates map[string]any) error
	// UpdateField 根据主键更新单个字段
	UpdateField(ctx context.Context, tx *gorm.DB, id int64, field string, value any) error
	// Delete 根据主键删除记录（软删除）
	Delete(ctx context.Context, tx *gorm.DB, id int64) error
	// DeleteByIDs 根据主键批量删除记录（软删除）
	DeleteByIDs(ctx context.Context, tx *gorm.DB, ids []int64) error
	// Count 统计记录总数
	Count(ctx context.Context, tx *gorm.DB) (int64, error)
	// Exists 判断指定字段的记录是否存在
	Exists(ctx context.Context, tx *gorm.DB, field string, value any) (bool, error)
	// IncrementField 原子调整指定字段的值（字段必须是 int64 类型，delta 可为负）
	IncrementField(ctx context.Context, tx *gorm.DB, id int64, field string, delta int64) error
}

// GetByID 根据主键查询记录
// 如果记录不存在，返回 (nil, nil)
func (r *Repo[T]) GetByID(ctx context.Context, tx *gorm.DB, id int64) (*T, error) {
	db := r.ResolveDB(ctx, tx)
	var entity T
	if err := db.First(&entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// GetByIDs 根据主键批量查询记录
// 返回匹配的记录列表，如果没有匹配记录则返回空切片
func (r *Repo[T]) GetByIDs(ctx context.Context, tx *gorm.DB, ids []int64) ([]T, error) {
	db := r.ResolveDB(ctx, tx)
	var list []T
	if err := db.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ListAll 查询所有记录（无排序、无分页）
func (r *Repo[T]) ListAll(ctx context.Context, tx *gorm.DB) ([]T, error) {
	db := r.ResolveDB(ctx, tx)
	var list []T
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ListByField 根据指定字段查询记录列表
func (r *Repo[T]) ListByField(ctx context.Context, tx *gorm.DB, field string, value any) ([]T, error) {
	db := r.ResolveDB(ctx, tx)
	var list []T
	if err := db.Where(field+" = ?", value).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetByField 根据指定字段查询单条记录
// 如果记录不存在，返回 (nil, nil)
func (r *Repo[T]) GetByField(ctx context.Context, tx *gorm.DB, field string, value any) (*T, error) {
	db := r.ResolveDB(ctx, tx)
	var entity T
	if err := db.Where(field+" = ?", value).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// Create 创建一条记录，返回创建后的模型（包含ID等数据库生成的值）
func (r *Repo[T]) Create(ctx context.Context, tx *gorm.DB, entity *T) (*T, error) {
	db := r.ResolveDB(ctx, tx)
	err := db.Create(entity).Error
	if err != nil {
		return nil, err
	}
	return entity, nil
}

// CreateBatch 批量创建记录
// 如果实体列表为空，则不执行任何操作
func (r *Repo[T]) CreateBatch(ctx context.Context, tx *gorm.DB, entities []T) error {
	db := r.ResolveDB(ctx, tx)
	if len(entities) == 0 {
		return nil
	}
	return db.Create(&entities).Error
}

// Save 保存整个实体：主键为零时插入，否则按主键整行覆盖
func (r *Repo[T]) Save(ctx context.Context, tx *gorm.DB, entity *T) error {
	db := r.ResolveDB(ctx, tx)
	return db.Save(entity).Error
}

// UpdateMap 使用map更新指定字段（命中多行则全部更新）
func (r *Repo[T]) UpdateMap(ctx context.Context, tx *gorm.DB, field string, value any, updates map[string]any) error {
	db := r.ResolveDB(ctx, tx)
	return db.Model(new(T)).Where(field+" = ?", value).Updates(updates).Error
}

// UpdateField 根据主键更新单个字段
func (r *Repo[T]) UpdateField(ctx context.Context, tx *gorm.DB, id int64, field string, value any) error {
	db := r.ResolveDB(ctx, tx)
	return db.Model(new(T)).Where("id = ?", id).Update(field, value).Error
}

// Delete 根据主键删除记录（软删除，模型内嵌 BaseModel 的 DeletedAt）
func (r *Repo[T]) Delete(ctx context.Context, tx *gorm.DB, id int64) error {
	db := r.ResolveDB(ctx, tx)
	return db.Delete(new(T), id).Error
}

// DeleteByIDs 根据主键批量删除记录（软删除）
// 如果主键列表为空，则不执行任何操作
func (r *Repo[T]) DeleteByIDs(ctx context.Context, tx *gorm.DB, ids []int64) error {
	db := r.ResolveDB(ctx, tx)
	if len(ids) == 0 {
		return nil
	}
	return db.Delete(new(T), ids).Error
}

// Count 统计记录总数
func (r *Repo[T]) Count(ctx context.Context, tx *gorm.DB) (int64, error) {
	db := r.ResolveDB(ctx, tx)
	var total int64
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// Exists 判断指定字段的记录是否存在
func (r *Repo[T]) Exists(ctx context.Context, tx *gorm.DB, field string, value any) (bool, error) {
	db := r.ResolveDB(ctx, tx)
	var count int64
	if err := db.Model(new(T)).Where(field+" = ?", value).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// IncrementField 原子调整指定字段的值（字段必须是 int64 类型，delta 可为负）
// 如果没有命中任何记录，返回 gorm.ErrRecordNotFound
func (r *Repo[T]) IncrementField(ctx context.Context, tx *gorm.DB, id int64, field string, delta int64) error {
	db := r.ResolveDB(ctx, tx)
	res := db.Model(new(T)).Where("id = ?", id).Updates(map[string]any{
		field: gorm.Expr(field+" + ?", delta),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
