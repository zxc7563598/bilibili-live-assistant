package live

import (
	"context"
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// liveEndProcessor 处理直播结束相关消息（CUT_OFF / ROOM_LOCK / PREPARING）
type liveEndProcessor struct{}

func newLiveEndProcessor() *liveEndProcessor {
	return &liveEndProcessor{}
}

func (p *liveEndProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdLiveCutOff, live.CmdLiveRoomLock, live.CmdLiveEnd}
}

func (p *liveEndProcessor) Process(ctx context.Context, cmd live.Cmd, data any) error {
	switch cmd {
	case live.CmdLiveCutOff:
		info, ok := data.(*live.CutOffInfo)
		if !ok {
			log.Printf("[live.LiveEnd] 数据类型断言失败，期望 *live.CutOffInfo，实际 %T", data)
			return nil
		}
		log.Printf("[live.LiveEnd] 直播被超管切断 — 房间号: %d, 原因: %s", info.RoomID, info.Msg)
		// TODO: 存储切断记录
		// TODO: 触发通知

	case live.CmdLiveRoomLock:
		info, ok := data.(*live.RoomLockInfo)
		if !ok {
			log.Printf("[live.LiveEnd] 数据类型断言失败，期望 *live.RoomLockInfo，实际 %T", data)
			return nil
		}
		log.Printf("[live.LiveEnd] 直播间被封 — 房间号: %d", info.RoomID)
		// TODO: 存储封禁记录
		// TODO: 触发通知

	case live.CmdLiveEnd:
		info, ok := data.(*live.PreparingInfo)
		if !ok {
			log.Printf("[live.LiveEnd] 数据类型断言失败，期望 *live.PreparingInfo，实际 %T", data)
			return nil
		}
		log.Printf("[live.LiveEnd] 直播结束（下播） — 房间号: %s", info.RoomID)
		// TODO: 存储下播记录
		// TODO: 触发下播通知
	}
	return nil
}
