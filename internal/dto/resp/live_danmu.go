package resp

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// LiveDanmuListPageResp 分页查询弹幕列表返回
type LiveDanmuListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []LiveDanmuListPageItem `json:"pageData"`
}

type LiveDanmuListPageItem struct {
	// 弹幕ID
	ID int64 `json:"id" example:"1"`
	// 用户UID
	UID int64 `json:"uid" example:"27461511"`
	// 用户名称
	Uname string `json:"uname" example:"哎呀又胖啦"`
	// 弹幕内容
	Msg string `json:"msg" example:"好耶"`
	// 勋章房间ID
	BadgeRoomID int64 `json:"badge_room_id" example:"22384516"`
	// 勋章名称
	BadgeName string `json:"badge_name" example:"小米星"`
	// 勋章等级
	BadgeLevel int64 `json:"badge_level" example:"30"`
	// 勋章类型
	BadgeType enum.BadgeType `json:"badge_type" example:"1"`
	// 创建时间
	SendAt string `json:"send_at" example:"2025-01-02 12:22:22"`
}

// LiveDanmuFetchRoomGroupsResp 获取全部房间ID返回
type LiveDanmuFetchRoomGroupsResp struct {
	Option []LiveDanmuFetchRoomGroupsItem `json:"option"`
}

type LiveDanmuFetchRoomGroupsItem struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}
