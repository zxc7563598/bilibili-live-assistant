package feedback

// CreateReq 新增一条用户投诉/反馈的入参；UserID 由调用方从认证上下文传入，不放进请求体
type CreateReq struct {
	Type    string
	Content string
	Contact string
}

// PageResp 通用分页请求参数
type PageResp struct {
	PageNo    int     `json:"pageNo"`
	PageSize  int     `json:"pageSize"`
	SortField *string `json:"sortField"`
	SortOrder *string `json:"sortOrder"`
}

func (r *PageResp) OffsetLimit() (int, int, *string, *string) {
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
	return offset, r.PageSize, r.SortField, r.SortOrder
}

// ListPageReq 后台分页查询投诉请求入参
type ListPageReq struct {
	PageResp
	// UID 联查 live_users.uid，精确匹配
	UID *int64
	// Uname 联查 live_users.uname，模糊匹配
	Uname *string
}

// ListPageItem 投诉列表项；不含投诉正文（列表不展示）
type ListPageItem struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	UID       int64  `json:"uid"`
	Uname     string `json:"uname"`
	Type      string `json:"type"`
	Contact   string `json:"contact"`
	CreatedAt string `json:"created_at"`
}

// ListPageResp 后台分页查询投诉请求返回
type ListPageResp struct {
	Total    int64          `json:"total"`
	PageData []ListPageItem `json:"page_data"`
}

// DetailsItem 投诉详情：投诉全部字段 + 用户 uid/uname
type DetailsItem struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	UID       int64  `json:"uid"`
	Uname     string `json:"uname"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	Contact   string `json:"contact"`
	CreatedAt string `json:"created_at"`
}
