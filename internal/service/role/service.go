package role

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/admin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/admin_role"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/menu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/role"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/role_menu"
	"gorm.io/gorm"
)

type Service struct {
	roleRepo      role.Repository
	adminRepo     admin.Repository
	roleMenuRepo  role_menu.Repository
	adminRoleRepo admin_role.Repository
	menuRepo      menu.Repository
	db            *gorm.DB
	rdb           *redis.Client
}

const RoleCodeSuperAdmin = "SUPER_ADMIN"

func New(roleRepo role.Repository, adminRepo admin.Repository, roleMenuRepo role_menu.Repository, adminRoleRepo admin_role.Repository, menuRepo menu.Repository, db *gorm.DB, rdb *redis.Client) *Service {
	return &Service{
		roleRepo:      roleRepo,
		adminRepo:     adminRepo,
		roleMenuRepo:  roleMenuRepo,
		adminRoleRepo: adminRoleRepo,
		menuRepo:      menuRepo,
		db:            db,
		rdb:           rdb,
	}
}

// ListPage 用于获取角色分页信息
func (s *Service) ListPage(ctx context.Context, req ListPageReq) (ListPageResp, int, error) {
	// 获取列表数据
	offset, limit := req.OffsetLimit()
	roles, total, err := s.roleRepo.ListPage(ctx, nil, model.RoleListPageQuery{
		Name:   req.Name,
		Enable: req.Enable,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return ListPageResp{}, 60201, err
	}
	// 获取菜单
	roleIDs := make([]uint64, 0, len(roles))
	for _, v := range roles {
		roleIDs = append(roleIDs, v.ID)
	}
	roleMenus, err := s.roleMenuRepo.ListByRoleIDs(ctx, nil, roleIDs)
	menuIDs := make(map[uint64][]uint64)
	for _, v := range roleMenus {
		menuIDs[v.RoleID] = append(menuIDs[v.RoleID], v.MenuID)
	}
	// 组装数据
	list := make([]ListPageItem, 0, len(roles))
	for _, v := range roles {
		list = append(list, ListPageItem{
			ID:            v.ID,
			Code:          v.Code,
			Name:          v.Name,
			Enable:        v.Enable == enum.EnableEnable,
			PermissionIds: menuIDs[v.ID],
		})
	}
	// 返回数据
	return ListPageResp{
		Total:    total,
		PageData: list,
	}, 0, nil
}

// ListAll 用于获取角色全部信息
func (s *Service) ListAll(ctx context.Context) ([]ListAllResp, int, error) {
	// 获取角色
	roles, err := s.roleRepo.FindAll(ctx, nil)
	if err != nil {
		return nil, 60201, err
	}
	// 组装数据
	list := make([]ListAllResp, 0, len(roles))
	for _, v := range roles {
		list = append(list, ListAllResp{
			ID:     v.ID,
			Code:   v.Code,
			Name:   v.Name,
			Enable: v.Enable == enum.EnableEnable,
		})
	}
	return list, 0, nil
}

// Save 用于创建或修改角色信息
func (s *Service) Save(ctx context.Context, req SaveReq) (int, error) {
	// 开启事务
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 变更角色信息
		var roleID uint64
		var err error
		isCreate := req.ID == nil || *req.ID == 0
		if isCreate {
			roleID, err = s.add(ctx, tx, req)
		} else {
			roleID, err = s.update(ctx, tx, req)
		}
		if err != nil {
			return err
		}
		// 重新绑定角色对应的权限
		if req.MenuIDs != nil {
			if err := s.resetRoleMenus(ctx, tx, roleID, *req.MenuIDs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 60202, err
	}
	return 0, nil
}

// Delete 用于删除角色信息
func (s *Service) Delete(ctx context.Context, roleID uint64) (int, error) {
	// 获取角色信息
	role, err := s.roleRepo.GetByID(ctx, nil, roleID)
	if err != nil {
		return 60201, err
	}
	if role == nil {
		return 50201, nil
	}
	// 开启事务，删除角色
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.roleMenuRepo.DeleteByRoleID(ctx, tx, roleID); err != nil {
			return err
		}
		if err := s.roleRepo.Delete(ctx, tx, roleID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 60203, err
	}
	// 返回数据
	return 0, nil
}

// RoleMenuTree 用于获取管理员权限内的菜单
func (s *Service) RoleMenuTree(ctx context.Context, roleID uint64, roleCode string) ([]RoleMenuItem, int, error) {
	// 获取菜单信息
	menus, err := s.getMenusByRole(ctx, roleID, roleCode)
	if err != nil {
		return nil, 60205, err
	}
	// 整理菜单信息
	list := make([]RoleMenuItem, 0, len(menus))
	for _, v := range menus {
		list = append(list, RoleMenuItem{
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
	// 返回权限树
	return s.buildTree(list, 0), 0, nil
}

// AddRoleUsers 用于分配角色到管理员
func (s *Service) AddRoleUsers(ctx context.Context, roleID uint64, adminIds []uint64) (int, error) {
	// 获取角色信息
	role, err := s.roleRepo.GetByID(ctx, nil, roleID)
	if err != nil {
		return 60201, err
	}
	if role == nil {
		return 50201, nil
	}
	if role.Enable == enum.EnableDisable {
		return 40202, nil
	}
	// 检查传入的管理员信息
	adminID := make([]uint64, 0, len(adminIds))
	for _, v := range adminIds {
		admin, err := s.adminRepo.GetByID(ctx, nil, v)
		if err != nil {
			continue
		}
		if admin == nil {
			continue
		}
		// 兜底: 如果管理员没有角色就分配角色
		if admin.RoleID == 0 {
			s.adminRepo.UpdateRoleIDByID(ctx, nil, v, roleID)
		}
		// 获取管理员是否拥有角色
		exists, err := s.adminRoleRepo.ExistsByAdminIDAndRoleID(ctx, nil, v, roleID)
		if err != nil {
			continue
		}
		if !exists {
			adminID = append(adminID, v)
			s.logout(ctx, v)
		}
	}
	// 批量为管理员添加角色
	err = s.adminRoleRepo.BindRoles(ctx, nil, adminID, roleID)
	if err != nil {
		return 60206, err
	}
	return 0, nil
}

// RemoveRoleUsers 用于取消分配角色到管理员
func (s *Service) RemoveRoleUsers(ctx context.Context, roleID uint64, adminIds []uint64) (int, error) {
	if len(adminIds) == 0 {
		return 0, nil
	}
	var logoutAdminIDs []uint64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 查所有管理员
		admins, err := s.adminRepo.GetByIDs(ctx, tx, adminIds)
		if err != nil {
			return fmt.Errorf("获取管理员失败: %w", err)
		}
		// 查所有角色关系（一次性查，避免 N+1）
		adminRoles, err := s.adminRoleRepo.ListByAdminIDs(ctx, tx, adminIds)
		if err != nil {
			return fmt.Errorf("获取多个管理员的所有角色失败: %w", err)
		}
		// 构建 map：adminID -> []roleID
		roleMap := make(map[uint64][]uint64, len(adminIds))
		for _, ar := range adminRoles {
			roleMap[ar.AdminID] = append(roleMap[ar.AdminID], ar.RoleID)
		}
		// 校验：是否会出现“无角色”
		for _, admin := range admins {
			roles := roleMap[admin.ID]
			// 统计删除后剩余角色数量
			remain := 0
			for _, r := range roles {
				if r != roleID {
					remain++
				}
			}
			if remain == 0 {
				return fmt.Errorf("admin %d only has this role, cannot remove", admin.ID)
			}
		}
		// 删除角色关系
		if err := s.adminRoleRepo.UnbindRoles(ctx, tx, adminIds, roleID); err != nil {
			return fmt.Errorf("删除管理员绑定角色失败: %w", err)
		}
		// 处理需要切换当前角色的管理员
		for _, admin := range admins {
			if admin.RoleID != roleID {
				continue
			}
			roles := roleMap[admin.ID]
			// 找一个新的 role（排除被删除的）
			var newRoleID uint64
			for _, r := range roles {
				if r != roleID {
					newRoleID = r
					break
				}
			}
			if err := s.adminRepo.UpdateRoleIDByID(ctx, tx, admin.ID, newRoleID); err != nil {
				return fmt.Errorf("更新管理员角色ID失败: %w", err)
			}
			logoutAdminIDs = append(logoutAdminIDs, admin.ID)
		}
		return nil
	})
	if err != nil {
		return 60206, err
	}
	// 事务外处理 Redis（避免 DB 回滚时 Redis 已清理）
	for _, id := range logoutAdminIDs {
		s.logout(ctx, id)
	}
	return 0, nil
}
