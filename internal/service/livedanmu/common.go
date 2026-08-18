package livedanmu

import (
	"strconv"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

func toListPageItems(liveDanmu []model.LiveDanmu) []ListPageItem {
	respList := make([]ListPageItem, 0, len(liveDanmu))
	for _, v := range liveDanmu {
		item := ListPageItem{
			ID:          v.ID,
			UID:         v.UID,
			Uname:       v.Uname,
			Msg:         v.Msg,
			BadgeRoomID: v.BadgeRoomID,
			BadgeName:   v.BadgeName,
			BadgeLevel:  v.BadgeLevel,
			BadgeType:   v.BadgeType,
			SendAt:      timeutil.Format(v.SendAt),
		}
		respList = append(respList, item)
	}
	return respList
}

func toFetchRoomGroupItems(ids []int64) []FetchRoomGroupsResp {
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
