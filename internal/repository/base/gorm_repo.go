package base

import (
	"context"

	"gorm.io/gorm"
)

// Repo 通用仓储实现。
//
// 提供 base 的全部通用方法，并把 *gorm.DB 的解析规则（事务优先、其次附加 context）
// 以 ResolveDB 暴露给外部——各模块仓库嵌入本类型即可同时获得通用方法与 ResolveDB，
// 不必再各自复制一份解析逻辑。
type Repo[T any] struct {
	db *gorm.DB
}

// NewRepo 构造函数
func NewRepo[T any](db *gorm.DB) *Repo[T] {
	return &Repo[T]{db: db}
}

// ResolveDB 返回本次调用应使用的 *gorm.DB：事务优先，其次附加 context
func (r *Repo[T]) ResolveDB(ctx context.Context, tx *gorm.DB) *gorm.DB {
	db := r.db
	if tx != nil {
		db = tx
	}
	if ctx != nil {
		db = db.WithContext(ctx)
	}
	return db
}
