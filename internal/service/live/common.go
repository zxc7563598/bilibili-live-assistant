package live

import (
	"context"
	"log"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_session"
)

// aggregateSessionStats 将指定场次时间范围内的弹幕/礼物关联到该场次，并汇总统计数据写入 live_sessions
func aggregateSessionStats(ctx context.Context, session model.LiveSession, endAt int64, liveDanmuRepo live_danmu.Repository, liveGiftRepo live_gift.Repository, liveSessionRepo live_session.Repository) error {
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

// endSessionAndAggregate 结束指定会话并聚合统计数据
func (s *Service) endSessionAndAggregate(ctx context.Context, session model.LiveSession, endAt int64, endReason enum.EndReason, endSource enum.EndSource, endDetail string) error {
	if err := s.liveSessionRepo.UpdateEndByID(ctx, nil, session.ID, model.LiveSessionUpdateEndForm{
		EndAt:     &endAt,
		EndReason: &endReason,
		EndSource: &endSource,
		EndDetail: &endDetail,
	}); err != nil {
		return err
	}
	return aggregateSessionStats(ctx, session, endAt, s.liveDanmuRepo, s.liveGiftRepo, s.liveSessionRepo)
}

// syncSessionsOnStart 在监听启动后同步会话状态
func (s *Service) syncSessionsOnStart(ctx context.Context, roomID int64) {
	now := time.Now().Unix()
	// 结束所有非监听房间的活跃会话
	activeSessions, err := s.liveSessionRepo.ListActive(ctx, nil)
	if err != nil {
		log.Printf("[live.Service] 同步会话：获取活跃会话失败: %v", err)
		return
	}
	for _, session := range activeSessions {
		if session.RoomID == roomID {
			continue // 跳过监听房间，后续根据 LiveStatus 处理
		}
		if err := s.endSessionAndAggregate(ctx, session, now, enum.EndReasonNormal, enum.EndSourceManual, "监听房间切换，系统自动结束"); err != nil {
			log.Printf("[live.Service] 同步会话：结束非监听房间会话失败 (ID:%d, RoomID:%d): %v", session.ID, session.RoomID, err)
		}
	}
	// 获取监听房间的实际直播状态
	roomInfo, err := s.client.Room.GetRealRoomInfo(ctx, roomID)
	if err != nil {
		log.Printf("[live.Service] 同步会话：获取直播间信息失败 (RoomID:%d): %v", roomID, err)
		return
	}
	// 同步直播间状态缓存
	s.roomState.Update(roomInfo)
	// 根据实际状态修正监听房间的会话
	roomActiveSessions, err := s.liveSessionRepo.ListActiveByRoomID(ctx, nil, roomID)
	if err != nil {
		log.Printf("[live.Service] 同步会话：获取房间活跃会话失败 (RoomID:%d): %v", roomID, err)
		return
	}
	isLive := roomInfo.LiveStatus == 1
	hasActiveSession := len(roomActiveSessions) > 0
	if isLive && !hasActiveSession {
		// 实际在直播，但没有活跃记录
		session := &model.LiveSession{
			RoomID:      roomID,
			UID:         roomInfo.UID,
			StartAt:     now,
			StartSource: enum.StartSourcePolling,
		}
		if _, err := s.liveSessionRepo.Create(ctx, nil, session); err != nil {
			log.Printf("[live.Service] 同步会话：补录直播记录失败 (RoomID:%d): %v", roomID, err)
		} else {
			log.Printf("[live.Service] 同步会话：检测到房间 %d 正在直播，已补录直播记录", roomID)
		}
	}
	if !isLive && hasActiveSession {
		// 实际未直播，但有活跃记录
		for _, session := range roomActiveSessions {
			if err := s.endSessionAndAggregate(ctx, session, now, enum.EndReasonNormal, enum.EndSourcePolling, "启动监听时轮询检测到直播已结束"); err != nil {
				log.Printf("[live.Service] 同步会话：结束当前房间会话失败 (ID:%d): %v", session.ID, err)
			}
		}
		log.Printf("[live.Service] 同步会话：检测到房间 %d 已下播，已结束 %d 条活跃记录", roomID, len(roomActiveSessions))
	}
}
