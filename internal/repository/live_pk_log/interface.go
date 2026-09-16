package live_pk_log

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

// sortColumns 允许的排序字段 → SQL 列名，白名单之外的排序请求静默回退默认
var sortColumns = map[string]string{
	"id":            "id",
	"room_id":       "room_id",
	"pk_id":         "pk_id",
	"pk_status":     "pk_status",
	"battle_type":   "battle_type",
	"match_type":    "match_type",
	"rival_uid":     "rival_uid",
	"rival_uname":   "rival_uname",
	"rival_room_id": "rival_room_id",
	"self_votes":    "self_votes",
	"rival_votes":   "rival_votes",
	"self_result":   "self_result",
	"rival_result":  "rival_result",
	"start_at":      "start_at",
	"settle_at":     "settle_at",
	"created_at":    "created_at",
	"updated_at":    "updated_at",
}

// sortOrder 允许的排序方向 → SQL 方向
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

// Repository PK 对战记录数据访问接口
type Repository interface {
	base.Repository[model.LivePkLog]
	// DistinctRoomIDs 获取全表中所有不重复的 RoomID
	DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error)
	// GetByPkID 按 PK ID 查询记录，不存在返回 (nil, nil)
	GetByPkID(ctx context.Context, tx *gorm.DB, pkID int64) (*model.LivePkLog, error)
	// ListPage 分页查询 PK 记录，支持房间号/对方UID/对方名称/我方胜负/开始时间筛选与白名单排序
	ListPage(ctx context.Context, tx *gorm.DB, query model.LivePkLogListPageQuery) ([]model.LivePkLog, int64, error)
	// ListStats 按与 ListPage 相同的筛选条件聚合出场次、我方胜利数与失败数
	ListStats(ctx context.Context, tx *gorm.DB, query model.LivePkLogListPageQuery) (totalNum, winNum, loseNum int64, err error)
}

// DistinctRoomIDs 获取全表中所有不重复的 RoomID
func (r *gormRepo) DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error) {
	db := r.getDB(ctx, tx)
	var roomIDs []int64
	if err := db.Model(&model.LivePkLog{}).Distinct("room_id").Pluck("room_id", &roomIDs).Error; err != nil {
		return nil, err
	}
	return roomIDs, nil
}

// GetByPkID 按 PK ID 查询记录，不存在返回 (nil, nil)
func (r *gormRepo) GetByPkID(ctx context.Context, tx *gorm.DB, pkID int64) (*model.LivePkLog, error) {
	return r.FindOneByField(ctx, tx, "pk_id", pkID)
}

// ListPage 分页查询 PK 记录
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LivePkLogListPageQuery) ([]model.LivePkLog, int64, error) {
	var list []model.LivePkLog
	var total int64
	db := r.getDB(ctx, tx).Model(&model.LivePkLog{})
	db = r.applyLivePkLogListQuery(db, query)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 默认按开始时间倒序；排序参数非法/缺失时静默回退该默认
	orderClause := "start_at desc"
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

// ListStats 与 ListPage 共用同一套筛选条件，保证统计口径与列表完全一致
//
// 场数、胜利数、失败数各自独立统计：self_result = 0 表示没能定位到本直播间，
// 既不算胜也不算负，所以场数不一定等于胜 + 负。
func (r *gormRepo) ListStats(ctx context.Context, tx *gorm.DB, query model.LivePkLogListPageQuery) (totalNum, winNum, loseNum int64, err error) {
	var result struct {
		TotalNum int64
		WinNum   int64
		LoseNum  int64
	}
	db := r.getDB(ctx, tx).Model(&model.LivePkLog{})
	db = r.applyLivePkLogListQuery(db, query)
	// 2 / -1 是 self_result 的取值口径（B站下发值）
	err = db.Select(`
		COUNT(*) AS total_num,
		COALESCE(SUM(CASE WHEN self_result = 2  THEN 1 ELSE 0 END), 0) AS win_num,
		COALESCE(SUM(CASE WHEN self_result = -1 THEN 1 ELSE 0 END), 0) AS lose_num
	`).Scan(&result).Error
	if err != nil {
		return 0, 0, 0, err
	}
	return result.TotalNum, result.WinNum, result.LoseNum, nil
}

// applyLivePkLogListQuery 构建 PK 记录列表筛选条件，ListPage 与 ListStats 共用
func (r *gormRepo) applyLivePkLogListQuery(db *gorm.DB, query model.LivePkLogListPageQuery) *gorm.DB {
	if v := query.RoomID; v != nil {
		db = db.Where("room_id = ?", *v)
	}
	if v := query.RivalUID; v != nil {
		db = db.Where("rival_uid = ?", *v)
	}
	if v := query.RivalUname; v != nil && *v != "" {
		db = db.Where("rival_uname LIKE ?", "%"+escapeLike(*v)+"%")
	}
	if v := query.SelfResult; v != nil {
		db = db.Where("self_result = ?", *v)
	}
	if v := query.StartAtStart; v != nil {
		db = db.Where("start_at >= ?", *v)
	}
	if v := query.StartAtEnd; v != nil {
		db = db.Where("start_at <= ?", *v)
	}
	return db
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
