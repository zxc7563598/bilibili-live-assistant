package base

import (
	"context"

	"gorm.io/gorm"
)

// gormRepo 实现
type gormRepo[T any] struct {
	db *gorm.DB
}

// New 构造函数
func New[T any](db *gorm.DB) Repository[T] {
	return &gormRepo[T]{db: db}
}

// getDB 封装 ctx + tx
func (r *gormRepo[T]) getDB(ctx context.Context, tx *gorm.DB) *gorm.DB {
	db := r.db
	if tx != nil {
		db = tx
	}
	if ctx != nil {
		db = db.WithContext(ctx)
	}
	return db
}
