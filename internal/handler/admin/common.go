package admin

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/admin"
	altchaSvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/altcha"
)

// Handler 管理员认证与账号管理 HTTP 接口处理器
type Handler struct {
	adminSvc  *admin.Service
	altchaSvc *altchaSvc.Service
}

// New 创建 Handler 实例
func New(adminSvc *admin.Service, altchaSvc *altchaSvc.Service) *Handler {
	return &Handler{
		adminSvc:  adminSvc,
		altchaSvc: altchaSvc,
	}
}

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
