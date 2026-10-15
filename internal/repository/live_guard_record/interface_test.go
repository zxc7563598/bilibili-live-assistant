package live_guard_record

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"gorm.io/gorm"
)

// 名单归属只按月：区间两端都是月首，所以区间求交的结果就是月份集合。
//
// 数据：舰长 7 月 29 日开 3 份（覆盖 7/8/9 月），同月 30 日再开 1 份（顺延到 10 月）；
// 提督单开一份，用来证明两个档位互不串门。
func TestListRosterPageMonthMembership(t *testing.T) {
	repo := newGuardRepo(t)
	ctx := context.Background()
	for _, row := range []model.LiveGuardRecord{
		guardRec(t, 1, enum.BadgeTypeL1, "2026-07-29 10:06:22", 3, "2026-07-01", "2026-10-01"),
		guardRec(t, 1, enum.BadgeTypeL1, "2026-07-30 09:00:00", 1, "2026-10-01", "2026-11-01"),
		guardRec(t, 2, enum.BadgeTypeL2, "2026-07-15 12:00:00", 1, "2026-07-01", "2026-08-01"),
	} {
		row := row
		if _, err := repo.Create(ctx, nil, &row); err != nil {
			t.Fatalf("写入台账失败: %v", err)
		}
	}

	tests := []struct {
		name      string
		month     string
		level     enum.BadgeType
		wantUIDs  []int64
		wantTotal int64
	}{
		{"开通前的月份查不到", "2026-06", enum.BadgeTypeL1, nil, 0},
		{"开通当月命中", "2026-07", enum.BadgeTypeL1, []int64{1}, 1},
		{"未再开通也按顺延命中", "2026-08", enum.BadgeTypeL1, []int64{1}, 1},
		{"顺延的最后一个月仍命中", "2026-09", enum.BadgeTypeL1, []int64{1}, 1},
		{"同月第二次开通的份排到 10 月", "2026-10", enum.BadgeTypeL1, []int64{1}, 1},
		{"区间终点是「不含」，11 月不再命中", "2026-11", enum.BadgeTypeL1, nil, 0},
		{"另一档位独立成名单", "2026-07", enum.BadgeTypeL2, []int64{2}, 1},
		{"另一档位不在舰长的名单里", "2026-08", enum.BadgeTypeL2, nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to := monthRange(t, tt.month)
			level := int(tt.level)
			list, total, err := repo.ListRosterPage(ctx, nil, model.LiveGuardRecordListPageQuery{
				GuardLevel: &level,
				MonthFrom:  from,
				MonthTo:    to,
				Limit:      100,
			})
			if err != nil {
				t.Fatalf("ListRosterPage 失败: %v", err)
			}
			if total != tt.wantTotal {
				t.Errorf("总人数 = %d, want %d", total, tt.wantTotal)
			}
			if len(list) != len(tt.wantUIDs) {
				t.Fatalf("名单行数 = %d, want %d", len(list), len(tt.wantUIDs))
			}
			for i, uid := range tt.wantUIDs {
				if list[i].UID != uid {
					t.Errorf("第 %d 行 uid = %d, want %d", i+1, list[i].UID, uid)
				}
			}
		})
	}
}

// 名单区间取「覆盖查询月的连续覆盖段」：分次续费连排出来的月份并成一段，
// 断档处断开，档位之间不并。一次买多份与分次续费买出来的跨度应当一致。
func TestListRosterPageCoverRange(t *testing.T) {
	repo := newGuardRepo(t)
	ctx := context.Background()
	for _, row := range []model.LiveGuardRecord{
		// 7 月一次买 3 份，同月再续 1 份，连排到 10 月
		guardRec(t, 1, enum.BadgeTypeL1, "2026-07-29 10:06:22", 3, "2026-07-01", "2026-10-01"),
		guardRec(t, 1, enum.BadgeTypeL1, "2026-07-30 09:00:00", 1, "2026-10-01", "2026-11-01"),
		// 8 月断档：7 月开一份，9 月连开两份排到 10 月
		guardRec(t, 2, enum.BadgeTypeL1, "2026-07-10 12:00:00", 1, "2026-07-01", "2026-08-01"),
		guardRec(t, 2, enum.BadgeTypeL1, "2026-09-10 12:00:00", 1, "2026-09-01", "2026-10-01"),
		guardRec(t, 2, enum.BadgeTypeL1, "2026-09-11 12:00:00", 1, "2026-10-01", "2026-11-01"),
		// 提督单独成段，不并进同一用户的舰长跨度
		guardRec(t, 1, enum.BadgeTypeL2, "2026-07-15 12:00:00", 2, "2026-07-01", "2026-09-01"),
	} {
		row := row
		if _, err := repo.Create(ctx, nil, &row); err != nil {
			t.Fatalf("写入台账失败: %v", err)
		}
	}

	tests := []struct {
		name     string
		uid      int64
		level    enum.BadgeType
		month    string
		wantFrom string
		wantTo   string
	}{
		{"一次买多份的跨度照旧", 1, enum.BadgeTypeL1, "2026-07", "2026-07-01", "2026-11-01"},
		{"中段月份向两端延伸", 1, enum.BadgeTypeL1, "2026-09", "2026-07-01", "2026-11-01"},
		{"分次续费的月份并成一段", 1, enum.BadgeTypeL1, "2026-10", "2026-07-01", "2026-11-01"},
		{"断档前只算断档前那段", 2, enum.BadgeTypeL1, "2026-07", "2026-07-01", "2026-08-01"},
		{"断档后从断档后那段起算", 2, enum.BadgeTypeL1, "2026-09", "2026-09-01", "2026-11-01"},
		{"同一用户另一档位不并进来", 1, enum.BadgeTypeL2, "2026-08", "2026-07-01", "2026-09-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to := monthRange(t, tt.month)
			level, uid := int(tt.level), tt.uid
			query := model.LiveGuardRecordListPageQuery{
				GuardLevel: &level,
				MonthFrom:  from,
				MonthTo:    to,
				UID:        &uid,
				Limit:      100,
			}
			list, _, err := repo.ListRosterPage(ctx, nil, query)
			if err != nil {
				t.Fatalf("ListRosterPage 失败: %v", err)
			}
			if len(list) != 1 {
				t.Fatalf("名单行数 = %d, want 1", len(list))
			}
			assertCover(t, list[0], tt.wantFrom, tt.wantTo)

			// 导出走 RosterChunk，与列表必须是同一份区间
			chunk, err := repo.RosterChunk(ctx, nil, query, 0, 10)
			if err != nil {
				t.Fatalf("RosterChunk 失败: %v", err)
			}
			if len(chunk) != 1 {
				t.Fatalf("导出分块行数 = %d, want 1", len(chunk))
			}
			assertCover(t, chunk[0], tt.wantFrom, tt.wantTo)
		})
	}
}

// 没有任何一段覆盖查询月时返回 false（查询条件与台账对不上）
func TestCoverSpanNoMatch(t *testing.T) {
	rows := []model.LiveGuardRecord{
		{CoverFrom: day(t, "2026-07-01"), CoverTo: day(t, "2026-08-01")},
	}
	from, to := monthRange(t, "2026-09")
	if _, _, ok := coverSpan(rows, from, to); ok {
		t.Errorf("查询月不在任何一段里，不该返回区间")
	}
	if _, _, ok := coverSpan(nil, from, to); ok {
		t.Errorf("空台账不该返回区间")
	}
}

func assertCover(t *testing.T, item RosterItem, wantFrom, wantTo string) {
	t.Helper()
	if item.CoverFrom != day(t, wantFrom) || item.CoverTo != day(t, wantTo) {
		t.Errorf("覆盖区间 = %s ~ %s, want %s ~ %s",
			stamp(item.CoverFrom), stamp(item.CoverTo), wantFrom, wantTo)
	}
}

// 当月开通份数按份数累加（同月开两次算两份），没开通的月份不返回统计
func TestListOpenStats(t *testing.T) {
	repo := newGuardRepo(t)
	ctx := context.Background()
	for _, row := range []model.LiveGuardRecord{
		guardRec(t, 1, enum.BadgeTypeL1, "2026-07-29 10:06:22", 3, "2026-07-01", "2026-10-01"),
		guardRec(t, 1, enum.BadgeTypeL1, "2026-07-30 09:00:00", 1, "2026-10-01", "2026-11-01"),
	} {
		row := row
		if _, err := repo.Create(ctx, nil, &row); err != nil {
			t.Fatalf("写入台账失败: %v", err)
		}
	}

	level := int(enum.BadgeTypeL1)
	from, to := monthRange(t, "2026-07")
	stats, err := repo.ListOpenStats(ctx, nil, model.LiveGuardRecordOpenStatsQuery{
		GuardLevel: &level,
		OpenFrom:   from,
		OpenTo:     to,
		UIDs:       []int64{1},
	})
	if err != nil {
		t.Fatalf("ListOpenStats 失败: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("统计行数 = %d, want 1", len(stats))
	}
	if stats[0].OpenNum != 4 {
		t.Errorf("当月开通份数 = %d, want 4", stats[0].OpenNum)
	}
	if stats[0].OpenAt != moment(t, "2026-07-29 10:06:22") {
		t.Errorf("当月最早开通 = %s, want 2026-07-29 10:06:22", stamp(stats[0].OpenAt))
	}

	// 顺延来的月份没有新开通，不产生统计行
	from, to = monthRange(t, "2026-08")
	stats, err = repo.ListOpenStats(ctx, nil, model.LiveGuardRecordOpenStatsQuery{
		GuardLevel: &level,
		OpenFrom:   from,
		OpenTo:     to,
		UIDs:       []int64{1},
	})
	if err != nil {
		t.Fatalf("ListOpenStats 失败: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("顺延月份不该有开通统计, got %d 行", len(stats))
	}
}

func newGuardRepo(t *testing.T) Repository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "roster.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开 sqlite 失败: %v", err)
	}
	if err := db.AutoMigrate(&model.LiveGuardRecord{}); err != nil {
		t.Fatalf("建表失败: %v", err)
	}
	return New(db)
}

func guardRec(t *testing.T, uid int64, level enum.BadgeType, openAt string, num int64, coverFrom, coverTo string) model.LiveGuardRecord {
	t.Helper()
	return model.LiveGuardRecord{
		RoomID:     1,
		UID:        uid,
		Uname:      "tester",
		GuardLevel: level,
		Num:        num,
		OpenAt:     moment(t, openAt),
		CoverFrom:  day(t, coverFrom),
		CoverTo:    day(t, coverTo),
	}
}

// monthRange 把 "2006-01" 解析成 [月首零点, 次月首零点) 的秒级区间
func monthRange(t *testing.T, month string) (int64, int64) {
	t.Helper()
	start, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		t.Fatalf("解析月份 %s 失败: %v", month, err)
	}
	return start.Unix(), start.AddDate(0, 1, 0).Unix()
}

func day(t *testing.T, s string) int64 {
	t.Helper()
	ts, err := time.ParseInLocation(time.DateOnly, s, time.Local)
	if err != nil {
		t.Fatalf("解析日期 %s 失败: %v", s, err)
	}
	return ts.Unix()
}

func moment(t *testing.T, s string) int64 {
	t.Helper()
	ts, err := time.ParseInLocation(time.DateTime, s, time.Local)
	if err != nil {
		t.Fatalf("解析时间 %s 失败: %v", s, err)
	}
	return ts.Unix()
}

func stamp(ts int64) string {
	if ts == 0 {
		return "空"
	}
	return time.Unix(ts, 0).Format(time.DateTime)
}
