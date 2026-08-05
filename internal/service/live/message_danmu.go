package live

import (
	"context"
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// danmuProcessor 处理弹幕消息（DANMU_MSG）
type danmuProcessor struct{}

func newDanmuProcessor() *danmuProcessor {
	return &danmuProcessor{}
}

func (p *danmuProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdDanmuMsg}
}

func (p *danmuProcessor) Process(ctx context.Context, cmd live.Cmd, data any) error {
	info, ok := data.(*live.DanmuMsgInfo)
	if !ok {
		log.Printf("[live.Danmu] 数据类型断言失败，期望 *live.DanmuMsgInfo，实际 %T", data)
		return nil
	}
	log.Printf("[live.Danmu] 弹幕 — 用户: %s(UID:%d), 内容: %s", info.Uname, info.UID, info.Msg)
	// TODO: 存储弹幕记录到数据库
	// TODO: 弹幕关键词检测/自动回复
	return nil
}
