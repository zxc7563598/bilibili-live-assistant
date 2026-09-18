package input

// ProductShopListPageReq 商城端获取主页商品分页列表请求
type ProductShopListPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=11001" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=11001" example:"20"`
	// 排序字段
	SortField *string `json:"sortField" example:"points"`
	// 排序方向 ascend/descend
	SortOrder *string `json:"sortOrder" example:"descend" enums:"ascend,descend"`
	// 商品名称，支持模糊搜索
	Name *string `json:"name" example:"测试"`
	// 货币类型
	CreditType *int `json:"credit_type" example:"1" enums:"0,1"`
}

// ProductDetailReq 商城端获取商品详细信息请求
type ProductDetailReq struct {
	// 商品ID
	ID int64 `json:"id" binding:"required" err:"required=11001" example:"1"`
}

// ProductAdminListPageReq 后台获取主页商品分页列表请求
type ProductAdminListPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=11001" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=11001" example:"20"`
	// 排序字段
	SortField *string `json:"sortField" example:"points"`
	// 排序方向 ascend/descend
	SortOrder *string `json:"sortOrder" example:"descend" enums:"ascend,descend"`
	// 商品名称，支持模糊搜索
	Name *string `json:"name" example:"测试"`
	// 货币类型
	CreditType *int `json:"credit_type" example:"1" enums:"0,1"`
	// 是否启用
	Enable *int `json:"enable" example:"1" enums:"0,1"`
}

// ProductUpdateEnableReq 后台变更商品是否启用请求
type ProductUpdateEnableReq struct {
	// 商品ID
	ID int64 `json:"id" binding:"required" err:"required=11001" example:"1"`
	// 是否启用；必传，避免调用方漏传后被当成下架
	Enable *bool `json:"enable" binding:"required" err:"required=11001" example:"true"`
}

// ProductSaveReq 后台创建或变更商品请求
//
// 子表（规格 / 规格值 / SKU / 图片）采用全量覆盖语义：请求里没有的即视为删除。
// 因此子表数据一律不接受 id；规格、规格值、图片每次保存都重建，
// SKU 则按规格快照做差量更新，快照相同的组合复用原行（ID 不变）。
// 商品库存 stock 不单独提交，由后端按落库后的 SKU 库存汇总。
type ProductSaveReq struct {
	// 商品ID，0 表示新增
	ID int64 `json:"id" example:"0"`
	// 商品名称
	Name string `json:"name" binding:"required,max=255" err:"required=11002,max=11003" example:"小立牌"`
	// 商品封面图
	Cover string `json:"cover" binding:"required,max=255" err:"required=11004,max=11005" example:"https://cdn.example.com/cover.png"`
	// 商品展示价格，单位分
	Price int64 `json:"price" binding:"gte=0" err:"gte=11006" example:"100"`
	// 积分类型，0 星光 / 1 积分
	CreditType int `json:"credit_type" binding:"oneof=0 1" err:"oneof=11007" example:"1" enums:"0,1"`
	// 商品类型，0 虚拟 / 1 实体
	ProductType int `json:"product_type" binding:"oneof=0 1" err:"oneof=11008" example:"1" enums:"0,1"`
	// 已售数量
	Sold int64 `json:"sold" binding:"gte=0" err:"gte=11009" example:"0"`
	// 商品标签，英文逗号隔开
	Tags string `json:"tags" binding:"max=255" err:"max=11010" example:"aa,bb"`
	// 商品说明
	Describe string `json:"describe" binding:"max=1000" err:"max=11011" example:"下单后 1-3 天发出"`
	// 排序，越大越靠前
	SortOrder int `json:"sort_order" example:"100"`
	// 是否启用；必传，避免调用方漏传后被当成下架
	Enable *bool `json:"enable" binding:"required" err:"required=11001" example:"true"`
	// 规格设置
	Specs []ProductSaveSpecReq `json:"specs" binding:"max=10" err:"max=11027"`
	// 上架 SKU，至少一个；未上架的规格组合不提交
	Skus []ProductSaveSkuReq `json:"skus" binding:"required,min=1,max=200" err:"required=11012,min=11012,max=11013"`
	// 商品图片，含轮播图与详情图
	Images []ProductSaveImageReq `json:"images" binding:"max=200" err:"max=11014"`
}

// ProductSaveSpecReq 商品规格及规格值
type ProductSaveSpecReq struct {
	// 规格名称
	KeyName string `json:"key_name" example:"类型"`
	// 规格值
	Values []ProductSaveSpecValueReq `json:"values"`
}

// ProductSaveSpecValueReq 商品规格值
type ProductSaveSpecValueReq struct {
	// 规格值内容
	ValueName string `json:"value_name" example:"镭射款"`
}

// ProductSaveSkuReq 商品 SKU
type ProductSaveSkuReq struct {
	// SKU 价格，单位分
	Price int64 `json:"price" example:"100"`
	// SKU 成本价，单位分，仅后台核算使用
	CostPrice int64 `json:"cost_price" example:"50"`
	// SKU 库存；不传表示本次不改动库存，后端保留库里的最新值，
	// 避免把保存期间被订单扣减的库存覆盖回去
	Stock *int64 `json:"stock" example:"500"`
	// 规格快照，形如 [{"类型":"镭射款"},{"形象":"wink小蓝"}]
	SpecProperties []map[string]string `json:"spec_properties"`
}

// ProductSaveImageReq 商品图片
type ProductSaveImageReq struct {
	// 图片路径
	ImagePath string `json:"image_path" example:"https://cdn.example.com/1.png"`
	// 排序，越大越靠前
	SortOrder int `json:"sort_order" example:"100"`
	// 图片类型，0 轮播图 / 1 详情图
	Type int `json:"type" example:"0" enums:"0,1"`
	// 是否启用
	Enable bool `json:"enable" example:"true"`
}
