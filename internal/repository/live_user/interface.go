package live_user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 资产字段名，AddCredit 只接受这两个值
const (
	CreditFieldPoints = "points"
	CreditFieldStars  = "stars"
)

var (
	// ErrInsufficientBalance 扣减时余额不足
	ErrInsufficientBalance = errors.New("用户资产余额不足")
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("用户不存在")
	// ErrInvalidCreditField 资产字段名非法
	ErrInvalidCreditField = errors.New("非法的用户资产字段")
)

type Repository interface {
	base.Repository[model.LiveUser]
	// GetByUID 根据 B站 UID 查询单条用户记录
	GetByUID(ctx context.Context, tx *gorm.DB, uid int64) (*model.LiveUser, error)
	// UpdateName 根据 ID 变更用户昵称
	UpdateName(ctx context.Context, tx *gorm.DB, id int64, uname string) error
	// CreateIfNotExist 若 uid 已存在则忽略创建并返回已有记录，否则创建新记录
	CreateIfNotExist(ctx context.Context, tx *gorm.DB, entity *model.LiveUser) (*model.LiveUser, error)
	// ListPage 分页查询用户，UID 精确匹配，Uname 模糊匹配，按 CreatedAt 倒序
	ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserListPageQuery) ([]model.LiveUser, int64, error)
	// AddCredit 原子增减用户资产（积分/星光），返回变更前、变更后的数值
	AddCredit(ctx context.Context, tx *gorm.DB, id int64, field string, delta int64) (int64, int64, error)
}

// GetByUID 根据 B站 UID 查询单条用户记录
func (r *gormRepo) GetByUID(ctx context.Context, tx *gorm.DB, uid int64) (*model.LiveUser, error) {
	return r.FindOneByField(ctx, tx, "uid", uid)
}

// ListPage 分页查询用户
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserListPageQuery) ([]model.LiveUser, int64, error) {
	var list []model.LiveUser
	var total int64
	db := r.getDB(ctx, tx)
	db = db.Model(&model.LiveUser{})
	if v := query.UID; v != nil {
		db = db.Where("uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("uname LIKE ?", "%"+escapeLike(*v)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("created_at desc").Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	return list, total, err
}

// UpdateName 根据 ID 变更用户昵称
func (r *gormRepo) UpdateName(ctx context.Context, tx *gorm.DB, id int64, uname string) error {
	return r.UpdateField(ctx, tx, id, "uname", uname)
}

// CreateIfNotExist 若 uid 已存在则忽略创建并返回已有记录，否则创建新记录
func (r *gormRepo) CreateIfNotExist(ctx context.Context, tx *gorm.DB, entity *model.LiveUser) (*model.LiveUser, error) {
	db := r.getDB(ctx, tx)
	res := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "uid"}},
		DoNothing: true,
	}).Create(entity)
	if res.Error != nil {
		return nil, res.Error
	}
	// 冲突未插入（RowsAffected == 0），返回已存在的记录
	if res.RowsAffected == 0 {
		return r.GetByUID(ctx, tx, entity.Uid)
	}
	return entity, nil
}

// AddCredit 原子增减用户资产（积分/星光），delta 为负数表示扣减，保证结果不会小于 0
//
// 直播消息每条都在独立 goroutine 中处理，同一用户的资产变更天然并发。
// 这里把读改写整体交给数据库（UPDATE ... SET field = field + ?），
// 避免"先查询再写绝对值"被并发覆盖导致积分丢失。
// 返回变更前、变更后的数值，供调用方写入变动流水。
func (r *gormRepo) AddCredit(ctx context.Context, tx *gorm.DB, id int64, field string, delta int64) (int64, int64, error) {
	if field != CreditFieldPoints && field != CreditFieldStars {
		return 0, 0, fmt.Errorf("%w: %s", ErrInvalidCreditField, field)
	}
	db := r.getDB(ctx, tx).Model(&model.LiveUser{}).Where("id = ?", id)
	if delta < 0 {
		// 余额不足时条件不成立，RowsAffected 为 0，从而拒绝本次扣减
		db = db.Where(field+" >= ?", -delta)
	}
	res := db.Updates(map[string]any{field: gorm.Expr(field+" + ?", delta)})
	if res.Error != nil {
		return 0, 0, res.Error
	}
	if res.RowsAffected == 0 {
		// 没更新到行，区分"用户不存在"和"余额不足"
		exists, err := r.Exists(ctx, tx, "id", id)
		if err != nil {
			return 0, 0, err
		}
		if !exists {
			return 0, 0, ErrUserNotFound
		}
		return 0, 0, ErrInsufficientBalance
	}
	// 同事务内回读变更后的数值，变更前的值由 after - delta 反推
	var after int64
	if err := r.getDB(ctx, tx).Model(&model.LiveUser{}).Where("id = ?", id).Pluck(field, &after).Error; err != nil {
		return 0, 0, err
	}
	return after - delta, after, nil
}

// escapeLike 转义 LIKE 查询中的特殊字符 _ %
func escapeLike(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '%' || r == '_' || r == '\\' {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
