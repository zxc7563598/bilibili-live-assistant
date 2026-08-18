package livedanmu

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livedanmu"
)

func toLiveDanmuListItems(list []livedanmu.ListPageItem) []resp.LiveDanmuListPageItem {
	res := make([]resp.LiveDanmuListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.LiveDanmuListPageItem{
			ID:          v.ID,
			UID:         v.UID,
			Uname:       v.Uname,
			Msg:         v.Msg,
			BadgeRoomID: v.BadgeRoomID,
			BadgeName:   v.BadgeName,
			BadgeLevel:  v.BadgeLevel,
			BadgeType:   v.BadgeType,
			SendAt:      v.SendAt,
		})
	}
	return res
}

func toFetchRoomGroupsItems(list []livedanmu.FetchRoomGroupsResp) []resp.LiveDanmuFetchRoomGroupsItem {
	res := make([]resp.LiveDanmuFetchRoomGroupsItem, 0, len(list))
	for _, item := range list {
		res = append(res, resp.LiveDanmuFetchRoomGroupsItem{
			Label: item.Label,
			Value: item.Value,
		})
	}
	return res
}
