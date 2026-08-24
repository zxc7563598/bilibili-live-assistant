package liveuser

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

// UserBalance 用户余额响应结构
type UserBalance struct {
	Points int64 // 积分
	Stars  int64 // 星光
}

// AddCreditLogParams 添加积分/星光
type AddCreditLogParams struct {
	UserID       int64             // 用户ID
	ChangeType   enum.ChangeType   // 变动类型（增加/减少）
	ChangeAmount int64             // 变动数值
	BizType      string            // 业务类型
	Remark       string            // 备注
	OperatorType enum.OperatorType // 操作方
	OperatorID   int64             // 操作人标识ID
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
	PageResp
	UID   *int64  `json:"uid"`
	Uname *string `json:"uname"`
}

// ListPage 请求返回
type ListPageResp struct {
	Total    int64 `json:"total"`
	PageData []ListPageItem
}

type ListPageItem struct {
	ID              int64  `json:"id"`
	UID             int64  `json:"uid"`
	Uname           string `json:"uname"`
	Points          int64  `json:"points"`
	Stars           int64  `json:"stars"`
	TotalDanmuCount int64  `json:"total_danmu_count"`
	TotalGiftAmount int64  `json:"total_gift_amount"`
}
