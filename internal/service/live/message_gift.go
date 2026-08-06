package live

import (
	"context"
	"log"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// giftProcessor 处理礼物相关消息（SEND_GIFT / SEND_GIFT_V2 / GUARD_BUY / SUPER_CHAT_MESSAGE）
type giftProcessor struct {
	liveGiftRepo live_gift.Repository
}

func newGiftProcessor(liveGiftRepo live_gift.Repository) *giftProcessor {
	return &giftProcessor{liveGiftRepo: liveGiftRepo}
}

func (p *giftProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdSendGift, live.CmdSendGiftV2, live.CmdGuardBuy, live.CmdSuperDanmuMsg}
}

func (p *giftProcessor) Process(ctx context.Context, cmd live.Cmd, data any, roomID int64) error {
	var gift *model.LiveGift

	switch cmd {
	case live.CmdSendGift, live.CmdSendGiftV2:
		info, ok := data.(*live.SendGiftInfo)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.SendGiftInfo，实际 %T", data)
			return nil
		}
		gift = &model.LiveGift{
			RoomID:     roomID,
			UID:        info.UID,
			Uname:      info.Uname,
			GiftType:   enum.GiftTypeNormal,
			GiftID:     info.GiftID,
			GiftName:   info.GiftName,
			Price:      info.Price,
			Num:        info.Num,
			BadgeUID:   info.BadgeUID,
			BadgeName:  info.BadgeName,
			BadgeLevel: info.BadgeLevel,
			BadgeType:  enum.BadgeType(info.BadgeType),
			LiveID:     0,
			SendAt:     time.Now().Unix(),
			Original:   enum.Yes,
		}
		if info.BlindGift != nil {
			gift.Original = enum.No
			gift.OriginalGiftID = info.BlindGift.OriginalGiftID
			gift.OriginalGiftName = info.BlindGift.OriginalGiftName
			gift.OriginalGiftPrice = info.BlindGift.OriginalGiftPrice
		}
	case live.CmdGuardBuy:
		info, ok := data.(*live.GuardBuyInfo)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.GuardBuyInfo，实际 %T", data)
			return nil
		}
		gift = &model.LiveGift{
			RoomID:   roomID,
			UID:      info.UID,
			Uname:    info.Uname,
			GiftType: enum.GiftTypeGuard,
			GiftID:   info.GiftID,
			GiftName: info.GiftName,
			Price:    info.Price,
			Num:      info.Num,
			LiveID:   0,
			SendAt:   time.Now().Unix(),
			Original: enum.Yes,
		}
	case live.CmdSuperDanmuMsg:
		info, ok := data.(*live.SuperChatMessage)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.SuperChatMessage，实际 %T", data)
			return nil
		}
		gift = &model.LiveGift{
			RoomID:     roomID,
			UID:        info.UID,
			Uname:      info.Uname,
			GiftType:   enum.GiftTypeSuperChat,
			GiftID:     info.GiftID,
			GiftName:   info.GiftName,
			Price:      info.Price,
			Num:        1,
			Message:    info.Message,
			BadgeName:  info.BadgeName,
			BadgeLevel: info.BadgeLevel,
			BadgeType:  enum.BadgeType(info.BadgeType),
			LiveID:     0,
			SendAt:     time.Now().Unix(),
			Original:   enum.Yes,
		}
	}
	if gift != nil {
		if _, err := p.liveGiftRepo.Create(ctx, nil, gift); err != nil {
			log.Printf("[live.Gift] 礼物存储失败: %v", err)
			return err
		}
	}
	return nil
}
