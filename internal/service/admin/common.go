package admin

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/crypto"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/jwt"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
	"gorm.io/gorm"
)

// updateToken 用于更新管理员token
func (s *Service) updateToken(ctx context.Context, adminID, roleID int64, roleCode string) (TokenResp, int, error) {
	accessToken, err := jwt.GenerateAccessToken(adminID, "admin", roleID, roleCode)
	if err != nil {
		return TokenResp{}, 60102, err
	}
	newRefreshToken, err := jwt.GenerateRefreshToken(adminID, "admin", roleID, roleCode)
	if err != nil {
		return TokenResp{}, 60103, err
	}
	if err := s.adminRepo.UpdateTokenByID(ctx, nil, adminID, &newRefreshToken); err != nil {
		return TokenResp{}, 60104, err
	}
	if s.rdb != nil {
		if err := s.rdb.Set(ctx, jwt.AdminTokenKey(adminID), accessToken, jwt.AccessTTL()).Err(); err != nil {
			return TokenResp{}, 60105, err
		}
		if err := s.rdb.Set(ctx, jwt.AdminRefreshKey(adminID), newRefreshToken, jwt.RefreshTTL()).Err(); err != nil {
			return TokenResp{}, 60106, err
		}
	}
	return TokenResp{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, 0, nil
}

// getAdminRolesMap 用于获取 管理员 → 角色 映射列表
func (s *Service) getAdminRolesMapByIDs(ctx context.Context, adminIDs []int64) (map[int64][]RoleItem, int, error) {
	// 获取全部角色
	roles, err := s.roleRepo.FindAll(ctx, nil)
	if err != nil {
		return nil, 60101, err
	}
	// 组装角色数据
	roleMap := make(map[int64]RoleItem)
	for _, v := range roles {
		roleMap[v.ID] = RoleItem{
			ID:     v.ID,
			Code:   v.Code,
			Name:   v.Name,
			Enable: v.Enable == enum.EnableEnable,
		}
	}
	// 获取全部管理员对应角色
	adminRoles, err := s.adminRoleRepo.ListByAdminIDs(ctx, nil, adminIDs)
	if err != nil {
		return nil, 60101, err
	}
	// 组装管理员对应角色数据
	adminRoleMap := make(map[int64][]RoleItem)
	for _, ar := range adminRoles {
		role, ok := roleMap[ar.RoleID]
		if !ok {
			continue
		}
		adminRoleMap[ar.AdminID] = append(adminRoleMap[ar.AdminID], role)
	}
	// 返回数据
	return adminRoleMap, 0, nil
}

// add 用于创建管理员基本信息
func (s *Service) add(ctx context.Context, tx *gorm.DB, req SaveReq) (int64, int, error) {
	if v := req.Username; v == nil || *v == "" {
		return 0, 10101, nil
	}
	if req.Enable == nil {
		return 0, 10101, nil
	}
	if v := req.Password; v == nil || *v == "" {
		return 0, 10103, nil
	}
	if len(*req.Password) < 6 {
		return 0, 10104, nil
	}
	if len(*req.Password) > 32 {
		return 0, 10105, nil
	}
	if len(req.RoleIds) == 0 {
		return 0, 10108, nil
	}
	// 添加数据
	password, err := crypto.HashPassword(*req.Password)
	if err != nil {
		return 0, 60109, err
	}
	admin, err := s.adminRepo.Create(ctx, tx, &model.Admin{
		Nickname: *req.Username,
		Username: *req.Username,
		Password: password,
		RoleID:   req.RoleIds[0],
		Gender:   enum.GenderUnknown,
		Enable:   *enum.BoolToEnablePtr(req.Enable),
	})
	if err != nil {
		return 0, 60111, err
	}
	return admin.ID, 0, nil
}

// update 用于更新管理员基本信息
func (s *Service) update(ctx context.Context, tx *gorm.DB, req SaveReq) (int64, int, error) {
	var roleID *int64
	if len(req.RoleIds) > 0 {
		roleID = &req.RoleIds[0]
	}
	// 变更数据
	if err := s.adminRepo.UpdateBasicInfoByID(ctx, tx, *req.ID, model.AdminUpdateBasicInfoByIdForm{
		Username: req.Username,
		Enable:   enum.BoolToEnablePtr(req.Enable),
		RoleID:   roleID,
	}); err != nil {
		return 0, 60111, err
	}
	return *req.ID, 0, nil
}

// bindRoles 重置管理员的角色绑定关系
//
// 先清空原有角色，再批量绑定新的角色。数据库操作在事务中进行。
// 仅影响 admin_role 表，不会修改 admin.RoleID。
func (s *Service) bindRoles(ctx context.Context, tx *gorm.DB, adminID int64, roleIds []int64) (int, error) {
	if len(roleIds) == 0 {
		return 0, nil
	}
	// 删除原有角色信息
	if err := s.adminRoleRepo.DeleteByAdminID(ctx, tx, adminID); err != nil {
		return 60114, err
	}
	// 重新绑定角色信息
	adminRoleList := make([]model.AdminRole, 0, len(roleIds))
	for _, v := range roleIds {
		adminRoleList = append(adminRoleList, model.AdminRole{
			AdminID: adminID,
			RoleID:  v,
		})
	}
	if err := s.adminRoleRepo.CreateBatch(ctx, tx, adminRoleList); err != nil {
		return 60115, err
	}
	return 0, nil
}

func toListPageItems(admins []model.AdminListItem, adminRoleMap map[int64][]RoleItem) []ListPageItem {
	respList := make([]ListPageItem, 0, len(admins))
	for _, v := range admins {
		item := ListPageItem{
			ID:        v.ID,
			Username:  v.Username,
			Enable:    v.Enable == enum.EnableEnable,
			Gender:    int(v.Gender),
			Avatar:    v.Avatar,
			Address:   ptr.Deref(v.Address),
			Email:     ptr.Deref(v.Email),
			Roles:     adminRoleMap[v.ID],
			CreatedAt: timeutil.Format(v.CreatedAt),
			UpdatedAt: timeutil.Format(v.UpdatedAt),
		}
		respList = append(respList, item)
	}
	return respList
}
