package resp

// RoleListPageResp 后台分页查询角色返回
type RoleListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []RoleListPageItem `json:"pageData"`
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
	List []RoleListAllItem `json:"list"`
}

// RoleListAllItem 下拉用角色项
type RoleListAllItem struct {
	// 角色ID
	ID int64 `json:"id" example:"1"`
	// 角色标识
	Code string `json:"code" example:"SUPER_ADMIN"`
	// 角色名称
	Name string `json:"name" example:"超级管理员"`
	// 状态
	Enable bool `json:"enable" example:"true"`
}

// RolePermissionsResp 获取当前管理员的菜单权限树返回
//
// 只含当前登录管理员有权访问的菜单，用于前端生成动态路由与按钮级权限。
type RolePermissionsResp struct {
	// 菜单权限树（含按钮类型节点）
	Menu []RoleMenuItem `json:"menu"`
}

// RoleMenuItem 权限树节点，结构同 resp.MenuItem，按层级递归
//
// 不能直接复用 MenuItem：这里是按权限过滤后的子树，Children 递归同一类型。
type RoleMenuItem struct {
	// 菜单ID
	ID int64 `json:"id" example:"1"`
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
	ParentID int64 `json:"parentId" example:"0"`
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
