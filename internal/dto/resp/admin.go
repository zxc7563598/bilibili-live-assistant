package resp

// AdminCaptchaResp 获取验证码启用状态返回
type AdminCaptchaResp struct {
	// 是否启用验证码
	Enabled bool `json:"enabled" example:"true"`
}

// AdminLoginResp 管理员登录/刷新登录返回
type AdminLoginResp struct {
	// access token
	AccessToken string `json:"accessToken" example:"Bearer xxxxxxxxxx"`
	// refresh token
	RefreshToken string `json:"refreshToken" example:"Bearer xxxxxxxxxx"`
}

// AdminListPageResp 获取管理员分页列表返回
type AdminListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []AdminListPageItem `json:"pageData"`
}

type AdminListPageItem struct {
	// 管理员ID
	ID int64 `json:"id" example:"1"`
	// 账号
	Username string `json:"username" example:"admin"`
	// 状态
	Enable bool `json:"enable" example:"true"`
	// 性别
	Gender int `json:"gender" example:"1" enums:"0,1,2"`
	// 头像地址
	Avatar string `json:"avatar" example:"https://cdn.hejunjie.life/avatars/oneadmin.jpeg"`
	// 居住地址
	Address string `json:"address" example:"xxxxxxx"`
	// email
	Email string `json:"email" example:"xxxxx@xxx.com"`
	// 拥有角色列表
	Roles []AdminDetailsRoleItem `json:"roles"`
	// 创建时间
	CreatedAt string `json:"createdAt"`
	// 更新时间
	UpdatedAt string `json:"updatedAt"`
}

// AdminDetailsResp 获取单个管理员详细信息返回
type AdminDetailsResp struct {
	// 管理员ID
	ID int64 `json:"id" example:"1"`
	// 账号
	Username string `json:"username" example:"admin"`
	// 状态
	Enable bool `json:"enable" example:"true"`
	// 创建时间
	CreatedAt string `json:"createdAt" example:"2026-03-25 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updatedAt" example:"2026-03-25 12:00:00"`
	// 个人资料
	Profile AdminDetailsProfileItem `json:"profile"`
	// 拥有角色列表
	Roles []AdminDetailsRoleItem `json:"roles"`
	// 当前角色信息
	CurrentRole AdminDetailsRoleItem `json:"currentRole"`
}

type AdminDetailsRoleItem struct {
	// 角色ID
	ID int64 `json:"id" example:"1"`
	// 角色标识
	Code string `json:"code" example:"SUPER_ADMIN"`
	// 角色名称
	Name string `json:"name" example:"超级管理员"`
	// 状态
	Enable bool `json:"enable" example:"true"`
}

type AdminDetailsProfileItem struct {
	// 管理员ID
	ID int64 `json:"id" example:"1"`
	// 名称
	Nickname string `json:"nickName" example:"test name"`
	// 性别
	Gender int `json:"gender" example:"1" enums:"0,1,2"`
	// 头像
	Avatar string `json:"avatar" example:"https://cdn.hejunjie.life/avatars/oneadmin.jpeg"`
	// 居住地址
	Address string `json:"address" example:"xxxxxxx"`
	// email
	Email string `json:"email" example:"xxxxx@xxx.com"`
}
