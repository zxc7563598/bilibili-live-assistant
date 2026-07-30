package admin

import (
	"github.com/zxc7563598/GoAdminKit/internal/dto/resp"
	"github.com/zxc7563598/GoAdminKit/internal/service/admin"
)

func toAdminListItems(list []admin.ListPageItem) []resp.AdminListPageItem {
	res := make([]resp.AdminListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.AdminListPageItem{
			ID:        v.ID,
			Username:  v.Username,
			Enable:    v.Enable,
			Gender:    v.Gender,
			Avatar:    v.Avatar,
			Address:   v.Address,
			Email:     v.Email,
			Roles:     toAdminDetailsRoleItem(v.Roles),
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		})
	}
	return res
}

func toAdminDetailsRoleItem(list []admin.RoleItem) []resp.AdminDetailsRoleItem {
	res := make([]resp.AdminDetailsRoleItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.AdminDetailsRoleItem{
			ID:     v.ID,
			Code:   v.Code,
			Name:   v.Name,
			Enable: v.Enable,
		})
	}
	return res
}
