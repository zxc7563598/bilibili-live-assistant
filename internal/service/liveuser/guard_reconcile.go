package liveuser

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// guardTiers 三个档位，高等级在前，与 GuardReconcileResult 的计数口径一致
var guardTiers = []enum.BadgeType{enum.BadgeTypeL3, enum.BadgeTypeL2, enum.BadgeTypeL1}

// ReconcileGuardExpire 按 B站 大航海名单校对用户的三个到期时间列
//
// 名单是权威事实，tomorrow 为明日 0 点、today 为今日 0 点：
//   - 名单内、库里没有 → 注册用户，该档位到期时间置 tomorrow
//   - 名单内、库里有 → 名单所示档位 < tomorrow 则置 tomorrow；
//     名单所示档位以外的列 >= tomorrow 则置 today（降级清理）
//   - 名单外、任一档位 >= tomorrow → 该列置 today
//
// 单条失败只记日志并计入 Failed，不中断整轮；只有初始查询失败才返回 error。
func (s *Service) ReconcileGuardExpire(ctx context.Context, entries []GuardEntry) (GuardReconcileResult, error) {
	return s.reconcileGuardExpireAt(ctx, entries, time.Now())
}

// reconcileGuardExpireAt 校对主体，now 注入只为可测试性
func (s *Service) reconcileGuardExpireAt(ctx context.Context, entries []GuardEntry, now time.Time) (GuardReconcileResult, error) {
	var res GuardReconcileResult
	todayStart := timeutil.StartOfDay(now).Unix()
	tomorrowStart := timeutil.StartOfDay(now).AddDate(0, 0, 1).Unix()

	// 名单归一：按 uid 去重，同一 uid 取最高档（guard_level 数值越小档位越高）
	levelOf := make(map[int64]enum.BadgeType, len(entries))
	nameOf := make(map[int64]string, len(entries))
	uids := make([]int64, 0, len(entries))
	for _, e := range entries {
		if e.UID <= 0 {
			continue
		}
		if old, ok := levelOf[e.UID]; ok {
			if e.Level.IsValid() && e.Level != enum.BadgeTypeL0 && (old == enum.BadgeTypeL0 || e.Level < old) {
				levelOf[e.UID] = e.Level
			}
		} else {
			levelOf[e.UID] = e.Level
			uids = append(uids, e.UID)
		}
		if e.Uname != "" {
			nameOf[e.UID] = e.Uname
		}
	}
	res.Inspected = len(uids)

	// 两条快照先全部读完，写入过程不再重读
	users, err := s.liveUserRepo.ListByUIDs(ctx, nil, uids)
	if err != nil {
		return res, fmt.Errorf("按 uid 批量查询用户失败: %w", err)
	}
	exist := make(map[int64]*model.LiveUser, len(users))
	for i := range users {
		exist[users[i].UID] = &users[i]
	}
	active, err := s.liveUserRepo.ListActiveGuard(ctx, nil, tomorrowStart)
	if err != nil {
		return res, fmt.Errorf("查询有效大航海用户失败: %w", err)
	}

	// 名单内：抬升名单所示档位，并清掉名单所示档位以外的残留
	for _, uid := range uids {
		level := levelOf[uid]
		if level == enum.BadgeTypeL0 {
			res.Unknown++
			continue
		}
		column, err := guardExpireField(level)
		if err != nil {
			res.Unknown++
			continue
		}
		user, ok := exist[uid]
		if !ok {
			id, err := s.EnsureUser(ctx, uid, guardEntryName(uid, nameOf))
			if err != nil {
				res.Failed++
				log.Printf("[liveuser.Guard] 注册用户失败: uid=%d err=%v", uid, err)
				continue
			}
			user = &model.LiveUser{ID: id, UID: uid}
			res.Registered++
		}
		changed, err := s.liveUserRepo.UpdateGuardExpireIfBelow(ctx, nil, user.ID, column, tomorrowStart, tomorrowStart)
		if err != nil {
			res.Failed++
			log.Printf("[liveuser.Guard] 抬升到期时间失败: uid=%d column=%s err=%v", uid, column, err)
			continue
		}
		if changed {
			res.Raised++
		} else {
			res.Skipped++
		}
		cleared, failed := s.clearTiers(ctx, user, guardOtherTiers(level), tomorrowStart, todayStart)
		res.Downgraded += cleared
		res.Failed += failed
	}

	// 名单外：库里仍认为有效的档位一律失效
	for i := range active {
		user := &active[i]
		if _, inList := levelOf[user.UID]; inList {
			continue
		}
		cleared, failed := s.clearTiers(ctx, user, guardTiers, tomorrowStart, todayStart)
		res.Expired += cleared
		res.Failed += failed
	}
	return res, nil
}

// guardEntryName 名单缺昵称时的兜底，空串会把库里已有的昵称清掉
func guardEntryName(uid int64, nameOf map[int64]string) string {
	if name := nameOf[uid]; name != "" {
		return name
	}
	return strconv.FormatInt(uid, 10)
}

// guardOtherTiers keep 档位以外的两个档位
func guardOtherTiers(keep enum.BadgeType) []enum.BadgeType {
	others := make([]enum.BadgeType, 0, len(guardTiers)-1)
	for _, level := range guardTiers {
		if level != keep {
			others = append(others, level)
		}
	}
	return others
}

// clearTiers 把指定档位中「到期时间仍在明日」的列失效为今日 0 点，
// 返回失效档位数与失败次数；期间值被改过（例如用户刚续费）的那一档放弃、不计入两者
func (s *Service) clearTiers(ctx context.Context, user *model.LiveUser, levels []enum.BadgeType, tomorrowStart, todayStart int64) (cleared, failed int) {
	for _, level := range levels {
		value := guardExpireValue(level, user)
		if value == nil || *value < tomorrowStart {
			continue
		}
		column, err := guardExpireField(level)
		if err != nil {
			continue
		}
		changed, err := s.liveUserRepo.UpdateGuardExpireIfEqual(ctx, nil, user.ID, column, *value, todayStart)
		if err != nil {
			failed++
			log.Printf("[liveuser.Guard] 失效到期时间失败: uid=%d column=%s err=%v", user.UID, column, err)
			continue
		}
		if changed {
			cleared++
		}
	}
	return cleared, failed
}
