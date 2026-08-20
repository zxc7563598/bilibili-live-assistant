package livegift

import (
	"strconv"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

func toListPageItems(liveGifts []model.LiveGift) []ListPageItem {
	respList := make([]ListPageItem, 0, len(liveGifts))
	for _, v := range liveGifts {
		item := ListPageItem{
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
			SendAt:     timeutil.Format(v.SendAt),
		}
		respList = append(respList, item)
	}
	return respList
}

func toBlindBoxListPageItems(liveGifts []model.LiveGift) []BlindBoxListPageItem {
	respList := make([]BlindBoxListPageItem, 0, len(liveGifts))
	for _, v := range liveGifts {
		item := BlindBoxListPageItem{
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
			SendAt:            timeutil.Format(v.SendAt),
		}
		respList = append(respList, item)
	}
	return respList
}

func toFetchRoomGroupsItems(ids []int64) []FetchRoomGroupsResp {
	options := make([]FetchRoomGroupsResp, 0, len(ids))
	for _, id := range ids {
		str := strconv.FormatInt(id, 10)
		options = append(options, FetchRoomGroupsResp{
			Label: str,
			Value: id,
		})
	}
	return options
}
