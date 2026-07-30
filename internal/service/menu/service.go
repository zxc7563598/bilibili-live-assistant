package menu

import (
	"context"
	"errors"

	"github.com/zxc7563598/GoAdminKit/internal/enum"
	"github.com/zxc7563598/GoAdminKit/internal/repository/menu"
)

type Service struct {
	menuRepo menu.Repository
}

func New(menuRepo menu.Repository) *Service {
	return &Service{
		menuRepo: menuRepo,
	}
}

// MenuTree 用于获取全部菜单
func (s *Service) MenuTree(ctx context.Context) ([]MenuItem, int, error) {
	// 获取菜单信息
	menus, err := s.menuRepo.FindAll(ctx, nil)
	if err != nil {
		return nil, 60301, err
	}
	// 返回权限树
	return s.buildTree(toMenuItems(menus), 0), 0, nil
}

// MenuExists 用于获取菜单是否存在
func (s *Service) MenuExists(ctx context.Context, path string) (bool, int, error) {
	has, err := s.menuRepo.ExistsByPath(ctx, nil, path)
	if err != nil {
		return false, 60301, err
	}
	return has, 0, nil
}

// MenuButtons 用于获取菜单下的按钮
func (s *Service) MenuButtons(ctx context.Context, parentID uint64) ([]MenuItem, int, error) {
	buttons, err := s.menuRepo.ListButtonsByParentID(ctx, nil, parentID)
	if err != nil {
		return nil, 60301, err
	}
	return toMenuItems(buttons), 0, nil
}

// Save 用于添加或变更数据
func (s *Service) Save(ctx context.Context, req SaveReq) (int, error) {
	t := enum.MenuType(req.Type)
	if !t.IsValid() {
		return 10301, errors.New("type 类型异常")
	}
	isCreate := req.ID == nil || *req.ID == 0
	if isCreate {
		// 添加数据
		errCode, err := s.add(ctx, req)
		if errCode > 0 {
			return errCode, err
		}
	} else {
		// 变更数据
		errCode, err := s.update(ctx, req)
		if errCode > 0 {
			return errCode, err
		}
	}
	// 返回成功
	return 0, nil
}

// SetMenuEnable 用于切换菜单启动状态
func (s *Service) SetMenuEnable(ctx context.Context, id uint64) (int, error) {
	err := s.menuRepo.UpdateEnableByID(ctx, nil, id)
	if err != nil {
		return 60304, err
	}
	return 0, nil
}

// Delete 用于删除菜单
func (s *Service) Delete(ctx context.Context, id uint64) (int, error) {
	err := s.menuRepo.Delete(ctx, nil, id)
	if err != nil {
		return 60305, err
	}
	return 0, nil
}
