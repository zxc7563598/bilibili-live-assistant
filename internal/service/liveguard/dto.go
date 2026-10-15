package liveguard

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"
)

// ListPageReq 名单查询入参
//
// Month 为 "2006-01"，GuardLevel 必传：它决定了名单里 uid 唯一，
// 也是导出按 uid 分块的前提。
type ListPageReq struct {
	pagination.PageResp
	GuardLevel *int    `json:"guard_level"`
	Month      string  `json:"month"`
	UID        *int64  `json:"uid"`
	Uname      *string `json:"uname"`
}

// ListPageResp 名单查询返回
type ListPageResp struct {
	Total    int64 `json:"total"`
	PageData []ListPageItem
}

// ListPageItem 名单一行，粒度为「用户 × 月份」
//
// 月份字段都是展示用的 "2006-01"，两端是「覆盖查询月的连续覆盖段」的端点：
// 分次续费连排出来的月份算同一段，中间断档的月份断开。CoverToMonth 是覆盖到的
// 最后一个月（含），台账里存的是不含的月首，两者差一个月。
type ListPageItem struct {
	UID            int64          `json:"uid"`
	Uname          string         `json:"uname"`
	GuardLevel     enum.BadgeType `json:"guard_level"`
	CoverFromMonth string         `json:"cover_from_month"` // 连续覆盖段起点所在月
	CoverToMonth   string         `json:"cover_to_month"`   // 连续覆盖段终点所在月, 含
	OpenNum        int64          `json:"open_num"`         // 当月开通份数
	OpenAt         string         `json:"open_at"`          // 当月最早一次开通时间, 空表示当月没开通
}

// BackfillReq 补录一次大航海开通
//
// 礼物ID与价格按档位取，不由调用方传（见 enum.GuardGiftByBadgeType）。
type BackfillReq struct {
	UID        int64  `json:"uid"`
	Uname      string `json:"uname"`
	RoomID     int64  `json:"room_id"`
	GuardLevel int    `json:"guard_level"`
	OpenAt     int64  `json:"open_at"`
	Num        int64  `json:"num"`
}
