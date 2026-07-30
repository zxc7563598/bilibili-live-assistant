package admin

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

type TokenResp struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Login 请求返回
type LoginResp struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// RefreshLogin 请求返回
type RefreshLoginResp struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// SwitchRoleResp 请求返回
type SwitchRoleResp struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// ListPage 请求入参
type ListPageReq struct {
	PageResp
	Username *string `json:"username"`
	Gender   *int    `json:"gender"`
	Enable   *int    `json:"enable"`
}

// ListPage 请求返回
type ListPageResp struct {
	Total    int64 `json:"total"`
	PageData []ListPageItem
}

type ListPageItem struct {
	ID        uint64     `json:"id"`
	Username  string     `json:"username"`
	Enable    bool       `json:"enable"`
	Gender    int        `json:"gender"`
	Avatar    string     `json:"avatar"`
	Address   string     `json:"address"`
	Email     string     `json:"email"`
	Roles     []RoleItem `json:"roles"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// Details 请求返回
type DetailsResp struct {
	ID          uint64             `json:"id"`
	Username    string             `json:"username"`
	Enable      bool               `json:"enable"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	Profile     DetailsProfileItem `json:"profile"`
	Roles       []RoleItem         `json:"roles"`
	CurrentRole RoleItem           `json:"currentRole"`
}

type RoleItem struct {
	ID     uint64 `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Enable bool   `json:"enable"`
}

type DetailsProfileItem struct {
	ID       uint64 `json:"id"`
	Nickname string `json:"nickName"`
	Gender   int    `json:"gender"`
	Avatar   string `json:"avatar"`
	Address  string `json:"address"`
	Email    string `json:"email"`
}

// Save 请求入参
type SaveReq struct {
	ID       *uint64  `json:"id"`
	Enable   *bool    `json:"enable"`
	Username *string  `json:"username"`
	Password *string  `json:"password"`
	RoleIds  []uint64 `json:"roleIds"`
}

// UpdateProfile 请求入参
type UpdateProfileReq struct {
	ID       uint64  `json:"id"`
	Nickname *string `json:"nickName"`
	Gender   *int    `json:"gender"`
	Address  *string `json:"address"`
	Email    *string `json:"email"`
	Avatar   *string `json:"avatar"`
}
