package input

// RoleListPageReq 获取角色分页列表请求
type RoleListPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=10201" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=10201" example:"20"`
	// 名称，支持模糊搜索
	Name *string `json:"name" example:"超级"`
	// 状态
	Enable *int `json:"enable" example:"1" enums:"0,1"`
}

// RoleSaveReq 处理创建或变更角色请求
type RoleSaveReq struct {
	// 角色ID
	ID *uint64 `json:"id" example:"2"`
	// 角色标识
	Code *string `json:"code" example:"TestAdmin"`
	// 角色名称
	Name *string `json:"name" example:"测试管理员"`
	// 状态
	Enable *bool `json:"enable" example:"true"`
	// 角色绑定菜单ID
	PermissionIds *[]uint64 `json:"permissionIds" example:"1,2,3"`
}

// RoleDeleteReq 删除角色请求
type RoleDeleteReq struct {
	// 角色ID
	ID uint64 `json:"id" binding:"required" err:"required=10205" example:"2"`
}

// RoleAddRoleUsersReq 分配角色到管理员请求
type RoleAddRoleUsersReq struct {
	// 角色ID
	RoleID uint64 `json:"roleId" binding:"required" err:"required=10205" example:"2"`
	// 管理员ID（支持多个）
	AdminIds []uint64 `json:"adminIds" binding:"required" err:"required=10207" example:"1,2,3"`
}

// RoleRemoveRoleUsersReq 分配角色到管理员请求
type RoleRemoveRoleUsersReq struct {
	// 角色ID
	RoleID uint64 `json:"roleId" binding:"required" err:"required=10205" example:"2"`
	// 管理员ID（支持多个）
	AdminIds []uint64 `json:"adminIds" binding:"required" err:"required=10207" example:"1,2,3"`
}
