package live_gift

import (
	"context"
	"strings"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/sqlutil"
	"gorm.io/gorm"
)

// TimeRange 时间范围
type TimeRange struct {
	Start int64 // 起始时间戳（含）
	End   int64 // 结束时间戳（含）
}

// BlindBoxProfit 盲盒盈利统计
type BlindBoxProfit struct {
	Daily   int64 // 本日盈利
	Weekly  int64 // 本周盈利
	Monthly int64 // 本月盈利
	Total   int64 // 总计盈利
}

// sortColumns 允许参与 ListPage 排序的 DB 列
var sortColumns = map[string]string{
	"id":                  "id",
	"room_id":             "room_id",
	"uid":                 "uid",
	"uname":               "uname",
	"gift_type":           "gift_type",
	"gift_name":           "gift_name",
	"price":               "price",
	"num":                 "num",
	"send_at":             "send_at",
	"original":            "original",
	"original_gift_name":  "original_gift_name",
	"original_gift_price": "original_gift_price",
	"created_at":          "created_at",
	"updated_at":          "updated_at",
}

// sortOrder 允许的排序方向 → SQL 方向。
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

type Repository interface {
	base.Repository[model.LiveGift]
	// DistinctRoomIDs 获取全表中所有不重复的 RoomID
	DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error)
	// ListPage 分页查询礼物，Uname/GiftName 模糊匹配，SendAt 范围查询，按 SendAt 倒序
	ListPage(ctx context.Context, tx *gorm.DB, query model.LiveGiftListPageQuery) ([]model.LiveGift, int64, error)
	// SumNumAndAmount 按列表查询条件聚合礼物总数与总金额（price * num）
	SumNumAndAmount(ctx context.Context, tx *gorm.DB, query model.LiveGiftListPageQuery) (totalNum, totalAmount int64, err error)
	// BlindBoxListPage 分页查询盲盒礼物，Uname/GiftName/OriginalGiftName 模糊匹配，SendAt 范围查询，按 SendAt 倒序
	BlindBoxListPage(ctx context.Context, tx *gorm.DB, query model.LiveGiftBlindBoxListPageQuery) ([]model.LiveGift, int64, error)
	// SumOriginalAndCurrentPrice 按盲盒列表查询条件聚合原价总额与现价总额
	SumOriginalAndCurrentPrice(ctx context.Context, tx *gorm.DB, query model.LiveGiftBlindBoxListPageQuery) (originalPrice, currentPrice int64, err error)
	// SumTotalGiftAmountByUID 获取指定uid赠送总金额
	SumTotalGiftAmountByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error)
	// ListByUID 根据 UID 查询礼物，按 SendAt 倒序，limit 控制最大条数
	ListByUID(ctx context.Context, tx *gorm.DB, uid int64, limit int) ([]model.LiveGift, error)
	// ListByLiveID 根据 LiveID 查询礼物，按 SendAt 倒序，limit 控制最大条数
	ListByLiveID(ctx context.Context, tx *gorm.DB, liveID int64, limit int) ([]model.LiveGift, error)
	// UpdateLiveIDByRoomIDAndTimeRange 将指定房间在时间范围内的礼物批量关联到直播记录
	// 判据 room_id + 时间区间，可能命中多行；区间为闭闭 [startTime, endTime]
	UpdateLiveIDByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID, liveID int64) error
	// CountAndRevenueByRoomIDAndTimeRange 统计指定房间在时间范围内的礼物数量与收益（price * num）
	CountAndRevenueByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (count int64, revenue int64, err error)
	// CountGuardByRoomIDAndTimeRange 统计指定房间在时间范围内的大航海数量
	CountGuardByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error)
	// CountSuperChatByRoomIDAndTimeRange 统计指定房间在时间范围内的醒目留言数量
	CountSuperChatByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error)
	// SumBlindBoxProfit 统计盲盒盈利
	SumBlindBoxProfit(ctx context.Context, tx *gorm.DB, uid, roomID int64, day, week, month TimeRange) (*BlindBoxProfit, error)
	// CountDailyByUID 根据uid统计用户在时间范围内的每日消费数量与金额，
	// map key 为「当月第几天」(1-31)
	CountDailyByUID(ctx context.Context, tx *gorm.DB, uid int64, startAt int64, endAt int64) (map[int64]model.LiveGiftDailyGiftStatistics, error)
}

// DistinctRoomIDs 获取全表中所有不重复的 RoomID
func (r *gormRepo) DistinctRoomIDs(ctx context.Context, tx *gorm.DB) ([]int64, error) {
	db := r.ResolveDB(ctx, tx)
	var roomIDs []int64
	if err := db.Model(&model.LiveGift{}).Distinct("room_id").Pluck("room_id", &roomIDs).Error; err != nil {
		return nil, err
	}
	return roomIDs, nil
}

// ListPage 分页查询礼物
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LiveGiftListPageQuery) ([]model.LiveGift, int64, error) {
	var list []model.LiveGift
	var total int64
	db := r.ResolveDB(ctx, tx)
	db = db.Model(&model.LiveGift{})
	db = r.applyLiveGiftListQuery(db, query)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 默认与现状一致；排序参数非法/缺失时静默回退该默认
	orderClause := "send_at desc"
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

// SumNumAndAmount 按列表查询条件聚合礼物总数与总金额
func (r *gormRepo) SumNumAndAmount(ctx context.Context, tx *gorm.DB, query model.LiveGiftListPageQuery) (totalNum, totalAmount int64, err error) {
	var result struct {
		TotalNum    int64
		TotalAmount int64
	}
	db := r.ResolveDB(ctx, tx).Model(&model.LiveGift{})
	db = r.applyLiveGiftListQuery(db, query)
	err = db.Select(`
		COALESCE(SUM(num), 0) AS total_num,
		COALESCE(SUM(num * price), 0) AS total_amount
	`).Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.TotalNum, result.TotalAmount, nil
}

// BlindBoxListPage 分页查询盲盒礼物，Uname/GiftName/OriginalGiftName 模糊匹配，SendAt 范围查询，按 SendAt 倒序
func (r *gormRepo) BlindBoxListPage(ctx context.Context, tx *gorm.DB, query model.LiveGiftBlindBoxListPageQuery) ([]model.LiveGift, int64, error) {
	var list []model.LiveGift
	var total int64
	db := r.ResolveDB(ctx, tx).Model(&model.LiveGift{}).Where("original = ?", enum.No)
	db = r.applyLiveGiftBlindBoxListQuery(db, query)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 默认与现状一致；排序参数非法/缺失时静默回退该默认
	orderClause := "send_at desc"
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

// SumOriginalAndCurrentPrice 按盲盒列表查询条件聚合原价总额与现价总额
func (r *gormRepo) SumOriginalAndCurrentPrice(ctx context.Context, tx *gorm.DB, query model.LiveGiftBlindBoxListPageQuery) (originalPrice, currentPrice int64, err error) {
	var result struct {
		OriginalPrice int64
		CurrentPrice  int64
	}
	db := r.ResolveDB(ctx, tx).Model(&model.LiveGift{}).Where("original = ?", enum.No)
	db = r.applyLiveGiftBlindBoxListQuery(db, query)
	err = db.Select(`
		COALESCE(SUM(num * original_gift_price), 0) AS original_price,
		COALESCE(SUM(num * price), 0) AS current_price
	`).Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.OriginalPrice, result.CurrentPrice, nil
}

// SumTotalGiftAmountByUID 获取指定uid赠送总金额
func (r *gormRepo) SumTotalGiftAmountByUID(ctx context.Context, tx *gorm.DB, uid int64) (int64, error) {
	var amount int64
	err := r.ResolveDB(ctx, tx).
		Model(&model.LiveGift{}).
		Where("uid = ?", uid).
		Select(`COALESCE(SUM(num * price), 0)`).
		Scan(&amount).
		Error
	return amount, err
}

// ListByUID 根据 UID 查询礼物
func (r *gormRepo) ListByUID(ctx context.Context, tx *gorm.DB, uid int64, limit int) ([]model.LiveGift, error) {
	db := r.ResolveDB(ctx, tx)
	var list []model.LiveGift
	err := db.Where("uid = ?", uid).Order("send_at desc").Limit(limit).Find(&list).Error
	return list, err
}

// ListByLiveID 根据 LiveID 查询礼物
func (r *gormRepo) ListByLiveID(ctx context.Context, tx *gorm.DB, liveID int64, limit int) ([]model.LiveGift, error) {
	db := r.ResolveDB(ctx, tx)
	var list []model.LiveGift
	err := db.Where("live_id = ?", liveID).Order("send_at desc").Limit(limit).Find(&list).Error
	return list, err
}

// UpdateLiveIDByRoomIDAndTimeRange 将指定房间在时间范围内的礼物批量关联到直播记录
func (r *gormRepo) UpdateLiveIDByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID, liveID int64) error {
	db := r.ResolveDB(ctx, tx)
	return db.Model(&model.LiveGift{}).
		Where("room_id = ? AND send_at >= ? AND send_at <= ?", roomID, startTime, endTime).
		Update("live_id", liveID).Error
}

// CountAndRevenueByRoomIDAndTimeRange 统计指定房间在时间范围内的礼物数量与收益（price * num）
func (r *gormRepo) CountAndRevenueByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (count int64, revenue int64, err error) {
	db := r.ResolveDB(ctx, tx)
	var result struct {
		Count   int64 `gorm:"column:count"`
		Revenue int64 `gorm:"column:revenue"`
	}
	err = db.Model(&model.LiveGift{}).
		Select("COUNT(*) AS count, COALESCE(SUM(price * num), 0) AS revenue").
		Where("room_id = ? AND send_at >= ? AND send_at <= ?", roomID, startTime, endTime).
		Scan(&result).Error
	return result.Count, result.Revenue, err
}

// CountGuardByRoomIDAndTimeRange 统计指定房间在时间范围内的大航海数量
func (r *gormRepo) CountGuardByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error) {
	db := r.ResolveDB(ctx, tx)
	var count int64
	err := db.Model(&model.LiveGift{}).
		Where("room_id = ? AND send_at >= ? AND send_at <= ? AND gift_type = ?", roomID, startTime, endTime, enum.GiftTypeGuard).
		Count(&count).Error
	return count, err
}

// CountSuperChatByRoomIDAndTimeRange 统计指定房间在时间范围内的醒目留言数量
func (r *gormRepo) CountSuperChatByRoomIDAndTimeRange(ctx context.Context, tx *gorm.DB, startTime, endTime, roomID int64) (int64, error) {
	db := r.ResolveDB(ctx, tx)
	var count int64
	err := db.Model(&model.LiveGift{}).
		Where("room_id = ? AND send_at >= ? AND send_at <= ? AND gift_type = ?", roomID, startTime, endTime, enum.GiftTypeSuperChat).
		Count(&count).Error
	return count, err
}

// SumBlindBoxProfit 统计盲盒盈利
func (r *gormRepo) SumBlindBoxProfit(ctx context.Context, tx *gorm.DB, uid, roomID int64, day, week, month TimeRange) (*BlindBoxProfit, error) {
	db := r.ResolveDB(ctx, tx)
	db = db.Model(&model.LiveGift{}).Where("original = ?", enum.No)
	// 筛选用户或房间
	if uid > 0 {
		db = db.Where("uid = ?", uid)
	}
	if roomID > 0 {
		db = db.Where("room_id = ?", roomID)
	}
	var result BlindBoxProfit
	err := db.Select(`
		COALESCE(SUM(CASE WHEN send_at >= ? AND send_at <= ? THEN (price - original_gift_price) * num ELSE 0 END), 0) AS daily,
		COALESCE(SUM(CASE WHEN send_at >= ? AND send_at <= ? THEN (price - original_gift_price) * num ELSE 0 END), 0) AS weekly,
		COALESCE(SUM(CASE WHEN send_at >= ? AND send_at <= ? THEN (price - original_gift_price) * num ELSE 0 END), 0) AS monthly,
		COALESCE(SUM((price - original_gift_price) * num), 0) AS total`,
		day.Start, day.End,
		week.Start, week.End,
		month.Start, month.End,
	).Scan(&result).Error
	return &result, err
}

// CountDailyByUID 根据uid统计用户在时间范围内的每日消费数量与金额，
// map key 为「当月第几天」(1-31)，value 为该日的礼物数量与金额(分)
func (r *gormRepo) CountDailyByUID(ctx context.Context, tx *gorm.DB, uid int64, startAt int64, endAt int64) (map[int64]model.LiveGiftDailyGiftStatistics, error) {
	db := r.ResolveDB(ctx, tx)
	type row struct {
		SendAt int64
		Num    int64
		Price  int64
	}
	var rows []row
	err := db.Model(&model.LiveGift{}).Select("send_at, num, price").Where("uid = ?", uid).Where("send_at >= ?", startAt).Where("send_at < ?", endAt).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]model.LiveGiftDailyGiftStatistics)
	for _, row := range rows {
		day := int64(time.Unix(row.SendAt, 0).In(time.Local).Day())
		result[day] = model.LiveGiftDailyGiftStatistics{
			Num:    result[day].Num + row.Num,
			Amount: result[day].Amount + row.Num*row.Price,
		}
	}
	return result, nil
}

// applyLiveGiftListQuery 构建礼物列表筛选条件
func (r *gormRepo) applyLiveGiftListQuery(db *gorm.DB, query model.LiveGiftListPageQuery) *gorm.DB {
	if v := query.RoomID; v != nil {
		db = db.Where("room_id = ?", *v)
	}
	if v := query.UID; v != nil {
		db = db.Where("uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("uname LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if v := query.GiftName; v != nil && *v != "" {
		db = db.Where("gift_name LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if v := query.GiftType; v != nil {
		g := enum.GiftType(*v)
		if g.IsValid() {
			db = db.Where("gift_type = ?", g)
		}
	}
	if v := query.Original; v != nil {
		o := enum.YesNo(*v)
		if o.IsValid() {
			db = db.Where("original = ?", o)
		}
	}
	if v := query.SendAtStart; v != nil {
		db = db.Where("send_at >= ?", *v)
	}
	if v := query.SendAtEnd; v != nil {
		db = db.Where("send_at <= ?", *v)
	}
	return db
}

// applyLiveGiftBlindBoxListQuery 构建盲盒礼物列表筛选条件
func (r *gormRepo) applyLiveGiftBlindBoxListQuery(db *gorm.DB, query model.LiveGiftBlindBoxListPageQuery) *gorm.DB {
	if v := query.RoomID; v != nil {
		db = db.Where("room_id = ?", *v)
	}
	if v := query.UID; v != nil {
		db = db.Where("uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("uname LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if v := query.GiftName; v != nil && *v != "" {
		db = db.Where("gift_name LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if v := query.OriginalGiftName; v != nil && *v != "" {
		db = db.Where("original_gift_name LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if v := query.SendAtStart; v != nil {
		db = db.Where("send_at >= ?", *v)
	}
	if v := query.SendAtEnd; v != nil {
		db = db.Where("send_at <= ?", *v)
	}
	return db
}
