package input

// AdminLoginReq 管理员登录请求
type AdminLoginReq struct {
	// 账号
	Username string `json:"username" binding:"required,min=1" err:"required=10102,min=10102" example:"admin"`
	// 密码
	Password string `json:"password" binding:"required,min=6,max=32" err:"required=10103,min=10104,max=10105" example:"123456"`
	// 验证码
	Captcha *string `json:"captcha"`
}

// AdminRefreshReq 刷新管理员登录凭证请求
type AdminRefreshReq struct {
	// refresh token
	Token string `json:"token" binding:"required" err:"required=10101" example:"Bearer xxxxxxxxxx"`
}

// AdminSwitchRoleReq 切换管理员角色请求
type AdminSwitchRoleReq struct {
	// 角色标识
	Code string `json:"code" binding:"required" err:"required=10106" example:"SUPER_ADMIN"`
}

// AdminChangePasswordReq 根据旧密码变更新密码请求
type AdminChangePasswordReq struct {
	// 旧密码
	OldPassword string `json:"oldPassword" binding:"required,min=6,max=32" err:"required=10103,min=10104,max=10105" example:"123456"`
	// 新密码
	NewPassword string `json:"newPassword" binding:"required,min=6,max=32" err:"required=10103,min=10104,max=10105" example:"654321"`
}

// AdminListPageReq 获取管理员分页列表请求
type AdminListPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=10101" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=10101" example:"20"`
	// 用户名，支持模糊搜索
	Username *string `json:"username" example:"admin"`
	// 性别
	Gender *int `json:"gender" example:"1" enums:"0,1,2"`
	// 是否启用
	Enable *int `json:"enable" example:"1" enums:"0,1"`
}

// AdminSaveReq 变更管理员信息请求
type AdminSaveReq struct {
	// 管理员ID
	ID *uint64 `json:"id" example:"1"`
	// 状态
	Enable *bool `json:"enable" example:"true"`
	// 账号
	Username *string `json:"username" binding:"required,min=1" err:"required=10102,min=10102" example:"admin"`
	// 密码
	Password *string `json:"password" example:"123456"`
	// 角色ID组
	RoleIds []uint64 `json:"roleIds" example:"1,2"`
}

// AdminDeleteReq 删除管理员信息请求
type AdminDeleteReq struct {
	// 管理员ID
	ID uint64 `json:"id" binding:"required" err:"required=10101" example:"2"`
}

// AdminResetAdminPasswordReq 变更管理员密码请求
type AdminResetAdminPasswordReq struct {
	// 管理员ID
	ID uint64 `json:"id" binding:"required" example:"1"`
	// 密码
	Password string `json:"password" binding:"required,min=6,max=32" err:"required=10103,min=10104,max=10105" example:"123456"`
}

// AdminUpdateProfileReq 修改管理员个人信息请求
type AdminUpdateProfileReq struct {
	// 管理员ID
	ID uint64 `json:"id" binding:"required" err:"required=10101" example:"1"`
	// 名称
	Nickname *string `json:"nickName" example:"test name"`
	// 性别
	Gender *int `json:"gender" example:"1" enums:"0,1,2"`
	// 居住地址
	Address *string `json:"address" example:"xxxxxxx"`
	// email
	Email *string `json:"email" example:"xxxxx@xxx.com"`
	// 头像地址
	Avatar *string `json:"avatar" example:"https://cdn.hejunjie.life/avatars/oneadmin.jpeg"`
}
