package livedanmu

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"
)

// ListPage 请求入参
type ListPageReq struct {
	pagination.PageResp
	RoomID      *int64  `json:"room_id"`
	UID         *int64  `json:"uid"`
	Uname       *string `json:"uname"`
	Msg         *string `json:"msg"`
	SendAtStart *int64  `json:"send_at_start"`
	SendAtEnd   *int64  `json:"send_at_end"`
}

// ListPage 请求返回
type ListPageResp struct {
	Total    int64 `json:"total"`
	PageData []ListPageItem
}

type ListPageItem struct {
	ID          int64          `json:"id"`
	UID         int64          `json:"uid"`
	Uname       string         `json:"uname"`
	Msg         string         `json:"msg"`
	BadgeRoomID int64          `json:"badge_room_id"`
	BadgeName   string         `json:"badge_name"`
	BadgeLevel  int64          `json:"badge_level"`
	BadgeType   enum.BadgeType `json:"badge_type"`
	SendAt      string         `json:"send_at"`
}
