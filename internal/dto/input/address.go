package input

// AddressGetDefaultAddressReq 获取用户指定类型的默认收货地址请求入参
//
// 不传（或传 0）按虚拟地址处理。
type AddressGetDefaultAddressReq struct {
	// 地址类型 0 虚拟 / 1 实体
	Type int `json:"type" binding:"omitempty,oneof=0 1" err:"oneof=11301" example:"1" enums:"0,1"`
}

// AddressGetAddressListReq 获取用户收货地址列表请求入参
//
// 请求体可为空；type 省略表示不限类型。
type AddressGetAddressListReq struct {
	// 地址类型 0 虚拟 / 1 实体
	Type *int `json:"type" binding:"omitempty,oneof=0 1" err:"oneof=11301" example:"1" enums:"0,1"`
}

// AddressGetAddressByIDReq 获取收货地址的详细信息请求入参
type AddressGetAddressByIDReq struct {
	ID int `json:"id"  binding:"required" err:"required=11301" example:"1"`
}

// AddressSaveAddressReq 添加/变更收货地址请求入参
// 除 id 外的字段均为可选指针：新增时省略按默认处理，修改时省略表示保留原值。
type AddressSaveAddressReq struct {
	ID         *int64  `json:"id"  example:"2"`
	Name       *string `json:"name" example:"哎呀又胖啦"`
	Phone      *string `json:"phone" example:"18888888888"`
	RegionCode *string `json:"region_code" example:"['370000', '370100', '370116']"`
	Region     *string `json:"region" example:"xxx xxx xxx"`
	Detail     *string `json:"detail" example:"xxxxxxxxxx"`
	Email      *string `json:"email" example:"xxxxx@xxx.xx"`
	// 地址类型 0 虚拟 / 1 实体
	Type *int `json:"type" binding:"omitempty,oneof=0 1" err:"oneof=11301" example:"1" enums:"0,1"`
	// 是否默认地址 0 否 / 1 是
	IsDefault *int `json:"is_default" binding:"omitempty,oneof=0 1" err:"oneof=11301" example:"1" enums:"0,1"`
}

// AddressDeleteAddressReq 删除收货地址请求入参
type AddressDeleteAddressReq struct {
	ID int64 `json:"id" binding:"required" err:"required=11301" example:"1"`
}
