package address

// AddressReq 保存地址的入参；type / is_default 为指针，省略时更新保留原值、新增走默认
type AddressReq struct {
	ID         *int64  `json:"id"`
	Name       *string `json:"name"`
	Phone      *string `json:"phone"`
	RegionCode *string `json:"region_code"`
	Region     *string `json:"region"`
	Detail     *string `json:"detail"`
	Email      *string `json:"email"`
	Type       *int    `json:"type"`
	IsDefault  *int    `json:"is_default"`
}

type AddressItem struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	RegionCode string `json:"region_code"`
	Region     string `json:"region"`
	Detail     string `json:"detail"`
	Email      string `json:"email"`
	Type       int    `json:"type"`
	IsDefault  int    `json:"is_default"`
}
