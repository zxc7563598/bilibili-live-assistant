package live

import (
	"context"
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// giftProcessor 处理礼物相关消息（SEND_GIFT / SEND_GIFT_V2 / GUARD_BUY / SUPER_CHAT_MESSAGE）
type giftProcessor struct{}

func newGiftProcessor() *giftProcessor {
	return &giftProcessor{}
}

func (p *giftProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdSendGift, live.CmdSendGiftV2, live.CmdGuardBuy, live.CmdSuperDanmuMsg}
}

func (p *giftProcessor) Process(ctx context.Context, cmd live.Cmd, data any) error {
	switch cmd {
	case live.CmdSendGift, live.CmdSendGiftV2:
		info, ok := data.(*live.SendGiftInfo)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.SendGiftInfo，实际 %T", data)
			return nil
		}
		log.Printf("[live.Gift] 收到礼物 — 用户: %s(UID:%d), 礼物: %s(ID:%d), 数量: %d, 总价: %d分",
			info.Uname, info.UID, info.GiftName, info.GiftID, info.Num, info.Price)
		if info.BlindGift != nil {
			log.Printf("[live.Gift] 盲盒礼物详情 — 原始礼物: %s(ID:%d), 价格: %d分",
				info.BlindGift.OriginalGiftName, info.BlindGift.OriginalGiftID, info.BlindGift.OriginalGiftPrice)
		}
		// TODO: 存储礼物记录到数据库
		// TODO: 礼物感谢/自动回复
		// TODO: 礼物统计/排行榜更新

	case live.CmdGuardBuy:
		info, ok := data.(*live.GuardBuyInfo)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.GuardBuyInfo，实际 %T", data)
			return nil
		}
		log.Printf("[live.Gift] 大航海购买 — 用户: %s(UID:%d), 类型: %s, 数量: %d, 金额: %d分",
			info.Uname, info.UID, info.GiftName, info.Num, info.Price)
		// TODO: 存储大航海记录到数据库
		// TODO: 上船感谢/自动回复

	case live.CmdSuperDanmuMsg:
		info, ok := data.(*live.SuperChatMessage)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.SuperChatMessage，实际 %T", data)
			return nil
		}
		log.Printf("[live.Gift] 醒目留言 — 用户: %s(UID:%d), 金额: %d分, 留言: %s",
			info.Uname, info.UID, info.Price, info.Message)
		// TODO: 存储醒目留言记录到数据库
		// TODO: 醒目留言感谢/自动回复
	}
	return nil
}
