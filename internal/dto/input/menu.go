package input

// 验证菜单是否存在请求
type MenuValidateReq struct {
	// 菜单路径
	Path string `json:"path" binding:"required,min=1" err:"required=10301,min=10302" example:"/path/url"`
}

// 获取菜单下的按钮请求
type MenuButtonsReq struct {
	// 菜单ID（作为父级ID进行查询）
	ParentID uint64 `json:"parentId" binding:"required" err:"required=10301" example:"0"`
}

// 添加或变更菜单信息请求
type MenuSaveReq struct {
	// 菜单ID
	ID *uint64 `json:"id" example:"1"`
	// 菜单标识
	Code string `json:"code" binding:"required,min=1" err:"required=10301,min=10303" example:"Base"`
	// 状态
	Enable bool `json:"enable" example:"true"`
	// 显示状态
	Show bool `json:"show" example:"true"`
	// 保持活跃
	KeepAlive bool `json:"keepAlive" example:"false"`
	// 布局
	Layout string `json:"layout" example:"full"`
	// 菜单类型
	Type string `json:"type" binding:"required,min=1" err:"required=10301,min=10304" example:"MENU" enums:"BUTTON,MENU"`
	// 父级ID
	ParentID uint64 `json:"parentId" example:"0"`
	// 菜单名称
	Name string `json:"name" binding:"required,min=1" err:"required=10301,min=10305" example:"基础菜单"`
	// 菜单图标
	Icon string `json:"icon" example:"i-fe:list"`
	// 菜单路径
	Path string `json:"path" example:"/path/url"`
	// 组件路径
	Component string `json:"component" example:"/src/list/list.vue"`
	// 排序（从小到大）
	Order int `json:"order" example:"0"`
}

// 快速切换菜单的启用状态请求
type MenuToggleReq struct {
	// 菜单ID
	ID uint64 `json:"id" binding:"required" err:"required=10301" example:"1"`
}

// 删除菜单请求
type MenuDeleteReq struct {
	// 菜单ID
	ID uint64 `json:"id" binding:"required" err:"required=10301" example:"1"`
}
