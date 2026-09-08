package feedback

// CreateReq 新增一条用户投诉/反馈的入参；UserID 由调用方从认证上下文传入，不放进请求体
type CreateReq struct {
	Type    string
	Content string
	Contact string
}
