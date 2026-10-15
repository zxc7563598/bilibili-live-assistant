package live_guard_record

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/sqlutil"
	"gorm.io/gorm"
)

// RosterItem 名单一行，粒度为「用户 × 月份」
type RosterItem struct {
	UID        int64
	Uname      string
	GuardLevel enum.BadgeType
	CoverFrom  int64 // 连续覆盖段的起点月首
	CoverTo    int64 // 连续覆盖段的终点月首, 不含
	OpenNum    int64 // 当月开通份数
	OpenAt     int64 // 当月最早一次开通时间, 0 表示当月没有开通
}

// OpenStat 用户在指定区间内的开通统计
type OpenStat struct {
	UID     int64
	OpenNum int64
	OpenAt  int64
}

// rosterSortColumns 允许参与名单排序的表达式。名单按 uid 聚合，
// 排序只能落在聚合结果上，所以这里存的是表达式而不是列名。
//
// 没有 cover_from / cover_to：这两列的展示值是连续覆盖段的端点，SQL 里只有
// 「命中查询月的那一段」，按它排序会和页面上显示的值对不上。
var rosterSortColumns = map[string]string{
	"uid":   "uid",
	"uname": "MAX(uname)",
}

// sortOrder 允许的排序方向 → SQL 方向
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

type Repository interface {
	base.Repository[model.LiveGuardRecord]

	// ListRosterPage 分页查询某月某档位的名单，按 uid 聚合，返回名单与去重人数。
	// 名单月份区间 [cover_from, cover_to) 与查询月份区间求交，两端都是月首，所以结果就是月份归属。
	// 返回行的区间是覆盖查询月的连续覆盖段，见 fillCoverSpan。
	ListRosterPage(ctx context.Context, tx *gorm.DB, query model.LiveGuardRecordListPageQuery) ([]RosterItem, int64, error)
	// ListOpenStats 批量统计给定用户在指定区间内的开通份数与最早开通时间。
	// UIDs 为空时直接返回 nil。
	ListOpenStats(ctx context.Context, tx *gorm.DB, query model.LiveGuardRecordOpenStatsQuery) ([]OpenStat, error)
	// CountRoster 统计名单去重人数。limit > 0 时只保证「不超过 limit」的语义
	CountRoster(ctx context.Context, tx *gorm.DB, query model.LiveGuardRecordListPageQuery, limit int) (int64, error)
	// RosterChunk 导出分块读取：uid < afterUID（afterUID 为 0 表示不限），按 uid 倒序取至多 limit 行。
	// 名单里 uid 唯一（调用方必须带 guard_level），所以 uid 就是分块游标。
	RosterChunk(ctx context.Context, tx *gorm.DB, query model.LiveGuardRecordListPageQuery, afterUID int64, limit int) ([]RosterItem, error)
	// ListByUIDAndLevel 按开通时间升序列出某用户某档位的全部台账。
	// 补录往历史里插数据后要重排名单月份区间，靠它拿准「谁在谁后面」。
	ListByUIDAndLevel(ctx context.Context, tx *gorm.DB, uid int64, level enum.BadgeType) ([]model.LiveGuardRecord, error)
	// ListByUIDs 批量取回给定用户的全部台账，按 uid、档位、开通时间升序。
	// 名单区间要落在连续覆盖段上，命中查询月的那一条只是段里的一行，所以按页整段取回再合并。
	ListByUIDs(ctx context.Context, tx *gorm.DB, uids []int64) ([]model.LiveGuardRecord, error)
	// LatestByUIDAndLevel 取某用户某档位开通时间最晚的一条台账，没有返回 nil。
	// 直播链路写台账时靠它拿顺延基准，那一行是低频写入，多一次索引查询可以接受。
	LatestByUIDAndLevel(ctx context.Context, tx *gorm.DB, uid int64, level enum.BadgeType) (*model.LiveGuardRecord, error)
	// UpdateCoverByID 改写一条台账的名单月份区间。
	// 只给补录后移后续区间用：台账里只有区间是可推导的，开通时间/价格/数量都是既成事实。
	UpdateCoverByID(ctx context.Context, tx *gorm.DB, id int64, coverFrom, coverTo int64) error
}

// applyRosterFilter 构建名单筛选条件，列表/计数/导出共用
//
// 区间两端都是月首、查询区间也是月首到次月首，所以半开区间求交等价于「这个月归不归他」。
func (r *gormRepo) applyRosterFilter(db *gorm.DB, query model.LiveGuardRecordListPageQuery) *gorm.DB {
	if v := query.GuardLevel; v != nil {
		if g := enum.BadgeType(*v); g != enum.BadgeTypeL0 && g.IsValid() {
			db = db.Where("guard_level = ?", g)
		}
	}
	if query.MonthTo > 0 {
		db = db.Where("cover_from < ?", query.MonthTo)
	}
	if query.MonthFrom > 0 {
		db = db.Where("cover_to > ?", query.MonthFrom)
	}
	if v := query.UID; v != nil {
		db = db.Where("uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("uname LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	return db
}

// rosterSelect 名单聚合的查询列。guard_level 取 MIN 是为了让它成为合法聚合列，
// 同一档位下取值恒定；直接选裸列在 PostgreSQL 上会因 GROUP BY 报错。
//
// 不含区间列：区间由 fillCoverSpan 在连续覆盖段上补齐。
const rosterSelect = "uid, MAX(uname) AS uname, MIN(guard_level) AS guard_level"

// ListRosterPage 分页查询某月某档位的名单
func (r *gormRepo) ListRosterPage(ctx context.Context, tx *gorm.DB, query model.LiveGuardRecordListPageQuery) ([]RosterItem, int64, error) {
	total, err := r.CountRoster(ctx, tx, query, 0)
	if err != nil {
		return nil, 0, err
	}
	db := r.applyRosterFilter(r.ResolveDB(ctx, tx).Model(&model.LiveGuardRecord{}), query)
	// 默认与现状一致；排序参数非法/缺失时静默回退该默认
	orderClause := "uid asc"
	if query.SortField != nil && query.SortOrder != nil {
		if expr, ok := rosterSortColumns[*query.SortField]; ok {
			if dir, ok := sortOrder[strings.ToLower(*query.SortOrder)]; ok {
				// expr/dir 均来自字面量白名单，杜绝注入；uid 兜底保证同键值时翻页稳定
				orderClause = expr + " " + dir + ", uid asc"
			}
		}
	}
	var list []RosterItem
	err = db.Select(rosterSelect).Group("uid").
		Order(orderClause).Offset(query.Offset).Limit(query.Limit).Scan(&list).Error
	if err != nil {
		return nil, 0, err
	}
	if err := r.fillCoverSpan(ctx, tx, list, query.MonthFrom, query.MonthTo); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// CountRoster 统计名单去重人数
func (r *gormRepo) CountRoster(ctx context.Context, tx *gorm.DB, query model.LiveGuardRecordListPageQuery, limit int) (int64, error) {
	db := r.ResolveDB(ctx, tx)
	sub := r.applyRosterFilter(db.Model(&model.LiveGuardRecord{}).Select("1"), query).Group("uid")
	if limit > 0 {
		sub = sub.Limit(limit)
	}
	var total int64
	// 子查询包一层再 count：GORM 的 Count 会剥掉外层 Limit，必须用派生表把早停固定下来
	err := db.Table("(?) AS t", sub).Count(&total).Error
	return total, err
}

// ListOpenStats 批量统计给定用户在指定区间内的开通份数与最早开通时间
func (r *gormRepo) ListOpenStats(ctx context.Context, tx *gorm.DB, query model.LiveGuardRecordOpenStatsQuery) ([]OpenStat, error) {
	if len(query.UIDs) == 0 {
		return nil, nil
	}
	db := r.ResolveDB(ctx, tx).Model(&model.LiveGuardRecord{}).
		Where("uid IN ?", query.UIDs).
		Where("open_at >= ? AND open_at < ?", query.OpenFrom, query.OpenTo)
	if v := query.GuardLevel; v != nil {
		if g := enum.BadgeType(*v); g != enum.BadgeTypeL0 && g.IsValid() {
			db = db.Where("guard_level = ?", g)
		}
	}
	var stats []OpenStat
	err := db.Select("uid, COALESCE(SUM(num), 0) AS open_num, MIN(open_at) AS open_at").
		Group("uid").Scan(&stats).Error
	return stats, err
}

// RosterChunk 导出分块读取
func (r *gormRepo) RosterChunk(ctx context.Context, tx *gorm.DB, query model.LiveGuardRecordListPageQuery, afterUID int64, limit int) ([]RosterItem, error) {
	db := r.applyRosterFilter(r.ResolveDB(ctx, tx).Model(&model.LiveGuardRecord{}), query)
	if afterUID > 0 {
		db = db.Where("uid < ?", afterUID)
	}
	var list []RosterItem
	if err := db.Select(rosterSelect).Group("uid").
		Order("uid desc").Limit(limit).Scan(&list).Error; err != nil {
		return nil, err
	}
	if err := r.fillCoverSpan(ctx, tx, list, query.MonthFrom, query.MonthTo); err != nil {
		return nil, err
	}
	uids := make([]int64, 0, len(list))
	for _, item := range list {
		uids = append(uids, item.UID)
	}
	stats, err := r.ListOpenStats(ctx, tx, model.LiveGuardRecordOpenStatsQuery{
		GuardLevel: query.GuardLevel,
		OpenFrom:   query.MonthFrom,
		OpenTo:     query.MonthTo,
		UIDs:       uids,
	})
	if err != nil {
		return nil, err
	}
	statByUID := make(map[int64]OpenStat, len(stats))
	for _, stat := range stats {
		statByUID[stat.UID] = stat
	}
	for i := range list {
		if stat, ok := statByUID[list[i].UID]; ok {
			list[i].OpenNum = stat.OpenNum
			list[i].OpenAt = stat.OpenAt
		}
	}
	return list, nil
}

// ListByUIDAndLevel 按开通时间升序列出某用户某档位的全部台账
//
// 顺序（含同一时刻的 id 序）要与写入时一致，重排出来的区间才接得上。
func (r *gormRepo) ListByUIDAndLevel(ctx context.Context, tx *gorm.DB, uid int64, level enum.BadgeType) ([]model.LiveGuardRecord, error) {
	var list []model.LiveGuardRecord
	err := r.ResolveDB(ctx, tx).
		Where("uid = ? AND guard_level = ?", uid, level).
		Order("open_at asc, id asc").Find(&list).Error
	return list, err
}

// ListByUIDs 批量取回给定用户的全部台账
func (r *gormRepo) ListByUIDs(ctx context.Context, tx *gorm.DB, uids []int64) ([]model.LiveGuardRecord, error) {
	var list []model.LiveGuardRecord
	if len(uids) == 0 {
		return list, nil
	}
	err := r.ResolveDB(ctx, tx).
		Where("uid IN ?", uids).
		Order("uid asc, guard_level asc, open_at asc, id asc").Find(&list).Error
	return list, err
}

// fillCoverSpan 把名单行的区间改写成「覆盖查询月的连续覆盖段」
//
// 名单行只回答「这个月归不归他」，区间要回答「他这一路连续覆盖到哪个月」。命中查询月的
// 台账只有一条，单取它就把分次续费连排出来的月份压成一个月（只有一次买多份才看得见跨度）。
// 所以按 uid 整段取回台账再合并，档位从名单行自身取——一行名单只代表一个档位。
func (r *gormRepo) fillCoverSpan(ctx context.Context, tx *gorm.DB, list []RosterItem, monthFrom, monthTo int64) error {
	if len(list) == 0 {
		return nil
	}
	uids := make([]int64, 0, len(list))
	for _, item := range list {
		uids = append(uids, item.UID)
	}
	rows, err := r.ListByUIDs(ctx, tx, uids)
	if err != nil {
		return err
	}
	// 档位不同就不是同一段，按 (uid, 档位) 分组；组内按开通时间升序，合并时靠的就是这个顺序
	type uidLevel struct {
		uid   int64
		level enum.BadgeType
	}
	groups := make(map[uidLevel][]model.LiveGuardRecord, len(list))
	for _, row := range rows {
		key := uidLevel{uid: row.UID, level: row.GuardLevel}
		groups[key] = append(groups[key], row)
	}
	for i := range list {
		key := uidLevel{uid: list[i].UID, level: list[i].GuardLevel}
		if from, to, ok := coverSpan(groups[key], monthFrom, monthTo); ok {
			list[i].CoverFrom, list[i].CoverTo = from, to
		}
	}
	return nil
}

// coverSpan 取覆盖查询月的连续覆盖段 [起点, 终点)
//
// 台账按开通时间升序时区间起点也升序，相邻两条首尾相接就是同一段，中间断了另起一段，
// 跨度不跨断档。返回 false 表示没有任何一段覆盖查询月（查询条件与台账对不上）。
func coverSpan(rows []model.LiveGuardRecord, monthFrom, monthTo int64) (int64, int64, bool) {
	for i := 0; i < len(rows); {
		from, to := rows[i].CoverFrom, rows[i].CoverTo
		i++
		for i < len(rows) && rows[i].CoverFrom <= to {
			if rows[i].CoverTo > to {
				to = rows[i].CoverTo
			}
			i++
		}
		if from < monthTo && to > monthFrom {
			return from, to, true
		}
	}
	return 0, 0, false
}

// LatestByUIDAndLevel 取某用户某档位开通时间最晚的一条台账，没有返回 nil
//
// 排序键与 ListByUIDAndLevel 对齐（同一时刻按 id 取后者），直播链路取到的
// 顺延基准才是「当前排到哪个月」。
func (r *gormRepo) LatestByUIDAndLevel(ctx context.Context, tx *gorm.DB, uid int64, level enum.BadgeType) (*model.LiveGuardRecord, error) {
	var record model.LiveGuardRecord
	err := r.ResolveDB(ctx, tx).
		Where("uid = ? AND guard_level = ?", uid, level).
		Order("open_at desc, id desc").First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateCoverByID 改写一条台账的名单月份区间
func (r *gormRepo) UpdateCoverByID(ctx context.Context, tx *gorm.DB, id int64, coverFrom, coverTo int64) error {
	return r.UpdateMap(ctx, tx, "id", id, map[string]any{
		"cover_from": coverFrom,
		"cover_to":   coverTo,
	})
}
