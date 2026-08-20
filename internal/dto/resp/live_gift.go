package resp

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// LiveGiftListPageResp 分页查询礼物列表返回
type LiveGiftListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []LiveGiftListPageItem `json:"pageData"`
	// 统计信息
	Stats LiveGiftListPageStats `json:"stats"`
}

type LiveGiftListPageItem struct {
	// 礼物ID
	ID int64 `json:"id" example:"1"`
	// 用户UID
	UID int64 `json:"uid" example:"27461511"`
	// 用户名称
	Uname string `json:"uname" example:"哎呀又胖啦"`
	// 礼物名称
	GiftName string `json:"gift_name" example:"盲盒"`
	// 礼物单价（分）
	Price int64 `json:"price" example:"300"`
	// 数量
	Num int64 `json:"num" example:"1"`
	// 醒目留言内容
	Message string `json:"message" example:"xxxxxxxxx"`
	// 勋章名称
	BadgeName string `json:"badge_name" example:"小米星"`
	// 勋章等级
	BadgeLevel int64 `json:"badge_level" example:"30"`
	// 勋章类型
	BadgeType enum.BadgeType `json:"badge_type" example:"1"`
	// 发送时间
	SendAt string `json:"send_at" example:"2025-01-02 12:22:22"`
}

type LiveGiftListPageStats struct {
	// 礼物总数
	TotalNum int64 `json:"total_num"`
	// 礼物总金额（分）
	TotalAmount int64 `json:"total_amount"`
}

// LiveGiftFetchRoomGroupsResp 获取全部房间ID返回
type LiveGiftFetchRoomGroupsResp struct {
	Option []LiveGiftFetchRoomGroupsItem `json:"option"`
}

type LiveGiftFetchRoomGroupsItem struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// LiveGiftBlindBoxListPageResp 分页查询盲盒礼物列表返回
type LiveGiftBlindBoxListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []LiveGiftBlindBoxListPageItem `json:"pageData"`
	// 统计信息
	Stats LiveGiftBlindBoxListPageStats `json:"stats"`
}

type LiveGiftBlindBoxListPageItem struct {
	// 礼物ID
	ID int64 `json:"id" example:"1"`
	// 用户UID
	UID int64 `json:"uid" example:"27461511"`
	// 用户名称
	Uname string `json:"uname" example:"哎呀又胖啦"`
	// 礼物名称
	GiftName string `json:"gift_name" example:"盲盒"`
	// 礼物单价（分）
	Price int64 `json:"price" example:"300"`
	// 数量
	Num int64 `json:"num" example:"1"`
	// 原礼物名称
	OriginalGiftName string `json:"original_gift_name" example:"盲盒"`
	// 原礼物单价（分）
	OriginalGiftPrice int64 `json:"original_gift_price" example:"300"`
	// 勋章名称
	BadgeName string `json:"badge_name" example:"小米星"`
	// 勋章等级
	BadgeLevel int64 `json:"badge_level" example:"30"`
	// 勋章类型
	BadgeType enum.BadgeType `json:"badge_type" example:"1"`
	// 发送时间
	SendAt string `json:"send_at" example:"2025-01-02 12:22:22"`
}

type LiveGiftBlindBoxListPageStats struct {
	// 原礼物总金额（分）
	OriginalPrice int64 `json:"original_price"`
	// 礼物总金额（分）
	CurrentPrice int64 `json:"current_price"`
}
