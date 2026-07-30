package admin

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/admin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/admin_role"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/role"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/crypto"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/jwt"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
	"gorm.io/gorm"
)

type Service struct {
	db            *gorm.DB
	adminRepo     admin.Repository
	adminRoleRepo admin_role.Repository
	roleRepo      role.Repository
	rdb           *redis.Client
}

func New(adminRepo admin.Repository, adminRoleRepo admin_role.Repository, roleRepo role.Repository, db *gorm.DB, rdb *redis.Client) *Service {
	return &Service{
		adminRepo:     adminRepo,
		adminRoleRepo: adminRoleRepo,
		roleRepo:      roleRepo,
		db:            db,
		rdb:           rdb,
	}
}

// Login 用于处理管理员登录
func (s *Service) Login(ctx context.Context, username, password string) (LoginResp, int, error) {
	// 获取管理员信息
	admin, err := s.adminRepo.GetByUsername(ctx, nil, username)
	if err != nil {
		return LoginResp{}, 60101, err
	}
	// 验证密码是否正确
	if admin == nil {
		return LoginResp{}, 50101, nil
	}
	if !crypto.CheckPassword(admin.Password, password) {
		return LoginResp{}, 40101, nil
	}
	// 验证账号是否启用
	if admin.Enable != enum.EnableEnable {
		return LoginResp{}, 40102, nil
	}
	// 获取角色code
	role, err := s.roleRepo.GetByID(ctx, nil, admin.RoleID)
	if err != nil {
		return LoginResp{}, 60101, err
	}
	if role == nil {
		return LoginResp{}, 50102, nil
	}
	// 更新token
	updateToken, errCode, err := s.updateToken(ctx, admin.ID, admin.RoleID, role.Code)
	if err != nil {
		return LoginResp{}, errCode, err
	}
	// 返回数据
	return LoginResp{
		AccessToken:  updateToken.AccessToken,
		RefreshToken: updateToken.RefreshToken,
	}, 0, nil
}

// RefreshLogin 用于刷新管理员登录状态
func (s *Service) RefreshLogin(ctx context.Context, refreshToken string) (RefreshLoginResp, int, error) {
	claims, err := jwt.ParseToken(refreshToken)
	if err != nil {
		return RefreshLoginResp{}, 10002, err
	}
	if claims.Type != "refresh" {
		return RefreshLoginResp{}, 10003, nil
	}
	// 获取管理员信息
	admin, err := s.adminRepo.GetByID(ctx, nil, claims.ID)
	if err != nil {
		return RefreshLoginResp{}, 60101, err
	}
	// 验证信息
	if admin == nil {
		return RefreshLoginResp{}, 50101, nil
	}
	if admin.Token == nil || *admin.Token != refreshToken {
		return RefreshLoginResp{}, 20001, nil
	}
	// 获取角色code
	role, err := s.roleRepo.GetByID(ctx, nil, admin.RoleID)
	if err != nil {
		return RefreshLoginResp{}, 60101, err
	}
	if role == nil {
		return RefreshLoginResp{}, 50102, nil
	}
	// 更新token
	updateToken, errCode, err := s.updateToken(ctx, claims.ID, admin.RoleID, role.Code)
	if err != nil {
		return RefreshLoginResp{}, errCode, err
	}
	// 返回数据
	return RefreshLoginResp{
		AccessToken:  updateToken.AccessToken,
		RefreshToken: updateToken.RefreshToken,
	}, 0, nil
}

// Logout 用于退出管理员登录
func (s *Service) Logout(ctx context.Context, adminID uint64) (int, error) {
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

// SwitchRole 用于切换管理员角色信息
func (s *Service) SwitchRole(ctx context.Context, adminID uint64, code string) (SwitchRoleResp, int, error) {
	// 获取角色信息
	role, err := s.roleRepo.GetByCode(ctx, nil, code)
	if err != nil {
		return SwitchRoleResp{}, 60101, err
	}
	if role == nil {
		return SwitchRoleResp{}, 50102, nil
	}
	if role.Enable != enum.EnableEnable {
		return SwitchRoleResp{}, 40103, nil
	}
	// 获取用户角色是否存在
	exists, err := s.adminRoleRepo.ExistsByAdminIDAndRoleID(ctx, nil, adminID, role.ID)
	if err != nil {
		return SwitchRoleResp{}, 60101, err
	}
	if !exists {
		return SwitchRoleResp{}, 30101, nil
	}
	// 变更角色
	if err := s.adminRepo.UpdateRoleIDByID(ctx, nil, adminID, role.ID); err != nil {
		return SwitchRoleResp{}, 60108, err
	}
	// 更新token
	updateToken, errCode, err := s.updateToken(ctx, adminID, role.ID, role.Code)
	if err != nil {
		return SwitchRoleResp{}, errCode, err
	}
	// 返回数据
	return SwitchRoleResp{
		AccessToken:  updateToken.AccessToken,
		RefreshToken: updateToken.RefreshToken,
	}, 0, nil
}

// ChangePassword 用于根据管理员旧密码修改密码
func (s *Service) ChangePassword(ctx context.Context, adminID uint64, oldPassword, newPassword string) (int, error) {
	// 获取管理员信息
	admin, err := s.adminRepo.GetByID(ctx, nil, adminID)
	if err != nil {
		return 60101, err
	}
	if admin == nil {
		return 50101, nil
	}
	// 验证旧密码
	if !crypto.CheckPassword(admin.Password, oldPassword) {
		return 40101, nil
	}
	// 防止新旧密码相同
	if crypto.CheckPassword(admin.Password, newPassword) {
		return 40104, nil
	}
	password, err := crypto.HashPassword(newPassword)
	if err != nil {
		return 60109, err
	}
	if err := s.adminRepo.UpdatePasswordByID(ctx, nil, adminID, password); err != nil {
		return 60110, err
	}
	// 清空管理员登录状态并返回
	return s.Logout(ctx, adminID)
}

// ListPage 用于获取管理员列表信息
func (s *Service) ListPage(ctx context.Context, req ListPageReq) (ListPageResp, int, error) {
	// 获取列表数据
	offset, limit := req.OffsetLimit()
	admins, total, err := s.adminRepo.ListPage(ctx, nil, model.AdminListPageQuery{
		Username: req.Username,
		Gender:   req.Gender,
		Enable:   req.Enable,
		Offset:   offset,
		Limit:    limit,
	})
	if err != nil {
		return ListPageResp{}, 60101, err
	}
	// 获取列表管理员角色
	adminIDs := make([]uint64, 0, len(admins))
	for _, v := range admins {
		adminIDs = append(adminIDs, v.ID)
	}
	// 获取角色映射
	adminRoleMap, errCode, err := s.getAdminRolesMapByIDs(ctx, adminIDs)
	if errCode > 0 {
		return ListPageResp{}, errCode, err
	}
	// 返回数据
	return ListPageResp{
		Total:    total,
		PageData: toListPageItems(admins, adminRoleMap),
	}, 0, nil
}

// Details 用于获取单个管理员的详细信息
func (s *Service) Details(ctx context.Context, adminID uint64) (DetailsResp, int, error) {
	// 获取管理员信息
	admin, err := s.adminRepo.GetByID(ctx, nil, adminID)
	if err != nil {
		return DetailsResp{}, 60101, err
	}
	if admin == nil {
		return DetailsResp{}, 50101, nil
	}
	// 获取所有启用的角色
	roles, err := s.roleRepo.ListEnabled(ctx, nil)
	if err != nil {
		return DetailsResp{}, 60101, err
	}
	// 组装参数
	roleList := make([]RoleItem, 0, len(roles))
	var currentRole RoleItem
	for _, v := range roles {
		item := RoleItem{
			ID:     v.ID,
			Code:   v.Code,
			Name:   v.Name,
			Enable: v.Enable == enum.EnableEnable,
		}
		roleList = append(roleList, item)
		if v.ID == admin.RoleID {
			currentRole = item
		}
	}
	// 返回数据
	return DetailsResp{
		ID:        admin.ID,
		Username:  admin.Username,
		Enable:    admin.Enable == enum.EnableEnable,
		CreatedAt: timeutil.Format(admin.CreatedAt),
		UpdatedAt: timeutil.Format(admin.UpdatedAt),
		Profile: DetailsProfileItem{
			ID:       admin.ID,
			Nickname: admin.Nickname,
			Gender:   int(admin.Gender),
			Avatar:   admin.Avatar,
			Address:  ptr.Deref(admin.Address),
			Email:    ptr.Deref(admin.Email),
		},
		Roles:       roleList,
		CurrentRole: currentRole,
	}, 0, nil
}

// Save 用于创建或修改管理员信息
func (s *Service) Save(ctx context.Context, req SaveReq) (int, error) {
	// 在事务中创建/更新管理员并绑定角色
	var adminID uint64
	var errCode int
	var err error
	isCreate := req.ID == nil || *req.ID == 0

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if isCreate {
			adminID, errCode, err = s.add(ctx, tx, req)
		} else {
			adminID, errCode, err = s.update(ctx, tx, req)
		}
		if errCode > 0 {
			return err
		}
		// 绑定角色身份（在同一个事务中）
		errCode, err = s.bindRoles(ctx, tx, adminID, req.RoleIds)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errCode, err
	}
	// 退出对应的管理员（事务外，避免 Redis 与 DB 不一致）
	return s.Logout(ctx, adminID)
}

// Delete 用于删除管理员信息
func (s *Service) Delete(ctx context.Context, adminID uint64) (int, error) {
	// 获取管理员信息
	admin, err := s.adminRepo.GetByID(ctx, nil, adminID)
	if err != nil {
		return 60101, err
	}
	if admin == nil {
		return 50101, nil
	}
	// 踢出登录
	errCode, err := s.Logout(ctx, adminID)
	if errCode > 0 {
		return errCode, err
	}
	// 开启事务，删除管理员
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.adminRoleRepo.DeleteByAdminID(ctx, tx, adminID); err != nil {
			return err
		}
		if err := s.adminRepo.Delete(ctx, tx, adminID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 60112, err
	}
	// 返回数据
	return 0, nil
}

// ResetAdminPassword 修改管理员密码
func (s *Service) ResetAdminPassword(ctx context.Context, adminID uint64, newPassword string) (int, error) {
	// 验证管理员是否存在
	admin, err := s.adminRepo.GetByID(ctx, nil, adminID)
	if err != nil {
		return 60101, err
	}
	if admin == nil {
		return 50101, nil
	}
	password, err := crypto.HashPassword(newPassword)
	if err != nil {
		return 60109, err
	}
	if err := s.adminRepo.UpdatePasswordByID(ctx, nil, adminID, password); err != nil {
		return 60110, err
	}
	// 清空管理员登录状态并返回
	return s.Logout(ctx, adminID)
}

// UpdateProfile 修改管理员个人信息
func (s *Service) UpdateProfile(ctx context.Context, req UpdateProfileReq) (int, error) {
	// 验证管理员是否存在
	admin, err := s.adminRepo.GetByID(ctx, nil, req.ID)
	if err != nil {
		return 60101, err
	}
	if admin == nil {
		return 50101, nil
	}
	// 变更管理员信息
	if err := s.adminRepo.UpdateProfileByID(ctx, nil, req.ID, model.AdminUpdateProfileByIdForm{
		Nickname: req.Nickname,
		Email:    req.Email,
		Address:  req.Address,
		Gender:   req.Gender,
		Avatar:   req.Avatar,
	}); err != nil {
		return 60113, err
	}
	return 0, nil
}
