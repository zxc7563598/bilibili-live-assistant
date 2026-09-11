package product

// 通用分页请求参数
type PageResp struct {
	PageNo    int     `json:"pageNo"`
	PageSize  int     `json:"pageSize"`
	SortField *string `json:"sortField"`
	SortOrder *string `json:"sortOrder"`
}

func (r *PageResp) OffsetLimit() (int, int, *string, *string) {
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
	return offset, r.PageSize, r.SortField, r.SortOrder
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
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Cover       string      `json:"cover"`
	Price       int64       `json:"price"`
	CreditType  int         `json:"credit_type"`
	Sold        int64       `json:"sold"`
	Stock       int64       `json:"stock"`
	Tags        string      `json:"tags"`
	Describe    string      `json:"describe"`
	SortOrder   int         `json:"sort_order"`
	Enable      bool        `json:"enable"`
	ProductType int         `json:"product_type"`
	Skus        []SkuItem   `json:"skus"`
	Specs       []SpecItem  `json:"specs"`
	Images      []ImageItem `json:"images"`
}

// SkuItem 商品 SKU
type SkuItem struct {
	ID             int64  `json:"id"`
	Price          int64  `json:"price"`
	CostPrice      int64  `json:"cost_price"`
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
	Enable    bool   `json:"enable"`
}

// SaveReq 创建或变更商品入参
type SaveReq struct {
	ID          int64
	Name        string
	Cover       string
	Price       int64
	CreditType  int
	ProductType int
	Sold        int64
	Tags        string
	Describe    string
	SortOrder   int
	Enable      bool
	Specs       []SaveSpec
	Skus        []SaveSku
	Images      []SaveImage
}

// SaveSpec 商品规格及规格值
type SaveSpec struct {
	KeyName string
	Values  []SaveSpecValue
}

// SaveSpecValue 商品规格值
type SaveSpecValue struct {
	ValueName string
}

// SaveSku 商品 SKU
type SaveSku struct {
	Price int64
	// CostPrice 成本价，仅后台核算使用，不对外展示
	CostPrice int64
	// Stock 为 nil 表示本次不改动库存，保留库里的最新值
	Stock *int64
	// SpecProperties 规格快照，保留请求里的原始形状，由 Service 拍平校验后序列化落库
	SpecProperties []map[string]string
}

// SaveImage 商品图片
type SaveImage struct {
	ImagePath string
	SortOrder int
	Type      int
	Enable    bool
}

// SaveResp 创建或变更商品出参
type SaveResp struct {
	ID int64
}
