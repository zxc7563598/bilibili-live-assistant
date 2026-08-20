package livegift

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// 通用分页请求参数
type PageResp struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

func (r *PageResp) OffsetLimit() (int, int) {
	if r.PageNo < 1 {
		r.PageNo = 1
	}
	if r.PageSize < 1 {
		r.PageSize = 10
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}
	offset := (r.PageNo - 1) * r.PageSize
	return offset, r.PageSize
}

// ListPage 请求入参
type ListPageReq struct {
	PageResp
	RoomID      *int64  `json:"room_id"`
	UID         *int64  `json:"uid"`
	Uname       *string `json:"uname"`
	GiftName    *string `json:"gift_name"`
	GiftType    *int    `json:"gift_type"`
	Original    *int    `json:"original"`
	SendAtStart *int64  `json:"send_at_start"`
	SendAtEnd   *int64  `json:"send_at_end"`
}

// ListPage 请求返回
type ListPageResp struct {
	Total    int64 `json:"total"`
	PageData []ListPageItem
	Stats    ListPageStats
}

type ListPageItem struct {
	ID         int64          `json:"id"`
	UID        int64          `json:"uid"`
	Uname      string         `json:"uname"`
	GiftName   string         `json:"gift_name"`
	Price      int64          `json:"price"`
	Num        int64          `json:"num"`
	Message    string         `json:"message"`
	BadgeName  string         `json:"badge_name"`
	BadgeLevel int64          `json:"badge_level"`
	BadgeType  enum.BadgeType `json:"badge_type"`
	SendAt     string         `json:"send_at"`
}

type ListPageStats struct {
	TotalNum    int64 `json:"total_num"`
	TotalAmount int64 `json:"total_amount"`
}

// FetchRoomGroups 请求返回
type FetchRoomGroupsResp struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}
