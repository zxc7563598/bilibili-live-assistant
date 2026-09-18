package liveuser

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/robotconfig"
)

// Handler 直播控制 HTTP 接口处理器
type Handler struct {
	liveuserSvc    *liveuser.Service
	robotConfigSvc *robotconfig.Service
}

// New 创建 Handler 实例
func New(liveuserSvc *liveuser.Service, robotConfigSvc *robotconfig.Service) *Handler {
	return &Handler{
		liveuserSvc:    liveuserSvc,
		robotConfigSvc: robotConfigSvc,
	}
}

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

func toLiveUserAssetsPageItems(list []liveuser.UserAssetsPageItem) []resp.LiveUserAssetsPageItem {
	res := make([]resp.LiveUserAssetsPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.LiveUserAssetsPageItem{
			ID:           v.ID,
			CreditType:   v.CreditType,
			ChangeAmount: v.ChangeAmount,
			ChangeType:   v.ChangeType,
			AfterValue:   v.AfterValue,
			Remark:       v.Remark,
			CreatedAt:    v.CreatedAt,
		})
	}
	return res
}
