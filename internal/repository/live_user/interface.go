package live_user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/sqlutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 资产字段名，AdjustCredit 只接受这两个值
const (
	CreditFieldPoints = "points"
	CreditFieldStars  = "stars"
)

// sortColumns 允许参与 ListPage 排序的 DB 列
var sortColumns = map[string]string{
	"id":                "id",
	"uid":               "uid",
	"uname":             "uname",
	"points":            "points",
	"stars":             "stars",
	"total_danmu_count": "total_danmu_count",
	"total_gift_amount": "total_gift_amount",
	"created_at":        "created_at",
	"updated_at":        "updated_at",
}

// sortOrder 允许的排序方向 → SQL 方向。
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

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
	// ExistsByUID 根据 B站 UID 获取用户是否存在
	ExistsByUID(ctx context.Context, tx *gorm.DB, uid int64) (bool, error)
	// UpdateNameByID 根据 ID 变更用户昵称
	UpdateNameByID(ctx context.Context, tx *gorm.DB, id int64, uname string) error
	// UpdateFaceByID 根据 ID 变更用户头像 URL
	UpdateFaceByID(ctx context.Context, tx *gorm.DB, id int64, face string) error
	// UpdatePasswordByID 根据 ID 变更用户密码
	UpdatePasswordByID(ctx context.Context, tx *gorm.DB, id int64, password string) error
	// CreateIfNotExist 若 uid 已存在则忽略创建并返回已有记录，否则创建新记录
	CreateIfNotExist(ctx context.Context, tx *gorm.DB, entity *model.LiveUser) (*model.LiveUser, error)
	// ListPage 分页查询用户，UID 精确匹配，Uname 模糊匹配；
	// 支持按白名单字段排序，非法/空排序参数回退按 created_at desc
	ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserListPageQuery) ([]model.LiveUser, int64, error)
	// CountFiltered 统计命中行数。limit > 0 时只保证「不超过 limit」的语义，
	// 实现上用「子查询 + LIMIT limit」早停，避免在大表上做全量 COUNT
	CountFiltered(ctx context.Context, tx *gorm.DB, query model.LiveUserListPageQuery, limit int) (int64, error)
	// ExportChunk 导出用的分块读取：主键 < afterID（afterID 为 0 表示不限）且满足筛选条件，
	// 按主键倒序取至多 limit 行。
	//
	// 按主键而非 created_at 分块：主键有索引且唯一，天然满足「已取过的行不会再出现、不会漏」。
	ExportChunk(ctx context.Context, tx *gorm.DB, query model.LiveUserListPageQuery, afterID int64, limit int) ([]model.LiveUser, error)
	// AdjustCredit 原子增减用户资产（积分/星光），返回变更前、变更后的数值
	AdjustCredit(ctx context.Context, tx *gorm.DB, id int64, field string, delta int64) (before, after int64, err error)
	// UpdateTokenByID 根据 id 更换用户 refreshToken
	UpdateTokenByID(ctx context.Context, tx *gorm.DB, id int64, token *string) error
}

// GetByUID 根据 B站 UID 查询单条用户记录
func (r *gormRepo) GetByUID(ctx context.Context, tx *gorm.DB, uid int64) (*model.LiveUser, error) {
	return r.GetByField(ctx, tx, "uid", uid)
}

// ExistsByUID 根据 B站 UID 获取用户是否存在
func (r *gormRepo) ExistsByUID(ctx context.Context, tx *gorm.DB, uid int64) (bool, error) {
	return r.Exists(ctx, tx, "uid", uid)
}

// applyFilters 应用筛选条件。
//
// ListPage 与导出共用同一份条件拼装 —— 两边口径必须一致，否则导出结果与页面上看到的对不上。
func applyFilters(db *gorm.DB, query model.LiveUserListPageQuery) *gorm.DB {
	if v := query.UID; v != nil {
		db = db.Where("uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("uname LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	return db
}

// ListPage 分页查询用户
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserListPageQuery) ([]model.LiveUser, int64, error) {
	var list []model.LiveUser
	var total int64
	db := applyFilters(r.ResolveDB(ctx, tx).Model(&model.LiveUser{}), query)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 默认与现状一致；排序参数非法/缺失时静默回退该默认
	orderClause := "created_at desc"
	if query.SortField != nil && query.SortOrder != nil {
		// 方向先小写归一化再查白名单，非法值直接回退默认
		if field, ok := sortColumns[*query.SortField]; ok {
			if dir, ok := sortOrder[strings.ToLower(*query.SortOrder)]; ok {
				// field/dir 均来自字面量白名单，杜绝注入；id asc 保证同键值时翻页稳定
				orderClause = field + " " + dir + ", id asc"
			}
		}
	}
	err := db.Order(orderClause).Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	return list, total, err
}

// CountFiltered 统计命中行数，limit > 0 时提前停止
func (r *gormRepo) CountFiltered(ctx context.Context, tx *gorm.DB, query model.LiveUserListPageQuery, limit int) (int64, error) {
	db := r.ResolveDB(ctx, tx)
	sub := applyFilters(db.Model(&model.LiveUser{}).Select("1"), query)
	if limit > 0 {
		sub = sub.Limit(limit)
	}
	var total int64
	// 子查询包一层再 count：GORM 的 Count 会剥掉外层 Limit，必须用派生表把早停固定下来
	err := db.Table("(?) AS t", sub).Count(&total).Error
	return total, err
}

// ExportChunk 导出用的分块读取
func (r *gormRepo) ExportChunk(ctx context.Context, tx *gorm.DB, query model.LiveUserListPageQuery, afterID int64, limit int) ([]model.LiveUser, error) {
	var list []model.LiveUser
	db := applyFilters(r.ResolveDB(ctx, tx).Model(&model.LiveUser{}), query)
	if afterID > 0 {
		db = db.Where("id < ?", afterID)
	}
	err := db.Order("id desc").Limit(limit).Find(&list).Error
	return list, err
}

// UpdateNameByID 根据 ID 变更用户昵称
func (r *gormRepo) UpdateNameByID(ctx context.Context, tx *gorm.DB, id int64, uname string) error {
	return r.UpdateField(ctx, tx, id, "uname", uname)
}

// UpdateFaceByID 根据 ID 变更用户头像 URL
func (r *gormRepo) UpdateFaceByID(ctx context.Context, tx *gorm.DB, id int64, face string) error {
	return r.UpdateField(ctx, tx, id, "face", face)
}

// UpdatePasswordByID 根据 ID 变更用户密码
func (r *gormRepo) UpdatePasswordByID(ctx context.Context, tx *gorm.DB, id int64, password string) error {
	return r.UpdateField(ctx, tx, id, "password", password)
}

// CreateIfNotExist 若 uid 已存在则忽略创建并返回已有记录，否则创建新记录
func (r *gormRepo) CreateIfNotExist(ctx context.Context, tx *gorm.DB, entity *model.LiveUser) (*model.LiveUser, error) {
	db := r.ResolveDB(ctx, tx)
	res := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "uid"}},
		DoNothing: true,
	}).Create(entity)
	if res.Error != nil {
		return nil, res.Error
	}
	// 冲突未插入（RowsAffected == 0），返回已存在的记录
	if res.RowsAffected == 0 {
		return r.GetByUID(ctx, tx, entity.UID)
	}
	return entity, nil
}

// AdjustCredit 原子增减用户资产（积分/星光），delta 为负数表示扣减，保证结果不会小于 0
//
// 直播消息每条都在独立 goroutine 中处理，同一用户的资产变更天然并发。
// 这里把读改写整体交给数据库（UPDATE ... SET field = field + ?），
// 避免"先查询再写绝对值"被并发覆盖导致积分丢失。
// 返回变更前、变更后的数值，供调用方写入变动流水。
func (r *gormRepo) AdjustCredit(ctx context.Context, tx *gorm.DB, id int64, field string, delta int64) (before, after int64, err error) {
	if field != CreditFieldPoints && field != CreditFieldStars {
		return 0, 0, fmt.Errorf("%w: %s", ErrInvalidCreditField, field)
	}
	db := r.ResolveDB(ctx, tx).Model(&model.LiveUser{}).Where("id = ?", id)
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
		var exists bool
		if exists, err = r.Exists(ctx, tx, "id", id); err != nil {
			return 0, 0, err
		}
		if !exists {
			return 0, 0, ErrUserNotFound
		}
		return 0, 0, ErrInsufficientBalance
	}
	// 同事务内回读变更后的数值，变更前的值由 after - delta 反推
	if err = r.ResolveDB(ctx, tx).Model(&model.LiveUser{}).Where("id = ?", id).Pluck(field, &after).Error; err != nil {
		return 0, 0, err
	}
	return after - delta, after, nil
}

// UpdateTokenByID 根据 id 更换用户 refreshToken
func (r *gormRepo) UpdateTokenByID(ctx context.Context, tx *gorm.DB, id int64, token *string) error {
	return r.UpdateField(ctx, tx, id, "token", token)
}
