package resp

// FeedbackListPageResp 后台分页查询投诉返回
type FeedbackListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []FeedbackListPageItem `json:"pageData"`
}

// FeedbackListPageItem 投诉列表项；不含投诉正文（列表不展示，避免下发大字段）
type FeedbackListPageItem struct {
	// 投诉ID
	ID int64 `json:"id" example:"1"`
	// 用户内部ID
	UserID int64 `json:"user_id" example:"1"`
	// 用户UID
	UID int64 `json:"uid" example:"54272611"`
	// 用户昵称
	Uname string `json:"uname" example:"哎呀又胖啦"`
	// 问题类型
	Type string `json:"type" example:"直播问题"`
	// 联系方式
	Contact string `json:"contact" example:"18888888888"`
	// 投诉时间
	CreatedAt string `json:"created_at" example:"xxxx-xx-xx xx:xx:xx"`
}

// FeedbackDetailsResp 后台获取投诉详情返回，包含投诉全部字段与用户 uid/uname
type FeedbackDetailsResp struct {
	// 投诉ID
	ID int64 `json:"id" example:"1"`
	// 用户内部ID
	UserID int64 `json:"user_id" example:"1"`
	// 用户UID
	UID int64 `json:"uid" example:"54272611"`
	// 用户昵称
	Uname string `json:"uname" example:"哎呀又胖啦"`
	// 问题类型
	Type string `json:"type" example:"直播问题"`
	// 投诉内容
	Content string `json:"content" example:"主播发布违规内容，请核查"`
	// 联系方式
	Contact string `json:"contact" example:"18888888888"`
	// 投诉时间
	CreatedAt string `json:"created_at" example:"xxxx-xx-xx xx:xx:xx"`
}
