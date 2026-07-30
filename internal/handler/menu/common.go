package menu

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/menu"
)

func toMenuItem(list []menu.MenuItem) []resp.MenuItem {
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
			Children:    toMenuItem(v.Children),
		})
	}
	return res
}
