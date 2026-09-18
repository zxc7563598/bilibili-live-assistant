// Package pagination 提供 Service 层通用的分页请求参数。
//
// 原先 admin / role / order / product / liveuser / livegift / livepk / livedanmu / feedback
// 九个 service 包各自复制了一份完全相同的 PageResp 与 OffsetLimit，这里收敛为一份。
package pagination

// PageResp 通用分页请求参数
type PageResp struct {
	PageNo    int     `json:"pageNo"`
	PageSize  int     `json:"pageSize"`
	SortField *string `json:"sortField"`
	SortOrder *string `json:"sortOrder"`
}

// OffsetLimit 归一化分页参数，返回 (offset, limit, sortField, sortOrder)。
// PageNo / PageSize 会被就地钳位：页码最小 1，页大小默认 10、上限 100。
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
