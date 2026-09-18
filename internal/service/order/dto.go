package order

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"
)

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
	pagination.PageResp
	OrderStatus *int `json:"order_status"`
}

// ListPageByUserResp 请求返回
type ListPageByUserResp struct {
	Total    int64 `json:"total"`
	PageData []ListPageItem
}

type ListPageItem struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
	// UID/Uname 联查 live_users 得到，仅后台列表使用；商城端 converter 不拷贝
	UID                   int64            `json:"uid"`
	Uname                 string           `json:"uname"`
	OrderSn               string           `json:"order_sn"`
	ProductID             int64            `json:"product_id"`
	ProductName           string           `json:"product_name"`
	ProductCover          string           `json:"product_cover"`
	ProductSpecProperties string           `json:"product_spec_properties"`
	Quantity              int64            `json:"quantity"`
	CreditType            enum.CreditType  `json:"credit_type"`
	Price                 int64            `json:"price"`
	ReceiverType          enum.AddressType `json:"receiver_type"`
	ReceiverName          string           `json:"receiver_name"`
	ReceiverPhone         string           `json:"receiver_phone"`
	ReceiverRegion        string           `json:"receiver_region"`
	ReceiverDetail        string           `json:"receiver_detail"`
	ReceiverEmail         string           `json:"receiver_email"`
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

// ListPageReq 后台分页查询订单请求入参
type ListPageReq struct {
	pagination.PageResp
	UID         *int64
	Uname       *string
	OrderSn     *string
	OrderStatus *int
	PayStatus   *int
	ShipStatus  *int
}

// ListPageResp 后台分页查询订单请求返回
type ListPageResp struct {
	Total    int64          `json:"total"`
	PageData []ListPageItem `json:"page_data"`
}

// DetailsItem 订单详情：订单全部字段 + 用户 uid/uname
type DetailsItem struct {
	ID                    int64            `json:"id"`
	UserID                int64            `json:"user_id"`
	UID                   int64            `json:"uid"`
	Uname                 string           `json:"uname"`
	OrderSn               string           `json:"order_sn"`
	ProductID             int64            `json:"product_id"`
	ProductSkuID          int64            `json:"product_sku_id"`
	ProductName           string           `json:"product_name"`
	ProductCover          string           `json:"product_cover"`
	ProductSpecProperties string           `json:"product_spec_properties"`
	Quantity              int64            `json:"quantity"`
	CreditType            enum.CreditType  `json:"credit_type"`
	Price                 int64            `json:"price"`
	ReceiverName          string           `json:"receiver_name"`
	ReceiverPhone         string           `json:"receiver_phone"`
	ReceiverRegionCode    string           `json:"receiver_region_code"`
	ReceiverRegion        string           `json:"receiver_region"`
	ReceiverDetail        string           `json:"receiver_detail"`
	ReceiverEmail         string           `json:"receiver_email"`
	ReceiverType          enum.AddressType `json:"receiver_type"`
	OrderStatus           enum.OrderStatus `json:"order_status"`
	PayStatus             enum.PayStatus   `json:"pay_status"`
	ShipStatus            enum.ShipStatus  `json:"ship_status"`
	ExpressCompany        string           `json:"express_company"`
	ExpressNo             string           `json:"express_no"`
	PayAt                 string           `json:"pay_at"`
	ProcessedAt           string           `json:"processed_at"`
	CancelAt              string           `json:"cancel_at"`
	CreatedAt             string           `json:"created_at"`
	UpdatedAt             string           `json:"updated_at"`
	Remark                string           `json:"remark"`
}

// UpdateShipStatusReq 变更发货状态请求入参
type UpdateShipStatusReq struct {
	ID             int64
	ShipStatus     enum.ShipStatus
	ExpressCompany *string
	ExpressNo      *string
}

// UpdateOrderStatusReq 变更订单状态请求入参
type UpdateOrderStatusReq struct {
	ID          int64
	OrderStatus enum.OrderStatus
}

// UpdateReceiverInfoReq 变更订单收货信息请求入参。
// 故意不含 ReceiverRegion：地区文案始终由后端从 region_code 派生，不接受前端自由文本。
type UpdateReceiverInfoReq struct {
	ID                 int64
	ReceiverName       *string
	ReceiverPhone      *string
	ReceiverRegionCode *string
	ReceiverDetail     *string
	ReceiverEmail      *string
}
