package livepk

import "github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"

// 我方胜负的取值，由 B站 下发的胜负字段收敛而来，与库中 self_result / rival_result 同口径
const (
	resultLose = -1 // 落败
	resultWin  = 2  // 获胜
)

// FetchRoomGroups 请求返回
type FetchRoomGroupsResp struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// ListPage 请求入参
type ListPageReq struct {
	pagination.PageResp
	RoomID     *int64
	RivalUID   *int64
	RivalUname *string
	// Result 我方胜负，与库中 self_result 同口径：-1=落败 2=获胜
	Result       *int
	StartAtStart *int64
	StartAtEnd   *int64
}

// ListPage 请求返回
type ListPageResp struct {
	Total    int64
	PageData []ListPageItem
	Stats    ListPageStats
}

type ListPageItem struct {
	ID          int64  `json:"id"`
	RoomID      int64  `json:"room_id"`
	PkID        int64  `json:"pk_id"`
	PkStatus    int64  `json:"pk_status"`
	BattleType  int64  `json:"battle_type"`
	MatchType   int64  `json:"match_type"`
	RivalUID    int64  `json:"rival_uid"`
	RivalUname  string `json:"rival_uname"`
	RivalRoomID int64  `json:"rival_room_id"`
	SelfVotes   int64  `json:"self_votes"`
	RivalVotes  int64  `json:"rival_votes"`
	SelfResult  int64  `json:"self_result"`
	RivalResult int64  `json:"rival_result"`
	StartAt     string `json:"start_at"`
	SettleAt    string `json:"settle_at"`
}

type ListPageStats struct {
	TotalNum int64 `json:"total_num"`
	WinNum   int64 `json:"win_num"`
	LoseNum  int64 `json:"lose_num"`
}
