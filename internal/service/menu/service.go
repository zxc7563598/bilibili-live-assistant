package menu

import (
	"context"
	"errors"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/menu"
)

type Service struct {
	menuRepo menu.Repository
}

func New(menuRepo menu.Repository) *Service {
	return &Service{
		menuRepo: menuRepo,
	}
}

// GetMenuTree 用于获取全部菜单
func (s *Service) GetMenuTree(ctx context.Context) ([]MenuItem, int, error) {
	// 获取菜单信息
	menus, err := s.menuRepo.ListAll(ctx, nil)
	if err != nil {
		return nil, CodeQueryFailed, err
	}
	// 返回权限树
	return s.buildTree(toMenuItems(menus), 0), 0, nil
}

// MenuExists 用于获取菜单是否存在
func (s *Service) MenuExists(ctx context.Context, path string) (bool, int, error) {
	has, err := s.menuRepo.ExistsByPath(ctx, nil, path)
	if err != nil {
		return false, CodeQueryFailed, err
	}
	return has, 0, nil
}

// ListMenuButtons 用于获取菜单下的按钮
func (s *Service) ListMenuButtons(ctx context.Context, parentID int64) ([]MenuItem, int, error) {
	buttons, err := s.menuRepo.ListButtonsByParentID(ctx, nil, parentID)
	if err != nil {
		return nil, CodeQueryFailed, err
	}
	return toMenuItems(buttons), 0, nil
}

// Save 用于添加或变更数据
func (s *Service) Save(ctx context.Context, req SaveReq) (int, error) {
	t := enum.MenuType(req.Type)
	if !t.IsValid() {
		return CodeParamInvalid, errors.New("type 类型异常")
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

// ToggleMenuEnable 用于切换菜单启动状态
func (s *Service) ToggleMenuEnable(ctx context.Context, id int64) (int, error) {
	// 翻转是 UPDATE ... WHERE id = ? 形式，命中 0 行不报错，
	// 不先查一次的话对不存在的 ID 会返回成功
	if errCode, err := s.checkMenuExists(ctx, id); errCode != 0 {
		return errCode, err
	}
	err := s.menuRepo.ToggleEnableByID(ctx, nil, id)
	if err != nil {
		return CodeToggleEnableFailed, err
	}
	return 0, nil
}

// Delete 用于删除菜单
func (s *Service) Delete(ctx context.Context, id int64) (int, error) {
	// 同上：软删除命中 0 行也不报错，需先确认菜单存在
	if errCode, err := s.checkMenuExists(ctx, id); errCode != 0 {
		return errCode, err
	}
	err := s.menuRepo.Delete(ctx, nil, id)
	if err != nil {
		return CodeDeleteFailed, err
	}
	return 0, nil
}

// checkMenuExists 确认菜单存在，存在返回 (0, nil)
func (s *Service) checkMenuExists(ctx context.Context, id int64) (int, error) {
	menu, err := s.menuRepo.GetByID(ctx, nil, id)
	if err != nil {
		return CodeQueryFailed, err
	}
	if menu == nil {
		return CodeNotFound, nil
	}
	return 0, nil
}
