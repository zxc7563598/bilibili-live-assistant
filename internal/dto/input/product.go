package input

// ProductListPageReq 商城端获取主页商品分页列表请求
type ProductListPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=10101" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=10101" example:"20"`
	// 商品名称，支持模糊搜索
	Name *string `json:"name" example:"测试"`
	// 货币类型
	CreditType *int `json:"credit_type" example:"1" enums:"0,1"`
}

// ProductDetailReq 商城端获取商品详细信息请求
type ProductDetailReq struct {
	// 商品ID
	ID int64 `json:"id" binding:"required" err:"required=10101" example:"1"`
}
