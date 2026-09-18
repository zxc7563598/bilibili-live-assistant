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

type OrderListPageByUserReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=11101" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=11101" example:"20"`
	// 排序字段
	SortField *string `json:"sortField" example:"points"`
	// 排序方向 ascend/descend
	SortOrder *string `json:"sortOrder" example:"descend" enums:"ascend,descend"`
	// 订单状态
	OrderStatus *int `json:"order_status" example:"0" enums:"0,1,2,3,4,5"`
}

// OrderListPageReq 后台分页查询订单请求。
// 状态类筛选字段使用指针：裸 int 上 required 会把合法的 0（未发货等）判为缺失，且无法区分未传与传 0。
type OrderListPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=11101" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=11101" example:"20"`
	// 排序字段，取订单表列名或 uid/uname
	SortField *string `json:"sortField" example:"created_at"`
	// 排序方向 ascend/descend
	SortOrder *string `json:"sortOrder" example:"descend" enums:"ascend,descend"`
	// 用户UID，精确匹配
	UID *int64 `json:"uid" example:"54272611"`
	// 用户昵称，模糊搜索
	Uname *string `json:"uname" example:"哎呀又胖啦"`
	// 订单号，模糊匹配
	OrderSn *string `json:"order_sn" example:"SOdleu2legnzxs13hal3"`
	// 订单状态
	OrderStatus *int `json:"order_status" example:"1" enums:"0,1,2,3,4,5"`
	// 支付状态
	PayStatus *int `json:"pay_status" example:"1" enums:"0,1,2"`
	// 发货状态
	ShipStatus *int `json:"ship_status" example:"0" enums:"0,1,2"`
}

// OrderDetailsReq 后台获取订单详情请求
type OrderDetailsReq struct {
	// 订单ID
	ID int64 `json:"id" binding:"required,min=1" err:"required=11101,min=11101" example:"1"`
}

// OrderUpdateShipStatusReq 后台变更发货状态请求。
// 快递信息可选，仅实体订单可填；虚拟订单传了快递字段会返回参数错误。
type OrderUpdateShipStatusReq struct {
	// 订单ID
	ID int64 `json:"id" binding:"required,min=1" err:"required=11101,min=11101" example:"1"`
	// 发货状态，0 未发货 / 1 已发货 / 2 已送达
	ShipStatus *int `json:"ship_status" binding:"required,oneof=0 1 2" err:"required=11101,oneof=11101" example:"1" enums:"0,1,2"`
	// 快递公司，仅实体订单可填
	ExpressCompany *string `json:"express_company" binding:"omitempty,max=100" err:"max=11101" example:"顺丰速运"`
	// 快递单号，仅实体订单可填
	ExpressNo *string `json:"express_no" binding:"omitempty,max=100" err:"max=11101" example:"SF1234567890"`
}

// OrderUpdateOrderStatusReq 后台变更订单状态请求
type OrderUpdateOrderStatusReq struct {
	// 订单ID
	ID int64 `json:"id" binding:"required,min=1" err:"required=11101,min=11101" example:"1"`
	// 订单状态，0 待付款 / 1 待发货 / 2 待收货 / 3 已完成 / 4 已取消 / 5 售后中
	OrderStatus *int `json:"order_status" binding:"required,oneof=0 1 2 3 4 5" err:"required=11101,oneof=11101" example:"3" enums:"0,1,2,3,4,5"`
}

// OrderUpdateReceiverInfoReq 后台变更订单收货信息请求。
// 按订单的收货类型分支：虚拟订单只接受 receiver_email，
// 实体订单只接受 receiver_name / receiver_phone / receiver_region_code / receiver_detail；
// 传了不适用于该类型的字段会返回参数错误。地区文案由后端按 region_code 派生，不接受前端传入。
type OrderUpdateReceiverInfoReq struct {
	// 订单ID
	ID int64 `json:"id" binding:"required,min=1" err:"required=11101,min=11101" example:"1"`
	// 收货人姓名，仅实体订单可填
	ReceiverName *string `json:"receiver_name" binding:"omitempty,max=100" err:"max=11101" example:"张三"`
	// 收货人手机号，仅实体订单可填
	ReceiverPhone *string `json:"receiver_phone" binding:"omitempty,max=100" err:"max=11101" example:"18888888888"`
	// 收货人地区code，JSON 数组字符串，仅实体订单可填
	ReceiverRegionCode *string `json:"receiver_region_code" binding:"omitempty,max=100" err:"max=11101" example:"['370000', '370100', '370116']"`
	// 收货人详细地址，仅实体订单可填
	ReceiverDetail *string `json:"receiver_detail" binding:"omitempty,max=255" err:"max=11101" example:"xx路xx号"`
	// 收货人邮箱地址，仅虚拟订单可填
	ReceiverEmail *string `json:"receiver_email" binding:"omitempty,max=255" err:"max=11101" example:"x@x.com"`
}
