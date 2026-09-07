package input

// OrderPlaceOrderReq 用户下单请求
type OrderPlaceOrderReq struct {
	// SKU ID
	SkuID int64 `json:"sku_id" binding:"required,min=1" err:"required=11101,min=11101" example:"0"`
	// 购买数量
	Count int64 `json:"count" binding:"required,min=1" err:"required=11101,min=11101" example:"0"`
}

// OrderReOrderReq 用户重新下单请求
type OrderReOrderReq struct {
	// 历史草稿ID
	ID int64 `json:"id" binding:"required,min=1" err:"required=11101,min=11101" example:"0"`
}

// OrderConfirmPaymentReq 用户确认支付请求
type OrderConfirmPaymentReq struct {
	// 历史草稿ID
	DraftID int64 `json:"draft_id" binding:"required,min=1" err:"required=11101,min=11101" example:"0"`
	// 收货地址ID
	AddressID int64 `json:"address_id" binding:"required,min=1" err:"required=11101,min=11101" example:"0"`
}
