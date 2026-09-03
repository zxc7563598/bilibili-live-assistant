package order

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

// PlaceOrderReq 请求入参
type PlaceOrderReq struct {
	SkuID int64 `json:"sku_id"`
	Count int64 `json:"count"`
}
