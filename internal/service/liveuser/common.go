package liveuser

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
)

func toListPageItems(liveUser []model.LiveUser) []ListPageItem {
	respList := make([]ListPageItem, 0, len(liveUser))
	for _, v := range liveUser {
		item := ListPageItem{
			ID:              v.ID,
			UID:             v.UID,
			Uname:           v.Uname,
			Points:          v.Points,
			Stars:           v.Stars,
			TotalDanmuCount: v.TotalDanmuCount,
			TotalGiftAmount: v.TotalGiftAmount,
		}
		respList = append(respList, item)
	}
	return respList
}
