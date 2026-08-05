package live

import (
	"context"
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// liveStatusProcessor 处理直播开始消息（LIVE）
type liveStatusProcessor struct{}

func newLiveStatusProcessor() *liveStatusProcessor {
	return &liveStatusProcessor{}
}

func (p *liveStatusProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdLiveStart}
}

func (p *liveStatusProcessor) Process(ctx context.Context, cmd live.Cmd, data any) error {
	info, ok := data.(*live.LiveInfo)
	if !ok {
		log.Printf("[live.LiveStatus] 数据类型断言失败，期望 *live.LiveInfo，实际 %T", data)
		return nil
	}
	log.Printf("[live.LiveStatus] 直播开始 — 开播平台: %s, 流标识: %s", info.LivePlatform, info.LiveKey)
	// TODO: 存储直播开始记录到数据库
	// TODO: 触发开播通知
	return nil
}
