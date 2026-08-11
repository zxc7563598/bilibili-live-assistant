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
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// liveStatusProcessor — 处理直播开始消息（LIVE）
type liveStatusProcessor struct {
	liveSessionRepo live_session.Repository
	liveDanmuRepo   live_danmu.Repository
	liveGiftRepo    live_gift.Repository
	roomState       *RoomState
}

func newLiveStatusProcessor(liveSessionRepo live_session.Repository, liveDanmuRepo live_danmu.Repository, liveGiftRepo live_gift.Repository, roomState *RoomState) *liveStatusProcessor {
	return &liveStatusProcessor{
		liveSessionRepo: liveSessionRepo,
		liveDanmuRepo:   liveDanmuRepo,
		liveGiftRepo:    liveGiftRepo,
		roomState:       roomState,
	}
}

func (p *liveStatusProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdLiveStart}
}

// endSession 结束指定会话并聚合统计数据
func (p *liveStatusProcessor) endSession(ctx context.Context, session model.LiveSession, endAt int64, endReason enum.EndReason, endSource enum.EndSource, endDetail string) error {
	if err := p.liveSessionRepo.UpdateEndByID(ctx, nil, session.ID, model.LiveSessionUpdateEndForm{
		EndAt:     &endAt,
		EndReason: &endReason,
		EndSource: &endSource,
		EndDetail: &endDetail,
	}); err != nil {
		return err
	}
	return aggregateSessionStats(ctx, session, endAt, p.liveDanmuRepo, p.liveGiftRepo, p.liveSessionRepo)
}

func (p *liveStatusProcessor) Process(ctx context.Context, cmd live.Cmd, data any, roomID int64) error {
	info, ok := data.(*live.LiveInfo)
	if !ok {
		log.Printf("[live.LiveStatus] 数据类型断言失败，期望 *live.LiveInfo，实际 %T", data)
		return nil
	}
	// 同步直播间状态缓存
	p.roomState.SetLiveStatus(1)
	now := time.Now().Unix()
	// 所有未下播记录转换为下播，记录为轮询下播
	activeSessions, err := p.liveSessionRepo.ListActive(ctx, nil)
	if err != nil {
		log.Printf("[live.Session] 获取未下播记录失败: %v", err)
		return err
	}
	for _, s := range activeSessions {
		if err := p.endSession(ctx, s, now, enum.EndReasonNormal, enum.EndSourcePolling, "系统检测到新的直播开始"); err != nil {
			log.Printf("[live.Session] 更新下播信息失败 (ID:%d): %v", s.ID, err)
			return err
		}
	}
	// 存储直播开始记录到数据库
	session := &model.LiveSession{
		RoomID:       roomID,
		UID:          0,
		LiveKey:      info.LiveKey,
		LivePlatform: info.LivePlatform,
		StartAt:      now,
		StartSource:  enum.StartSourceEvent,
	}
	if _, err := p.liveSessionRepo.Create(ctx, nil, session); err != nil {
		log.Printf("[live.Session] 直播记录存储失败: %v", err)
		return err
	}
	return nil
}

// liveEndProcessor — 处理直播结束相关消息（CUT_OFF / ROOM_LOCK / PREPARING）
type liveEndProcessor struct {
	liveSessionRepo live_session.Repository
	liveDanmuRepo   live_danmu.Repository
	liveGiftRepo    live_gift.Repository
	roomState       *RoomState
}

func newLiveEndProcessor(liveSessionRepo live_session.Repository, liveDanmuRepo live_danmu.Repository, liveGiftRepo live_gift.Repository, roomState *RoomState) *liveEndProcessor {
	return &liveEndProcessor{
		liveSessionRepo: liveSessionRepo,
		liveDanmuRepo:   liveDanmuRepo,
		liveGiftRepo:    liveGiftRepo,
		roomState:       roomState,
	}
}

func (p *liveEndProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdLiveCutOff, live.CmdLiveRoomLock, live.CmdLiveEnd}
}

// endSession 结束指定会话并聚合统计数据
func (p *liveEndProcessor) endSession(ctx context.Context, session model.LiveSession, endAt int64, endReason enum.EndReason, endSource enum.EndSource, endDetail string) error {
	if err := p.liveSessionRepo.UpdateEndByID(ctx, nil, session.ID, model.LiveSessionUpdateEndForm{
		EndAt:     &endAt,
		EndReason: &endReason,
		EndSource: &endSource,
		EndDetail: &endDetail,
	}); err != nil {
		return err
	}
	return aggregateSessionStats(ctx, session, endAt, p.liveDanmuRepo, p.liveGiftRepo, p.liveSessionRepo)
}

func (p *liveEndProcessor) Process(ctx context.Context, cmd live.Cmd, data any, roomID int64) error {
	// 根据命令字确定下播原因与日志描述
	var endReason enum.EndReason
	var logDesc string
	switch cmd {
	case live.CmdLiveCutOff:
		info, ok := data.(*live.CutOffInfo)
		if !ok {
			log.Printf("[live.LiveEnd] 数据类型断言失败，期望 *live.CutOffInfo，实际 %T", data)
			return nil
		}
		endReason = enum.EndReasonForced
		logDesc = info.Msg
	case live.CmdLiveRoomLock:
		_, ok := data.(*live.RoomLockInfo)
		if !ok {
			log.Printf("[live.LiveEnd] 数据类型断言失败，期望 *live.RoomLockInfo，实际 %T", data)
			return nil
		}
		endReason = enum.EndReasonBanned
		logDesc = "直播间被封"
	case live.CmdLiveEnd:
		_, ok := data.(*live.PreparingInfo)
		if !ok {
			log.Printf("[live.LiveEnd] 数据类型断言失败，期望 *live.PreparingInfo，实际 %T", data)
			return nil
		}
		endReason = enum.EndReasonNormal
		logDesc = "正常下播"
	default:
		return nil
	}
	// 同步直播间状态缓存
	p.roomState.SetLiveStatus(0)
	// 所有未下播记录设置为下播
	now := time.Now().Unix()
	activeSessions, err := p.liveSessionRepo.ListActive(ctx, nil)
	if err != nil {
		log.Printf("[live.Session] 获取未下播记录失败: %v", err)
		return err
	}
	for _, s := range activeSessions {
		if err := p.endSession(ctx, s, now, endReason, enum.EndSourceEvent, logDesc); err != nil {
			log.Printf("[live.Session] 更新下播信息失败 (ID:%d): %v", s.ID, err)
			return err
		}
	}
	return nil
}
