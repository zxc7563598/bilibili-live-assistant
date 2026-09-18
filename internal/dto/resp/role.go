package resp

// RoleListPageResp 后台分页查询角色返回
type RoleListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []RoleListPageItem `json:"pageData"`
}

// RoleItem 角色项：管理员详情里的角色、以及角色下拉共用
type RoleItem struct {
	// 角色ID
	ID int64 `json:"id" example:"1"`
	// 角色标识
	Code string `json:"code" example:"SUPER_ADMIN"`
	// 角色名称
	Name string `json:"name" example:"超级管理员"`
	// 状态
	Enable bool `json:"enable" example:"true"`
}

// RoleListPageItem 角色列表中的单条角色
type RoleListPageItem struct {
	// 角色ID
	ID int64 `json:"id" example:"1"`
	// 角色标识
	Code string `json:"code" example:"SUPER_ADMIN"`
	// 角色名称
	Name string `json:"name" example:"超级管理员"`
	// 状态
	Enable bool `json:"enable" example:"true"`
	// 角色绑定菜单ID
	PermissionIds []int64 `json:"permissionIds"`
}

// RoleListAllResp 获取全部启用角色返回
//
// 供分配角色等下拉场景使用，不带分页与权限明细。
type RoleListAllResp struct {
	// 角色列表
	List []RoleItem `json:"list"`
}

// RoleListAllItem 下拉用角色项

// RolePermissionsResp 获取当前管理员的菜单权限树返回
//
// 只含当前登录管理员有权访问的菜单，用于前端生成动态路由与按钮级权限。
type RolePermissionsResp struct {
	// 菜单权限树（含按钮类型节点），结构同 MenuListResp.Menu
	Menu []MenuItem `json:"menu"`
}
