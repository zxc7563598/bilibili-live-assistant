package resp

// RoleListPageResp 获取管理员分页列表返回
type RoleListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []RoleListPageItem `json:"pageData"`
}

type RoleListPageItem struct {
	// 角色ID
	ID uint64 `json:"id" example:"1"`
	// 角色标识
	Code string `json:"code" example:"SUPER_ADMIN"`
	// 角色名称
	Name string `json:"name" example:"超级管理员"`
	// 状态
	Enable bool `json:"enable" example:"true"`
	// 角色绑定菜单ID
	PermissionIds []uint64 `json:"permissionIds"`
}

// RoleListAllResp 获取全部角色请求返回
type RoleListAllResp struct {
	// 角色列表
	List []RoleListAllItem `json:"list"`
}

type RoleListAllItem struct {
	// 角色ID
	ID uint64 `json:"id" example:"1"`
	// 角色标识
	Code string `json:"code" example:"SUPER_ADMIN"`
	// 角色名称
	Name string `json:"name" example:"超级管理员"`
	// 状态
	Enable bool `json:"enable" example:"true"`
}

// RolePermissionsResp 获取管理员菜单权限树请求返回
type RolePermissionsResp struct {
	Menu []RoleMenuItem `json:"menu"`
}

type RoleMenuItem struct {
	// 菜单ID
	ID uint64 `json:"id" example:"1"`
	// 菜单标识
	Code string `json:"code" example:"Base"`
	// 状态
	Enable bool `json:"enable" example:"true"`
	// 显示状态
	Show bool `json:"show" example:"true"`
	// 保持活跃
	KeepAlive bool `json:"keepAlive" example:"false"`
	// 布局
	Layout string `json:"layout" example:"full"`
	// 菜单类型
	Type string `json:"type" example:"MENU" enums:"BUTTON,MENU"`
	// 父级ID
	ParentID uint64 `json:"parentId" example:"0"`
	// 菜单名称
	Name string `json:"name" example:"基础菜单"`
	// 菜单图标
	Icon string `json:"icon" example:"i-fe:list"`
	// 菜单路径
	Path string `json:"path" example:"/path/url"`
	// 组件路径
	Component string `json:"component" example:"/src/list/list.vue"`
	// 排序（从小到大）
	Order int `json:"order" example:"0"`
	// 重定向
	Redirect string `json:"redirect"`
	// 方法
	Method string `json:"method"`
	// 描述
	Description string `json:"description"`
	// 子菜单
	Children []RoleMenuItem `json:"children"`
}
