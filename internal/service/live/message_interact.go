package live

import (
	"context"
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// interactProcessor 处理用户互动消息（INTERACT_WORD_V2）
//
// 消息类型（MsgType）：
//
//	1 — 进入直播间
//	2 — 关注
//	3 — 分享
type interactProcessor struct{}

func newInteractProcessor() *interactProcessor {
	return &interactProcessor{}
}

func (p *interactProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdInteractWord}
}

func (p *interactProcessor) Process(ctx context.Context, cmd live.Cmd, data any) error {
	info, ok := data.(*live.InteractWordV2Info)
	if !ok {
		log.Printf("[live.Interact] 数据类型断言失败，期望 *live.InteractWordV2Info，实际 %T", data)
		return nil
	}

	msgTypeText := "未知"
	switch info.MsgType {
	case 1:
		msgTypeText = "进入直播间"
	case 2:
		msgTypeText = "关注"
	case 3:
		msgTypeText = "分享"
	}

	log.Printf("[live.Interact] 用户互动 — 用户: %s(UID:%d), 类型: %s, 房间号: %d",
		info.Uname, info.UID, msgTypeText, info.RoomID)
	// TODO: 存储互动记录到数据库
	// TODO: 进入欢迎/关注感谢等自动回复
	return nil
}
