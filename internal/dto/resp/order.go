package resp

// OrderGetConfirmResp 获取用户下单数据请求返回
type OrderGetConfirmResp struct {
	// id
	ID int64 `json:"id" example:"1"`
	// 到期时间(毫秒级时间戳)
	ExpireAt int64 `json:"expire_at" example:"1788417485000"`
	// 产品信息
	Product ProductItem `json:"product"`
}

// ProductItem 单个产品信息
type ProductItem struct {
	// 产品ID
	ID int64 `json:"id" example:"2"`
	// 产品名称
	Name string `json:"name" example:"蓝牙耳机"`
	// 产品封面
	Cover string `json:"cover" example:"https://cdn.hejunjie.life/avatars/oneadmin.jpeg"`
	// 产品价格
	Price int64 `json:"price" example:"30"`
	// 产品价格类型
	CreditType int `json:"credit_type" example:"1" enums:"0,1"`
	// 产品类型
	ProductType int `json:"product_type" example:"1" enums:"0,1"`
	// 产品SKU
	Sku string `json:"sku" example:"[{'aa':'bb'},{'aa':'bb'}]"`
	// 购买数量
	Count int64 `json:"count" example:"0"`
}
