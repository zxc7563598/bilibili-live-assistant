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
