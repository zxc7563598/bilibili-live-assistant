package role

import (
	"github.com/zxc7563598/GoAdminKit/internal/dto/resp"
	"github.com/zxc7563598/GoAdminKit/internal/service/role"
)

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

func toRoleListAllItems(list []role.ListAllResp) []resp.RoleListAllItem {
	res := make([]resp.RoleListAllItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.RoleListAllItem{
			ID:     v.ID,
			Code:   v.Code,
			Name:   v.Name,
			Enable: v.Enable,
		})
	}
	return res
}

func toRoleMenuItem(list []role.RoleMenuItem) []resp.RoleMenuItem {
	res := make([]resp.RoleMenuItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.RoleMenuItem{
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
			Children:    toRoleMenuItem(v.Children),
		})
	}
	return res
}
