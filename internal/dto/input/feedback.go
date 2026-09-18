package input

// FeedbackSubmitReq 用户提交投诉/反馈请求入参
type FeedbackSubmitReq struct {
	// 问题类型
	Type string `json:"type" binding:"required,max=100" err:"required=11201,max=11202" example:"直播问题"`
	// 投诉内容
	Content string `json:"content" binding:"required,max=5000" err:"required=11203,max=11204" example:"主播发布违规内容，请核查"`
	// 联系方式（手机号 / 邮箱 / QQ 等，便于平台回访）
	Contact string `json:"contact" binding:"required,max=100" err:"required=11205,max=11206" example:"18888888888"`
}

// FeedbackListPageReq 后台分页查询投诉请求
type FeedbackListPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=11207" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=11207" example:"20"`
	// 排序字段，取投诉表列名或 uid/uname
	SortField *string `json:"sortField" example:"created_at"`
	// 排序方向 ascend/descend
	SortOrder *string `json:"sortOrder" binding:"omitempty,oneof=ascend descend" err:"oneof=11207" example:"descend" enums:"ascend,descend"`
	// 用户UID，精确匹配
	UID *int64 `json:"uid" example:"54272611"`
	// 用户昵称，模糊搜索
	Uname *string `json:"uname" example:"哎呀又胖啦"`
}

// FeedbackDetailsReq 后台获取投诉详情请求
type FeedbackDetailsReq struct {
	// 投诉ID
	ID int64 `json:"id" binding:"required,min=1" err:"required=11207,min=11207" example:"1"`
}
