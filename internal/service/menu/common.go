package menu

import (
	"context"
	"sort"

	"github.com/zxc7563598/GoAdminKit/internal/enum"
	"github.com/zxc7563598/GoAdminKit/internal/model"
)

// buildTree 用于获取菜单权限树
func (s *Service) buildTree(list []MenuItem, parentID uint64) []MenuItem {
	// 构建 parent -> children map
	childrenMap := make(map[uint64][]MenuItem)
	for _, el := range list {
		childrenMap[el.ParentID] = append(childrenMap[el.ParentID], el)
	}
	// 递归
	var build func(pid uint64) []MenuItem
	build = func(pid uint64) []MenuItem {
		branch := childrenMap[pid]
		// 排序
		sort.Slice(branch, func(i, j int) bool {
			return branch[i].Order < branch[j].Order
		})
		// 构建 children
		for i := range branch {
			branch[i].Children = build(branch[i].ID)
		}
		return branch
	}
	return build(parentID)
}

func (s *Service) add(ctx context.Context, req SaveReq) (int, error) {
	_, err := s.menuRepo.Create(ctx, nil, &model.Menu{
		Code:      req.Code,
		Enable:    enum.BoolToEnable(req.Enable),
		Show:      enum.BoolToYesNo(req.Show),
		KeepAlive: enum.BoolToYesNo(req.KeepAlive),
		Layout:    req.Layout,
		Type:      enum.MenuType(req.Type),
		ParentID:  req.ParentID,
		Name:      req.Name,
		Icon:      req.Icon,
		Path:      req.Path,
		Component: req.Component,
		Order:     req.Order,
	})
	if err != nil {
		return 60302, err
	}
	return 0, nil
}

func (s *Service) update(ctx context.Context, req SaveReq) (int, error) {
	if err := s.menuRepo.UpdateByID(ctx, nil, *req.ID, model.MenuUpdateByIdForm{
		Code:      req.Code,
		Enable:    enum.BoolToEnable(req.Enable),
		Show:      enum.BoolToYesNo(req.Show),
		KeepAlive: enum.BoolToYesNo(req.KeepAlive),
		Layout:    req.Layout,
		Type:      enum.MenuType(req.Type),
		ParentID:  req.ParentID,
		Name:      req.Name,
		Icon:      req.Icon,
		Path:      req.Path,
		Component: req.Component,
		Order:     req.Order,
	}); err != nil {
		return 60303, err
	}
	return 0, nil
}

func toMenuItems(menus []model.Menu) []MenuItem {
	list := make([]MenuItem, 0, len(menus))
	for _, v := range menus {
		list = append(list, MenuItem{
			ID:        v.ID,
			Code:      v.Code,
			Enable:    v.Enable == enum.EnableEnable,
			Show:      v.Show == enum.Yes,
			KeepAlive: v.KeepAlive == enum.Yes,
			Layout:    v.Layout,
			Type:      string(v.Type),
			ParentID:  v.ParentID,
			Name:      v.Name,
			Icon:      v.Icon,
			Path:      v.Path,
			Component: v.Component,
			Order:     v.Order,
		})
	}
	return list
}
