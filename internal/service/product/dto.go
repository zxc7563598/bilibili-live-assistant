package product

// 通用分页请求参数
type PageResp struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

func (r *PageResp) OffsetLimit() (int, int) {
	if r.PageNo < 1 {
		r.PageNo = 1
	}
	if r.PageSize < 1 {
		r.PageSize = 10
	}
	if r.PageSize > 100 {
		r.PageSize = 100
	}
	offset := (r.PageNo - 1) * r.PageSize
	return offset, r.PageSize
}

// ListPage 请求入参
type ListPageReq struct {
	PageResp
	Name       *string `json:"name"`
	CreditType *int    `json:"credit_type"`
	Enable     *int    `json:"enable"`
}

// ListPage 请求返回
type ListPageResp struct {
	Total    int64 `json:"total"`
	PageData []ListPageItem
}

type ListPageItem struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Cover      string `json:"cover"`
	Price      int64  `json:"price"`
	CreditType int    `json:"credit_type"`
	Sold       int64  `json:"sold"`
	Stock      int64  `json:"stock"`
	Tags       string `json:"tags"`
	Describe   string `json:"describe"`
	SortOrder  int    `json:"sort_order"`
	Enable     bool   `json:"enable"`
}

// DetailsResp 商品详情返回，含商品基础信息、SKU、规格、图片
type DetailsResp struct {
	ID         int64       `json:"id"`
	Name       string      `json:"name"`
	Cover      string      `json:"cover"`
	Price      int64       `json:"price"`
	CreditType int         `json:"credit_type"`
	Sold       int64       `json:"sold"`
	Stock      int64       `json:"stock"`
	Tags       string      `json:"tags"`
	Describe   string      `json:"describe"`
	SortOrder  int         `json:"sort_order"`
	Enable     bool        `json:"enable"`
	Skus       []SkuItem   `json:"skus"`
	Specs      []SpecItem  `json:"specs"`
	Images     []ImageItem `json:"images"`
}

// SkuItem 商品 SKU
type SkuItem struct {
	ID             int64  `json:"id"`
	Price          int64  `json:"price"`
	Stock          int64  `json:"stock"`
	SpecProperties string `json:"spec_properties"`
}

// SpecItem 商品规格及其可选值
type SpecItem struct {
	ID      int64       `json:"id"`
	KeyName string      `json:"key_name"`
	Values  []SpecValue `json:"values"`
}

// SpecValue 规格可选值
type SpecValue struct {
	ID        int64  `json:"id"`
	ValueName string `json:"value_name"`
}

// ImageItem 商品图片
type ImageItem struct {
	ID        int64  `json:"id"`
	ImagePath string `json:"image_path"`
	SortOrder int    `json:"sort_order"`
	Type      int    `json:"type"`
}
