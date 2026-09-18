package livegift

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livegift"
)

// secondsPerDay 一天的秒数，用于将日期范围结束时间戳推到当天最后一秒（23:59:59）
const secondsPerDay = 24 * 60 * 60

// Handler 礼物列表 HTTP 接口处理器
type Handler struct {
	livegiftSvc *livegift.Service
}

// New 创建 Handler 实例
func New(livegiftSvc *livegift.Service) *Handler {
	return &Handler{livegiftSvc: livegiftSvc}
}

func toLiveGiftListItems(list []livegift.ListPageItem) []resp.LiveGiftListPageItem {
	res := make([]resp.LiveGiftListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.LiveGiftListPageItem{
			ID:         v.ID,
			UID:        v.UID,
			Uname:      v.Uname,
			GiftName:   v.GiftName,
			Price:      v.Price,
			Num:        v.Num,
			Message:    v.Message,
			BadgeName:  v.BadgeName,
			BadgeLevel: v.BadgeLevel,
			BadgeType:  v.BadgeType,
			SendAt:     v.SendAt,
		})
	}
	return res
}

func toLiveGiftBlindBoxListItems(list []livegift.BlindBoxListPageItem) []resp.LiveGiftBlindBoxListPageItem {
	res := make([]resp.LiveGiftBlindBoxListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.LiveGiftBlindBoxListPageItem{
			ID:                v.ID,
			UID:               v.UID,
			Uname:             v.Uname,
			GiftName:          v.GiftName,
			Price:             v.Price,
			Num:               v.Num,
			OriginalGiftName:  v.OriginalGiftName,
			OriginalGiftPrice: v.OriginalGiftPrice,
			BadgeName:         v.BadgeName,
			BadgeLevel:        v.BadgeLevel,
			BadgeType:         v.BadgeType,
			SendAt:            v.SendAt,
		})
	}
	return res
}
