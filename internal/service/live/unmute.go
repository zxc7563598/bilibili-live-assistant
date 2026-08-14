package live

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_blacklist"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
)

// UnmuteDueUsers 定时任务：解禁所有已到解禁时间但仍处于禁言状态的黑名单用户
func (s *Service) UnmuteDueUsers(ctx context.Context) error {
	// 快照 client，避免与 Logout 重赋值 s.client 产生数据竞争
	s.mu.Lock()
	client := s.client
	s.mu.Unlock()

	csrf, err := client.CSRF()
	if err != nil {
		log.Printf("[live.Unmute] 机器人未登录或缺少 CSRF，跳过解禁任务: %v", err)
		return nil
	}
	list, err := s.LiveUserBlacklistRepo.ListExpiredMuted(ctx, nil, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("查询到期黑名单失败: %w", err)
	}
	for i := range list {
		unmuteUser(ctx, client, s.LiveUserBlacklistRepo, &list[i], csrf)
	}
	return nil
}

// unmuteUser 根据黑名单记录解除对应直播间禁言并更新解禁结果。
//
// 与礼物赎回逻辑共用，处理规则一致：
//   - 禁言列表中找不到该用户 → 标记为 NotFound（失败次数 +1）
//   - 解禁失败 → 失败次数 +1，累计达到 unmuteFailLimit 后标记为 UnmuteFailed
//   - 解禁成功 → 标记为 Unmuted（失败次数保持不变）
func unmuteUser(ctx context.Context, client *bilibili.Client, repo live_user_blacklist.Repository, black *model.LiveUserBlacklist, csrf string) {
	// 遍历直播间禁言列表，找到对应用户的禁言记录 ID
	blackID := int64(0)
	page := int64(1)
	for {
		silentList, silentListErr := client.Room.GetSilentUserList(ctx, black.RoomID, page, csrf)
		if silentListErr != nil {
			log.Printf("[live.Unmute] 禁言列表第%d页失败: %v", page, silentListErr)
			break
		}
		for _, item := range silentList.Items {
			if item.UID == black.UID {
				blackID = item.ID
				break
			}
		}
		if blackID != 0 || page >= int64(silentList.TotalPage) {
			break
		}
		page++
	}
	// 禁言列表中未找到该用户
	if blackID == 0 {
		if err := repo.UpdateUnmuteResult(ctx, nil, black.ID, enum.MuteStatusNotFound, black.UnmuteFailCount+1); err != nil {
			log.Printf("[live.Unmute] 更新黑名单解禁结果失败: %v", err)
		}
		return
	}
	// 解除直播间禁言
	if err := client.Room.DelSilentUser(ctx, black.RoomID, blackID, csrf); err != nil {
		// 解禁失败：失败次数 +1，累计达到阈值后标记为解禁失败
		failCount := black.UnmuteFailCount + 1
		status := black.Status
		if failCount >= unmuteFailLimit {
			status = enum.MuteStatusUnmuteFailed
		}
		if updateErr := repo.UpdateUnmuteResult(ctx, nil, black.ID, status, failCount); updateErr != nil {
			log.Printf("[live.Unmute] 更新黑名单解禁结果失败: %v", updateErr)
		}
		return
	}
	// 解禁成功：状态变更为已解禁，失败次数保持不变
	if err := repo.UpdateUnmuteResult(ctx, nil, black.ID, enum.MuteStatusUnmuted, black.UnmuteFailCount); err != nil {
		log.Printf("[live.Unmute] 更新黑名单解禁结果失败: %v", err)
	}
}
