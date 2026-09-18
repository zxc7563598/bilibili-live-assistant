package livegift

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

// BlindBoxListPage 请求入参
type BlindBoxListPageReq struct {
	pagination.PageResp
	RoomID           *int64  `json:"room_id"`
	UID              *int64  `json:"uid"`
	Uname            *string `json:"uname"`
	GiftName         *string `json:"gift_name"`
	OriginalGiftName *string `json:"original_gift_name"`
	SendAtStart      *int64  `json:"send_at_start"`
	SendAtEnd        *int64  `json:"send_at_end"`
}

// BlindBoxListPage 请求返回
type BlindBoxListPageResp struct {
	Total    int64 `json:"total"`
	PageData []BlindBoxListPageItem
	Stats    BlindBoxListPageStats
}

type BlindBoxListPageItem struct {
	ID                int64          `json:"id"`
	UID               int64          `json:"uid"`
	Uname             string         `json:"uname"`
	GiftName          string         `json:"gift_name"`
	Price             int64          `json:"price"`
	Num               int64          `json:"num"`
	OriginalGiftName  string         `json:"original_gift_name"`
	OriginalGiftPrice int64          `json:"original_gift_price"`
	BadgeName         string         `json:"badge_name"`
	BadgeLevel        int64          `json:"badge_level"`
	BadgeType         enum.BadgeType `json:"badge_type"`
	SendAt            string         `json:"send_at"`
}

type BlindBoxListPageStats struct {
	OriginalPrice int64 `json:"original_price"`
	CurrentPrice  int64 `json:"current_price"`
}
