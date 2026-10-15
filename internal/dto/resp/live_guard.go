package resp

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// LiveGuardListPageResp 大航海名单查询返回
type LiveGuardListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []LiveGuardListPageItem `json:"pageData"`
}

// LiveGuardListPageItem 名单一行，粒度为「用户 × 月份」
type LiveGuardListPageItem struct {
	// 用户uid
	UID int64 `json:"uid"`
	// 用户昵称
	Uname string `json:"uname"`
	// 大航海等级, 1-总督, 2-提督, 3-舰长
	GuardLevel enum.BadgeType `json:"guard_level"`
	// 连续覆盖段起点所在月, 格式 2006-01
	CoverFromMonth string `json:"cover_from_month" example:"2026-07"`
	// 连续覆盖段终点所在月, 格式 2006-01, 含（中间断档的月份不算在同一段里）
	CoverToMonth string `json:"cover_to_month" example:"2026-09"`
	// 当月开通份数
	OpenNum int64 `json:"open_num" example:"3"`
	// 当月最早一次开通时间, 空表示当月没开通（是往月延续来的）
	OpenAt string `json:"open_at"`
}
