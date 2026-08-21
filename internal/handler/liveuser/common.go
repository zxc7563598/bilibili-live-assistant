package liveuser

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
)

func toLiveUserListItems(list []liveuser.ListPageItem) []resp.LiveUserListPageItem {
	res := make([]resp.LiveUserListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.LiveUserListPageItem{
			ID:              v.ID,
			UID:             v.UID,
			Uname:           v.Uname,
			Points:          v.Points,
			Stars:           v.Stars,
			TotalDanmuCount: v.TotalDanmuCount,
			TotalGiftAmount: v.TotalGiftAmount,
		})
	}
	return res
}

func toLiveUserWordFrequencyItems(item []liveuser.WordFrequency) []resp.LiveUserWordFrequency {
	res := make([]resp.LiveUserWordFrequency, 0, len(item))
	for _, v := range item {
		res = append(res, resp.LiveUserWordFrequency{
			Word:  v.Word,
			Count: v.Count,
		})
	}
	return res
}
