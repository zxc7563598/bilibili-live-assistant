package live

import (
	"context"
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// pkProcessor 处理 PK 相关消息（PK_BATTLE_PRE_NEW）
type pkProcessor struct{}

func newPkProcessor() *pkProcessor {
	return &pkProcessor{}
}

func (p *pkProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdPkStart}
}

func (p *pkProcessor) Process(ctx context.Context, cmd live.Cmd, data any) error {
	info, ok := data.(*live.PkBattlePreNewInfo)
	if !ok {
		log.Printf("[live.PK] 数据类型断言失败，期望 *live.PkBattlePreNewInfo，实际 %T", data)
		return nil
	}
	log.Printf("[live.PK] PK即将开始 — 对手: %s(UID:%d), 房间号: %d, PkID: %d, 对战类型: %d",
		info.Uname, info.UID, info.RoomID, info.PkID, info.BattleType)
	// TODO: 存储PK记录到数据库
	// TODO: PK开始自动提醒
	return nil
}
