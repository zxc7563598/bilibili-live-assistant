package input

// AddressGetDefaultAddressReq 获取用户指定类型的默认收货地址请求入参
type AddressGetDefaultAddressReq struct {
	Type int `json:"type" example:"1" enums:"0,1"`
}

// AddressGetAddressListReq 获取用户收货地址列表请求入参
type AddressGetAddressListReq struct {
	Type *int `json:"type" example:"1" enums:"0,1"`
}

// AddressGetAddressByIDReq 获取收货地址的详细信息请求入参
type AddressGetAddressByIDReq struct {
	ID int `json:"id"  binding:"required" err:"required=11301" example:"1"`
}

// AddressSaveAddressReq 添加/变更收货地址请求入餐
// 除 id 外的字段均为可选指针：新增时省略按默认处理，修改时省略表示保留原值。
type AddressSaveAddressReq struct {
	ID         *int64  `json:"id"  example:"2"`
	Name       *string `json:"name" example:"哎呀又胖啦"`
	Phone      *string `json:"phone" example:"18888888888"`
	RegionCode *string `json:"region_code" example:"['370000', '370100', '370116']"`
	Region     *string `json:"region" example:"xxx xxx xxx"`
	Detail     *string `json:"detail" example:"xxxxxxxxxx"`
	Email      *string `json:"email" example:"xxxxx@xxx.xx"`
	Type       *int    `json:"type" example:"1" enums:"0,1"`
	IsDefault  *int    `json:"is_default" example:"1" enums:"0,1"`
}

// AddressDeleteAddressReq 删除收货地址请求入参
type AddressDeleteAddressReq struct {
	ID int64 `json:"id" binding:"required" err:"required=11301" example:"1"`
}
