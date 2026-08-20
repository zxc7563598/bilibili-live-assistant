package livegift

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livegift"
)

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

func toFetchRoomGroupsItems(list []livegift.FetchRoomGroupsResp) []resp.LiveGiftFetchRoomGroupsItem {
	res := make([]resp.LiveGiftFetchRoomGroupsItem, 0, len(list))
	for _, item := range list {
		res = append(res, resp.LiveGiftFetchRoomGroupsItem{
			Label: item.Label,
			Value: item.Value,
		})
	}
	return res
}
