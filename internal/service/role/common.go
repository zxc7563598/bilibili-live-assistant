package role

import (
	"context"
	"errors"
	"sort"

	"github.com/zxc7563598/GoAdminKit/internal/enum"
	"github.com/zxc7563598/GoAdminKit/internal/model"
	"github.com/zxc7563598/GoAdminKit/pkg/jwt"
	"gorm.io/gorm"
)

// add 用于添加角色基本信息
func (s *Service) add(ctx context.Context, tx *gorm.DB, req SaveReq) (uint64, error) {
	if v := req.Code; v == nil || *v == "" {
		return 0, errors.New("code 不允许为空")
	}
	if v := req.Name; v == nil || *v == "" {
		return 0, errors.New("name 不允许为空")
	}
	if v := req.Enable; v == nil {
		return 0, errors.New("enable 不允许为空")
	}
	// 添加数据
	role, err := s.roleRepo.Create(ctx, tx, &model.Role{
		Code:   *req.Code,
		Name:   *req.Name,
		Enable: *enum.BoolToEnablePtr(req.Enable),
	})
	if err != nil {
		return 0, err
	}
	// 返回结果
	return role.ID, nil
}

// update 用于变更角色基本信息
func (s *Service) update(ctx context.Context, tx *gorm.DB, req SaveReq) (uint64, error) {
	// 修改数据
	if err := s.roleRepo.UpdateByID(ctx, tx, *req.ID, model.RoleUpdateByIdForm{
		Code:   req.Code,
		Name:   req.Name,
		Enable: enum.BoolToEnablePtr(req.Enable),
	}); err != nil {
		return 0, err
	}
	// 返回结果
	return *req.ID, nil
}

// resetRoleMenus 用于重新绑定角色对应的权限
func (s *Service) resetRoleMenus(ctx context.Context, tx *gorm.DB, roleID uint64, menuIDs []uint64) error {
	// 删除旧关系
	if err := s.roleMenuRepo.DeleteByRoleID(ctx, tx, roleID); err != nil {
		return err
	}
	// 没有菜单，直接结束
	if len(menuIDs) == 0 {
		return nil
	}
	// 构建新关系
	roleMenus := make([]model.RoleMenu, 0, len(menuIDs))
	for _, v := range menuIDs {
		roleMenus = append(roleMenus, model.RoleMenu{
			RoleID: roleID,
			MenuID: v,
		})
	}
	// 批量插入
	if err := s.roleMenuRepo.CreateBatch(ctx, tx, roleMenus); err != nil {
		return err
	}
	return nil
}

// getMenusByRole 用于获取指定角色拥有的菜单
func (s *Service) getMenusByRole(ctx context.Context, roleID uint64, roleCode string) ([]model.Menu, error) {
	// 超级管理员默认拥有全部菜单
	if roleCode == RoleCodeSuperAdmin {
		return s.menuRepo.ListEnabled(ctx, nil)
	}
	// 非超管，根据角色ID获取菜单
	roleMenus, err := s.roleMenuRepo.ListByRoleID(ctx, nil, roleID)
	if err != nil {
		return nil, err
	}
	menuIDs := make([]uint64, 0, len(roleMenus))
	for _, v := range roleMenus {
		menuIDs = append(menuIDs, v.MenuID)
	}
	if len(menuIDs) == 0 {
		return nil, nil
	}
	return s.menuRepo.ListEnabledByIDs(ctx, nil, menuIDs)
}

// buildTree 用于获取菜单权限树
func (s *Service) buildTree(list []RoleMenuItem, parentID uint64) []RoleMenuItem {
	// 构建 parent -> children map
	childrenMap := make(map[uint64][]RoleMenuItem)
	for _, el := range list {
		childrenMap[el.ParentID] = append(childrenMap[el.ParentID], el)
	}
	// 递归
	var build func(pid uint64) []RoleMenuItem
	build = func(pid uint64) []RoleMenuItem {
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

// logout 用于退出管理员的登录
func (s *Service) logout(ctx context.Context, adminID uint64) (int, error) {
	// 清空用户token
	if s.rdb != nil {
		err := s.rdb.Del(ctx,
			jwt.AdminTokenKey(adminID),
			jwt.AdminRefreshKey(adminID),
		).Err()
		if err != nil {
			return 60107, err
		}
	}
	if err := s.adminRepo.UpdateTokenByID(ctx, nil, adminID, nil); err != nil {
		return 60104, err
	}
	// 返回数据
	return 0, nil
}
