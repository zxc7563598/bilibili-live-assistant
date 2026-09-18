package role

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/role"
)

// Handler 角色 HTTP 接口处理器
type Handler struct {
	roleSvc *role.Service
}

// New 创建 Handler 实例
func New(roleSvc *role.Service) *Handler {
	return &Handler{
		roleSvc: roleSvc,
	}
}

func toRoleListItems(list []role.ListPageItem) []resp.RoleListPageItem {
	res := make([]resp.RoleListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.RoleListPageItem{
			ID:            v.ID,
			Code:          v.Code,
			Name:          v.Name,
			Enable:        v.Enable,
			PermissionIds: v.PermissionIds,
		})
	}
	return res
}

func toRoleListAllItems(list []role.ListAllResp) []resp.RoleItem {
	res := make([]resp.RoleItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.RoleItem{
			ID:     v.ID,
			Code:   v.Code,
			Name:   v.Name,
			Enable: v.Enable,
		})
	}
	return res
}

// toRoleMenuItems 权限树转换：Service 出参 → resp.MenuItem
//
// 菜单树与权限树的结构完全一致，共用 resp.MenuItem。
func toRoleMenuItems(list []role.RoleMenuItem) []resp.MenuItem {
	res := make([]resp.MenuItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.MenuItem{
			ID:          v.ID,
			Code:        v.Code,
			Enable:      v.Enable,
			Show:        v.Show,
			KeepAlive:   v.KeepAlive,
			Layout:      v.Layout,
			Type:        v.Type,
			ParentID:    v.ParentID,
			Name:        v.Name,
			Icon:        v.Icon,
			Path:        v.Path,
			Component:   v.Component,
			Order:       v.Order,
			Redirect:    v.Redirect,
			Method:      v.Method,
			Description: v.Description,
			Children:    toRoleMenuItems(v.Children),
		})
	}
	return res
}
