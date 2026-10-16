package liveguard

import (
	"context"
	"fmt"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_guard_record"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/order"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// guardRestack 一次补录并入台账序列后重排出来的结果
type guardRestack struct {
	// CoverFrom/CoverTo 本次补录覆盖的名单月份区间
	CoverFrom int64
	CoverTo   int64
	// Shifts 被本次补录顶到后面、需要回写的记录
	Shifts []model.LiveGuardRecord
	// ExpireAt 重排后该用户该档位的到期时间，天数口径，给回写 live_users 用
	ExpireAt int64
}

// planGuardRestack 把一次补录并入该用户该档位的台账序列，重排名单月份区间
//
// 名单月份的唯一依据是台账：补录可能插在已有记录中间，插进去之后后面那些区间都要
// 整体后移（一次开通买多份就是连着排几个月），所以这里从最早一条开始整段重算。
//
// 不能拿 live_users 的到期时间当顺延基准：那是「身份有效到什么时候」的天数口径，
// 往历史里插数据时它会盖上后来的记录，月份区间直接串位。同一个循环里顺带按天数口径
// 走一遍，是为了补录回写 live_users 时有值可依，两份结果各写各的表。
//
// 区间两端都是推导值，重排不会碰开通时间、价格、数量这些既成事实。
func (s *Service) planGuardRestack(ctx context.Context, uid int64, level enum.BadgeType, openAt, num int64) (guardRestack, int, error) {
	rows, err := s.liveGuardRepo.ListByUIDAndLevel(ctx, nil, uid, level)
	if err != nil {
		return guardRestack{}, CodeQueryFailed, err
	}
	// 本次补录并入序列：同一时刻的开通，已有记录在前
	newRow := model.LiveGuardRecord{UID: uid, GuardLevel: level, OpenAt: openAt, Num: num}
	merged := make([]model.LiveGuardRecord, 0, len(rows)+1)
	placed := false
	for _, row := range rows {
		if !placed && row.OpenAt > openAt {
			merged = append(merged, newRow)
			placed = true
		}
		merged = append(merged, row)
	}
	if !placed {
		merged = append(merged, newRow)
	}

	result := guardRestack{}
	var prevCoverTo, prevExpireAt int64
	for _, row := range merged {
		openTime := time.Unix(row.OpenAt, 0)
		coverFrom, coverTo := liveuser.GuardCoverWindow(prevCoverTo, openTime, row.Num)
		prevCoverTo = coverTo
		_, expireEnd := liveuser.GuardExpireWindow(prevExpireAt, openTime, row.Num)
		prevExpireAt = expireEnd
		result.ExpireAt = expireEnd
		// 新记录没有主键，就是本次要插的那条
		if row.ID == 0 {
			result.CoverFrom, result.CoverTo = coverFrom, coverTo
			continue
		}
		// 已有记录只有区间真的变了才回写
		if row.CoverFrom == coverFrom && row.CoverTo == coverTo {
			continue
		}
		row.CoverFrom, row.CoverTo = coverFrom, coverTo
		result.Shifts = append(result.Shifts, row)
	}
	return result, 0, nil
}

// parseGuardLevel 校验并转换大航海等级，L0 不是有效的航海档位
func parseGuardLevel(v int) (enum.BadgeType, bool) {
	level := enum.BadgeType(v)
	if level == enum.BadgeTypeL0 || !level.IsValid() {
		return enum.BadgeTypeL0, false
	}
	return level, true
}

// monthRange 把 "2006-01" 解析成 [月首零点, 次月首零点) 的秒级区间
func monthRange(month string) (from, to int64, err error) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return 0, 0, fmt.Errorf("月份格式应为 2006-01: %w", err)
	}
	start := timeutil.StartOfMonth(t.In(time.Local))
	return start.Unix(), timeutil.AddMonths(start, 1).Unix(), nil
}

// formatUnix 秒级时间戳 → 展示文本，0 输出空串
func formatUnix(ts int64) string {
	if ts == 0 {
		return ""
	}
	return timeutil.Format(ts)
}

// formatMonth 秒级时间戳 → "2006-01"，0 输出空串
func formatMonth(ts int64) string {
	if ts == 0 {
		return ""
	}
	return timeutil.FormatWithLayout(ts, "2006-01")
}

// formatCoverTo 名单区间的终点月首是「不含」的，展示时退回上一个月
func formatCoverTo(ts int64) string {
	if ts == 0 {
		return ""
	}
	return formatMonth(timeutil.AddMonths(time.Unix(ts, 0), -1).Unix())
}

// toRedeemedProducts 兑换汇总 → 出参，没兑换过也返回空切片而不是 nil，前端不必再判空
func toRedeemedProducts(list []order.MonthlyRedeemedItem) []RedeemedProduct {
	items := make([]RedeemedProduct, 0, len(list))
	for _, row := range list {
		items = append(items, RedeemedProduct{
			ProductName: row.ProductName,
			Quantity:    row.Quantity,
		})
	}
	return items
}

// toListPageItems 名单行 → 出参
func toListPageItems(list []live_guard_record.RosterItem, level enum.BadgeType) []ListPageItem {
	items := make([]ListPageItem, 0, len(list))
	for _, row := range list {
		items = append(items, ListPageItem{
			UID:            row.UID,
			Uname:          row.Uname,
			GuardLevel:     level,
			CoverFromMonth: formatMonth(row.CoverFrom),
			CoverToMonth:   formatCoverTo(row.CoverTo),
			OpenNum:        row.OpenNum,
			OpenAt:         formatUnix(row.OpenAt),
		})
	}
	return items
}
