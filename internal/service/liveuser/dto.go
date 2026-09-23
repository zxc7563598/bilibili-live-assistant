package liveuser

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"
)

type TokenResp struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// UserBalance 用户余额响应结构
type UserBalance struct {
	Points int64 // 积分
	Stars  int64 // 星光
}

// AdjustCreditParams 添加积分/星光
type AdjustCreditParams struct {
	UserID       int64             // 用户ID
	ChangeType   enum.ChangeType   // 变动类型（增加/减少）
	ChangeAmount int64             // 变动数值
	BizType      string            // 业务类型
	Remark       string            // 备注
	OperatorType enum.OperatorType // 操作方
	OperatorID   int64             // 操作人标识ID
}

// GuardExpire 用户三个档位的大航海到期时间（Unix 秒），nil 表示该档位未设置
//
// 读写共用同一个结构：GetGuardExpire 原样返回，UpdateGuardExpire 整体覆盖，
// 三个字段的存取口径完全对称，没必要拆成两份逐字相同的类型。
// 用结构体而不是三个相邻的 *int64 参数，避免调用处按位置传参传错档位。
type GuardExpire struct {
	Captain  *int64
	Admiral  *int64
	Governor *int64
}

// GetUserMonthlyAnalysis 请求返回
type GetUserMonthlyAnalysisResp struct {
	DanmuCount map[int64]int64 // 每日弹幕数量
	GiftCount  map[int64]int64 // 每日礼物数量
	GiftAmount map[int64]int64 // 每日礼物金额
	LiveDays   map[int64]bool  // 每日是否有开播
}

// GetUserDanmuAnalysis 请求返回
type GetUserDanmuAnalysisResp struct {
	Words    []WordFrequency // 单词
	Bigrams  []WordFrequency // 双词
	Trigrams []WordFrequency // 三词
	Messages []WordFrequency // 短句
}

type WordFrequency struct {
	Word  string
	Count int64
}

// ListPage 请求入参
type ListPageReq struct {
	pagination.PageResp
	UID   *int64  `json:"uid"`
	Uname *string `json:"uname"`
}

// ListPage 请求返回
type ListPageResp struct {
	Total    int64 `json:"total"`
	PageData []ListPageItem
}

type ListPageItem struct {
	ID              int64          `json:"id"`
	UID             int64          `json:"uid"`
	Uname           string         `json:"uname"`
	Points          int64          `json:"points"`
	Stars           int64          `json:"stars"`
	TotalDanmuCount int64          `json:"total_danmu_count"`
	TotalGiftAmount int64          `json:"total_gift_amount"`
	VipType         enum.BadgeType `json:"vip_type"`
}

// UserInfo 请求返回
type UserInfoResp struct {
	UID     int64          `json:"uid"`
	Avatar  string         `json:"avatar"`
	Name    string         `json:"name"`
	Points  int64          `json:"points"`
	Stars   int64          `json:"stars"`
	VipType enum.BadgeType `json:"vip_type"`
}

// UserAssetsPage 请求入参
type UserAssetsPageReq struct {
	pagination.PageResp
	UID        *int64  `json:"uid"`
	Uname      *string `json:"uname"`
	CreditType *int    `json:"credit_type"`
	ChangeType *int    `json:"change_type"`
}

// UserAssetsPage 请求返回
type UserAssetsPageResp struct {
	Total    int64 `json:"total"`
	PageData []UserAssetsPageItem
}

type UserAssetsPageItem struct {
	ID           int64             `json:"id"`
	UserID       int64             `json:"user_id"`
	UID          int64             `json:"uid"`
	Uname        string            `json:"uname"`
	Face         string            `json:"face"`
	CreditType   enum.CreditType   `json:"credit_type"`
	ChangeType   enum.ChangeType   `json:"change_type"`
	ChangeAmount int64             `json:"change_amount"`
	BeforeValue  int64             `json:"before_value"`
	AfterValue   int64             `json:"after_value"`
	BizType      string            `json:"biz_type"`
	Remark       string            `json:"remark"`
	OperatorType enum.OperatorType `json:"operator_type"`
	OperatorID   int64             `json:"operator_id"`
	CreatedAt    string            `json:"created_at"`
}
