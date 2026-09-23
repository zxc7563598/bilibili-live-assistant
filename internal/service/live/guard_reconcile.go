package live

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/room"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// guardListClient 本任务用到的 B站 接口能力，便于用假实现覆盖分页与容错分支
type guardListClient interface {
	GetGuardTopListPage(ctx context.Context, anchorUID, roomID int64, page int) (*room.GuardTopListPage, error)
}

const (
	// guardReconcileEarliest 每日首次执行的下界，落在 00:00 那一分钟的 tick 跳过
	guardReconcileEarliest = time.Minute
	// guardReconcileMaxAttempts 同一自然日的最大尝试次数，失败借下一次 tick 重试的封顶
	guardReconcileMaxAttempts = 5
	// guardReconcileMaxPages 页数上限，防接口返回异常页数导致循环不停
	guardReconcileMaxPages = 1000
)

// guardReconcilePageGapMin/Max 分页之间的随机停顿区间，写成变量以便测试置 0
var (
	guardReconcilePageGapMin = 1500 * time.Millisecond
	guardReconcilePageGapMax = 3500 * time.Millisecond
)

// guardReconcileState 每个自然日只成功跑一次的闸门，仅内存态，重启后当日会重跑一轮
type guardReconcileState struct {
	mu       sync.Mutex
	day      string
	attempts int
	done     bool
}

// begin 判断该自然日是否还要执行，并记一次尝试
func (g *guardReconcileState) begin(day string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.day != day {
		g.day, g.attempts, g.done = day, 0, false
	}
	if g.done || g.attempts >= guardReconcileMaxAttempts {
		return false
	}
	g.attempts++
	return true
}

// finish 记录本轮结果，成功即当日不再执行
func (g *guardReconcileState) finish(day string, ok bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if ok && g.day == day {
		g.done = true
	}
}

// ReconcileGuardExpire 定时任务：每天 00:01 之后按 B站 大航海名单校对用户的到期时间
func (s *Service) ReconcileGuardExpire(ctx context.Context) error {
	now := time.Now()
	if !guardReconcileDue(now) {
		return nil
	}
	day := timeutil.StartOfDay(now).Format(time.DateOnly)
	if !s.guardReconcile.begin(day) {
		return nil
	}
	err := s.syncGuardExpireOnce(ctx)
	s.guardReconcile.finish(day, err == nil)
	return err
}

// guardReconcileDue 判断当前时刻是否已过每日执行下界
func guardReconcileDue(now time.Time) bool {
	return !now.Before(timeutil.StartOfDay(now).Add(guardReconcileEarliest))
}

// syncGuardExpireOnce 单轮校对的编排：定房间与主播 → 抓名单 → 交给 liveuser 写库
func (s *Service) syncGuardExpireOnce(ctx context.Context) error {
	// 快照 client 与配置房间号，避免与 Logout/SwitchRoom 竞争
	s.mu.Lock()
	client := s.client
	roomID := s.roomID
	s.mu.Unlock()
	if client == nil || roomID <= 0 {
		log.Printf("[live.Guard] 未配置房间号，跳过本轮校对")
		return nil
	}

	// 主播 UID 与真实房间号：缓存为空时主动拉一次回填（配置里可能是短号）
	anchorUID, realRoomID := s.roomState.UID(), s.roomState.RoomID()
	if anchorUID == 0 || realRoomID == 0 {
		info, err := client.Room.GetRealRoomInfo(ctx, roomID)
		if err != nil {
			return fmt.Errorf("获取直播间信息失败: %w", err)
		}
		s.roomState.Update(info)
		anchorUID, realRoomID = info.UID, info.RoomID
	}
	if anchorUID <= 0 || realRoomID <= 0 {
		return fmt.Errorf("直播间信息不完整: anchor=%d room=%d", anchorUID, realRoomID)
	}

	log.Printf("[live.Guard] 开始校对大航海名单: room=%d anchor=%d", realRoomID, anchorUID)
	entries, err := s.fetchGuardEntries(ctx, client.Room, anchorUID, realRoomID)
	if err != nil {
		log.Printf("[live.Guard] 名单抓取不完整，本轮不做任何写库: %v", err)
		return err
	}
	res, err := s.liveUserSvc.ReconcileGuardExpire(ctx, entries)
	if err != nil {
		return err
	}
	log.Printf(
		"[live.Guard] 校对完成: 名单=%d 新注册=%d 抬升=%d 降级=%d 失效=%d 未变=%d 未知档位=%d 失败=%d",
		res.Inspected, res.Registered, res.Raised, res.Downgraded, res.Expired, res.Skipped, res.Unknown, res.Failed,
	)
	return nil
}

// fetchGuardEntries 抓取全部页并归一为名单
func (s *Service) fetchGuardEntries(ctx context.Context, c guardListClient, anchorUID, roomID int64) ([]liveuser.GuardEntry, error) {
	first, err := c.GetGuardTopListPage(ctx, anchorUID, roomID, 1)
	if err != nil {
		return nil, fmt.Errorf("大航海名单第 1 页抓取失败: %w", err)
	}
	// 空名单是合法结果：交给校对层把库里仍有效的档位清掉
	if first.Total == 0 {
		if first.TotalPage != 0 || len(first.Top3) > 0 || len(first.Items) > 0 {
			return nil, fmt.Errorf("大航海名单响应异常: num=0 但 page=%d top3=%d list=%d", first.TotalPage, len(first.Top3), len(first.Items))
		}
		return nil, nil
	}
	if first.TotalPage <= 0 || first.TotalPage > guardReconcileMaxPages {
		return nil, fmt.Errorf("大航海名单页数异常: page=%d num=%d", first.TotalPage, first.Total)
	}
	levelOf := make(map[int64]enum.BadgeType, first.Total)
	nameOf := make(map[int64]string, first.Total)
	merge := func(items []room.GuardTopListItem) {
		for _, item := range items {
			if item.UID <= 0 {
				continue
			}
			// guard_level 与 enum.BadgeType 数值恒等，非法值归一为 L0 表示档位未知
			level := enum.BadgeType(item.GuardLevel)
			if !level.IsValid() {
				level = enum.BadgeTypeL0
			}
			if old, ok := levelOf[item.UID]; ok {
				if level != enum.BadgeTypeL0 && (old == enum.BadgeTypeL0 || level < old) {
					levelOf[item.UID] = level
				}
			} else {
				levelOf[item.UID] = level
			}
			if item.Name != "" {
				nameOf[item.UID] = item.Name
			}
		}
	}
	merge(first.Top3)
	merge(first.Items)
	for page := 2; page <= first.TotalPage; page++ {
		if err := guardReconcilePause(ctx); err != nil {
			return nil, err
		}
		p, err := c.GetGuardTopListPage(ctx, anchorUID, roomID, page)
		if err != nil {
			return nil, fmt.Errorf("大航海名单第 %d/%d 页抓取失败: %w", page, first.TotalPage, err)
		}
		if p.Now != page {
			return nil, fmt.Errorf("大航海名单第 %d 页返回的 now=%d 与请求不一致", page, p.Now)
		}
		if len(p.Items) == 0 {
			return nil, fmt.Errorf("大航海名单第 %d/%d 页为空", page, first.TotalPage)
		}
		merge(p.Items)
	}
	if len(levelOf) < first.Total {
		return nil, fmt.Errorf("大航海名单抓取不完整: 去重后 %d 人 / 接口声明 %d 人", len(levelOf), first.Total)
	}
	entries := make([]liveuser.GuardEntry, 0, len(levelOf))
	for uid, level := range levelOf {
		entries = append(entries, liveuser.GuardEntry{UID: uid, Uname: nameOf[uid], Level: level})
	}
	return entries, nil
}

// guardReconcileGap 本次分页要停顿的时长，上下界相等时不随机
func guardReconcileGap() time.Duration {
	if guardReconcilePageGapMax <= guardReconcilePageGapMin {
		return guardReconcilePageGapMin
	}
	return guardReconcilePageGapMin + rand.N(guardReconcilePageGapMax-guardReconcilePageGapMin)
}

// guardReconcilePause 分页之间的随机停顿，防风控；ctx 取消时立即返回
func guardReconcilePause(ctx context.Context) error {
	timer := time.NewTimer(guardReconcileGap())
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
