package live

import (
	"context"
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// interactProcessor 处理用户互动消息（INTERACT_WORD_V2）
//
// 消息类型（MsgType）：
//
//	1 — 进入直播间
//	2 — 关注
//	3 — 分享
type interactProcessor struct {
	liveUserRepo live_user.Repository
}

func newInteractProcessor(liveUserRepo live_user.Repository) *interactProcessor {
	return &interactProcessor{liveUserRepo: liveUserRepo}
}

func (p *interactProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdInteractWord}
}

func (p *interactProcessor) Process(ctx context.Context, cmd live.Cmd, data any, roomID int64) error {
	info, ok := data.(*live.InteractWordV2Info)
	if !ok {
		log.Printf("[live.Interact] 数据类型断言失败，期望 *live.InteractWordV2Info，实际 %T", data)
		return nil
	}
	// 进入欢迎/关注感谢等自动回复
	switch info.MsgType {
	case 1: // 进入直播间

	case 2: // 关注

	case 3: // 分享

	}
	return nil
}
