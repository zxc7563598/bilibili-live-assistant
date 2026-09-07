package order

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// 通用分页请求参数
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

// PlaceOrderReq 请求入参
type PlaceOrderReq struct {
	SkuID int64 `json:"sku_id"`
	Count int64 `json:"count"`
}

// UserOrderDraftResp 请求返回
type UserOrderDraftResp struct {
	ID       int64       `json:"id"`
	ExpireAt int64       `json:"expire_at"`
	Product  ProductItem `json:"product"`
}

type ProductItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Cover       string `json:"cover"`
	Price       int64  `json:"price"`
	CreditType  int    `json:"credit_type"`
	ProductType int    `json:"product_type"`
	Sku         string `json:"sku"`
	Count       int64  `json:"count"`
}

// ListPageByUserReq 请求入参
type ListPageByUserReq struct {
	PageResp
	OrderStatus *int `json:"order_status"`
}

// ListPageByUserResp 请求返回
type ListPageByUserResp struct {
	Total    int64 `json:"total"`
	PageData []ListPageItem
}

type ListPageItem struct {
	ID                    int64            `json:"id"`
	OrderSn               string           `json:"order_sn"`
	ProductID             int64            `json:"product_id"`
	ProductName           string           `json:"product_name"`
	ProductCover          string           `json:"product_cover"`
	ProductSpecProperties string           `json:"product_spec_properties"`
	Quantity              int64            `json:"quantity"`
	CreditType            enum.CreditType  `json:"credit_type"`
	Price                 int64            `json:"price"`
	OrderStatus           enum.OrderStatus `json:"order_status"`
	PayStatus             enum.PayStatus   `json:"pay_status"`
	ShipStatus            enum.ShipStatus  `json:"ship_status"`
	ExpressCompany        string           `json:"express_company"`
	ExpressNo             string           `json:"express_no"`
	PayAt                 string           `json:"pay_at"`
	ProcessedAt           string           `json:"processed_at"`
	CancelAt              string           `json:"cancel_at"`
	CreatedAt             string           `json:"created_at"`
	Remark                string           `json:"remark"`
}
