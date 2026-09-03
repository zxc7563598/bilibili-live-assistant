package input

// OrderPlaceOrderReq 用户下单请求
type OrderPlaceOrderReq struct {
	// SKU ID
	SkuID int64 `json:"sku_id" binding:"required" err:"required=11101" example:"0"`
	// 购买数量
	Count int64 `json:"count" binding:"required,min=1" err:"required=11101,min=11101" example:"0"`
}
