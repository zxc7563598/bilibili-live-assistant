package resp

// MenuItem 单条菜单信息
type MenuItem struct {
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
	Children []MenuItem `json:"children"`
}

// MenuListResp 获取全部菜单返回
type MenuListResp struct {
	// 菜单列表
	Menu []MenuItem `json:"menu"`
}

// MenuValidateResp 验证菜单是否存在返回
type MenuValidateResp struct {
	// 是否存在
	Has bool `json:"has" example:"true"`
}

// MenuButtonsResp 获取菜单下的按钮返回
type MenuButtonsResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []MenuItem `json:"pageData"`
}
