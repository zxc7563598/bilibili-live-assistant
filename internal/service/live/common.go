package live

import (
	"context"
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_session"
)

// aggregateSessionStats 将指定场次时间范围内的弹幕/礼物关联到该场次，并汇总统计数据写入 live_sessions
func aggregateSessionStats(
	ctx context.Context,
	session model.LiveSession,
	endAt int64,
	liveDanmuRepo live_danmu.Repository,
	liveGiftRepo live_gift.Repository,
	liveSessionRepo live_session.Repository,
) error {
	// 关联弹幕
	if err := liveDanmuRepo.UpdateLiveIDByRoomIDAndTimeRange(ctx, nil, session.StartAt, endAt, session.RoomID, session.ID); err != nil {
		log.Printf("[live.Danmu] 关联弹幕失败: %v", err)
		return err
	}
	// 关联礼物
	if err := liveGiftRepo.UpdateLiveIDByRoomIDAndTimeRange(ctx, nil, session.StartAt, endAt, session.RoomID, session.ID); err != nil {
		log.Printf("[live.Gift] 关联礼物失败: %v", err)
		return err
	}
	// 统计弹幕数量
	danmuCount, err := liveDanmuRepo.CountByRoomIDAndTimeRange(ctx, nil, session.StartAt, endAt, session.RoomID)
	if err != nil {
		log.Printf("[live.Danmu] 统计弹幕数量失败: %v", err)
		return err
	}
	// 统计礼物数量与收益
	giftCount, revenue, err := liveGiftRepo.CountAndRevenueByRoomIDAndTimeRange(ctx, nil, session.StartAt, endAt, session.RoomID)
	if err != nil {
		log.Printf("[live.Gift] 统计礼物数据失败: %v", err)
		return err
	}
	// 统计大航海数量
	guardCount, err := liveGiftRepo.CountGuardByRoomIDAndTimeRange(ctx, nil, session.StartAt, endAt, session.RoomID)
	if err != nil {
		log.Printf("[live.Gift] 统计大航海数量失败: %v", err)
		return err
	}
	// 统计醒目留言数量
	superChatCount, err := liveGiftRepo.CountSuperChatByRoomIDAndTimeRange(ctx, nil, session.StartAt, endAt, session.RoomID)
	if err != nil {
		log.Printf("[live.Gift] 统计醒目留言数量失败: %v", err)
		return err
	}
	// 更新统计数据
	if err := liveSessionRepo.UpdateStatsByID(ctx, nil, session.ID, model.LiveSessionUpdateStatsForm{
		DanmuCount:     &danmuCount,
		GiftCount:      &giftCount,
		GuardCount:     &guardCount,
		SuperChatCount: &superChatCount,
		TotalRevenue:   &revenue,
	}); err != nil {
		log.Printf("[live.Session] 更新统计数据失败 (ID:%d): %v", session.ID, err)
		return err
	}
	return nil
}
