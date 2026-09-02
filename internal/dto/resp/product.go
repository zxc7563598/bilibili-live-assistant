package resp

// ProductListPageResp 商城端获取主页商品分页列表返回
type ProductListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []ProductListPageItem `json:"pageData"`
}

type ProductListPageItem struct {
	// 管理员ID
	ID int64 `json:"id" example:"1"`
	// 商品名称
	Name string `json:"name" example:"小立牌"`
	// 商品封面URL
	Cover string `json:"cover" example:"https://cdn.hejunjie.life/avatars/oneadmin.jpeg"`
	// 价格
	Price int64 `json:"price" example:"100"`
	// 价格类型
	CreditType int `json:"credit_type" example:"1" enums:"0,1"`
	// 已售
	Sold int64 `json:"sold" example:"100"`
	// 库存
	Stock int64 `json:"stock" example:"100"`
	// 商品标签，英文逗号隔开
	Tags string `json:"tags" example:"aa,bb,cc,dd"`
	// 商品说明
	Describe string `json:"describe" example:"xxxxxxxxxxxxxxxxx"`
}

// ProductDetailResp 商城端获取商品详细信息返回
type ProductDetailResp struct {
	// 商品ID
	ID int64 `json:"id" example:"1"`
	// 商品名称
	Name string `json:"name" example:"小立牌"`
	// 商品封面URL
	Cover string `json:"cover" example:"https://cdn.hejunjie.life/avatars/oneadmin.jpeg"`
	// 价格
	Price int64 `json:"price" example:"100"`
	// 价格类型
	CreditType int `json:"credit_type" example:"1" enums:"0,1"`
	// 已售
	Sold int64 `json:"sold" example:"100"`
	// 库存
	Stock int64 `json:"stock" example:"100"`
	// 商品标签，英文逗号隔开
	Tags string `json:"tags" example:"aa,bb,cc,dd"`
	// 商品说明
	Describe string `json:"describe" example:"xxxxxxxxxxxxxxxxx"`
	// 商品SKU信息
	Skus []SkuItem `json:"skus"`
	// 商品规格
	Specs []SpecItem `json:"specs"`
	// 商品详情图
	Images []ImageItem `json:"images"`
}

// SkuItem 商品 SKU
type SkuItem struct {
	// SKU ID
	ID int64 `json:"id" example:"1"`
	// SKU 价格
	Price int64 `json:"price" example:"100"`
	// SKU 库存
	Stock int64 `json:"stock" example:"100"`
	// SKU 规格快照
	SpecProperties string `json:"spec_properties" example:"[{'aa':'bb'},{'aa':'bb'}]"`
}

// SpecItem 商品规格及其可选值
type SpecItem struct {
	// 规格ID
	ID int64 `json:"id" example:"1"`
	// 规格名称
	KeyName string `json:"key_name" example:"aa"`
	// 规格值
	Values []SpecValue `json:"values"`
}

// SpecValue 规格可选值
type SpecValue struct {
	// 规格值ID
	ID int64 `json:"id" example:"1"`
	// 规格值内容
	ValueName string `json:"value_name" example:"bb"`
}

// ImageItem 商品图片
type ImageItem struct {
	// 图片ID
	ID int64 `json:"id" example:"1"`
	// 图片URL
	ImagePath string `json:"image_path" example:"https://cdn.hejunjie.life/avatars/oneadmin.jpeg"`
	// 图片位置
	Type int `json:"type" example:"1" enums:"0,1"`
}
