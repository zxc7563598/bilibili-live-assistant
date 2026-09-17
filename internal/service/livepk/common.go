package livepk

import (
	"strconv"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// toListPageItems 把 PK 记录转换成列表出参
func toListPageItems(logs []model.LivePkLog) []ListPageItem {
	respList := make([]ListPageItem, 0, len(logs))
	for _, v := range logs {
		respList = append(respList, ListPageItem{
			ID:          v.ID,
			RoomID:      v.RoomID,
			PkID:        v.PkID,
			PkStatus:    v.PkStatus,
			BattleType:  v.BattleType,
			MatchType:   v.MatchType,
			RivalUID:    v.RivalUID,
			RivalUname:  v.RivalUname,
			RivalRoomID: v.RivalRoomID,
			SelfVotes:   v.SelfVotes,
			RivalVotes:  v.RivalVotes,
			SelfResult:  v.SelfResult,
			RivalResult: v.RivalResult,
			StartAt:     timeutil.Format(v.StartAt),
			SettleAt:    timeutil.Format(v.SettleAt),
		})
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
