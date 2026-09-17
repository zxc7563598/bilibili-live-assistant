package resp

// AddressItem 收货地址信息
type AddressItem struct {
	// id
	ID int64 `json:"id" example:"1"`
	// 收货人姓名
	Name string `json:"name" example:"哎呀又胖啦"`
	// 收货人手机号
	Phone string `json:"phone" example:"18888888888"`
	// 地区code
	RegionCode string `json:"region_code" example:"['370000', '370100', '370116']"`
	// 地区文字描述
	Region string `json:"region" example:"xxxxx xxxxx xxxxx"`
	// 详细地址
	Detail string `json:"detail" example:"xxxxxxxxxxxxxxx"`
	// 邮箱地址
	Email string `json:"email" example:"xxxxxxxx@xxx.xx"`
	// 类型
	Type int `json:"type" example:"1" enums:"0,1"`
	// 默认地址
	IsDefault int `json:"is_default" example:"1" enums:"0,1"`
}

// AddressGetDefaultAddressResp 获取用户指定类型的默认收货地址返回
type AddressGetDefaultAddressResp struct {
	AddressItem
}

// AddressGetAddressListResp 获取用户收货地址列表返回
type AddressGetAddressListResp struct {
	// 收货地址列表
	List []AddressItem `json:"list"`
}

// AddressGetAddressByIDResp 获取收货地址详细信息返回
type AddressGetAddressByIDResp struct {
	AddressItem
}
