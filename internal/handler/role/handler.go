package role

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/role"
	"go.uber.org/zap"
)

type Handler struct {
	roleSvc *role.Service
}

func New(roleSvc *role.Service) *Handler {
	return &Handler{
		roleSvc: roleSvc,
	}
}

// @Summary 获取角色列表（分页）
// @Description 获取角色分页列表，支持按角色名称模糊搜索及状态筛选，用于后台角色管理与权限分配
// @Tags 角色
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.RoleListPageReq true "分页查询参数"
// @Success 200 {object} response.Response{data=resp.RoleListPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/roles/list [post]
func (h *Handler) ListPage(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取请求参数
	var req input.RoleListPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.RoleLogger,
			"ListPage 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.roleSvc.ListPage(ctx, role.ListPageReq{
		PageResp: role.PageResp{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
		},
		Name:   req.Name,
		Enable: req.Enable,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.RoleLogger,
			"roleSvc.ListPage 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int("req.pageNo", req.PageNo),
			zap.Int("req.pageSize", req.PageSize),
			zap.Any("req.name", req.Name),
			zap.Any("req.enable", req.Enable),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.RoleListPageResp{
		Total:    svcResp.Total,
		PageData: toRoleListItems(svcResp.PageData),
	})
}

// @Summary 获取全部角色列表
// @Description 获取系统所有角色的简要信息（不分页），用于下拉选择或管理员绑定角色
// @Tags 角色
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.RoleListAllResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/roles/all [post]
func (h *Handler) ListAll(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.roleSvc.ListAll(ctx)
	if errCode != 0 {
		handler.ErrorLog(
			logger.RoleLogger,
			"roleSvc.ListAll 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.RoleListAllResp{
		List: toRoleListAllItems(svcResp),
	})
}

// @Summary 创建或更新角色
// @Description 根据是否传入 ID 创建或更新角色信息，同时可绑定菜单及按钮权限（权限ID集合）
// @Tags 角色
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.RoleSaveReq true "角色保存参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/roles/save [post]
func (h *Handler) Save(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取请求参数
	var req input.RoleSaveReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.RoleLogger,
			"Save 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.roleSvc.Save(ctx, role.SaveReq{
		ID:      req.ID,
		Code:    req.Code,
		Name:    req.Name,
		Enable:  req.Enable,
		MenuIDs: req.PermissionIds,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.RoleLogger,
			"roleSvc.Save 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Any("req.id", req.ID),
			zap.Any("req.code", req.Code),
			zap.Any("req.name", req.Name),
			zap.Any("req.enable", req.Enable),
			zap.Any("req.menu_ids", req.PermissionIds),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 删除角色
// @Description 根据角色 ID 删除角色（通常不允许删除已绑定管理员或正在使用的角色）
// @Tags 角色
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.RoleDeleteReq true "删除参数（角色 ID）"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/roles/delete [post]
func (h *Handler) Delete(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取请求参数
	var req input.RoleDeleteReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.RoleLogger,
			"Delete 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.roleSvc.Delete(ctx, req.ID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.RoleLogger,
			"roleSvc.Delete 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.id", req.ID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 为角色分配管理员
// @Description 将指定角色批量分配给管理员（建立管理员与角色的关联关系）
// @Tags 角色
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.RoleAddRoleUsersReq true "分配参数（角色ID + 管理员ID列表）"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/roles/add-role-users [post]
func (h *Handler) AddRoleUsers(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取请求参数
	var req input.RoleAddRoleUsersReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.RoleLogger,
			"AddRoleUsers 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.roleSvc.AddRoleUsers(ctx, req.RoleID, req.AdminIds)
	if errCode != 0 {
		handler.ErrorLog(
			logger.RoleLogger,
			"roleSvc.AddRoleUsers 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.roleId", req.RoleID),
			zap.Any("req.adminIds", req.AdminIds),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 从角色移除管理员
// @Description 将指定管理员从角色中批量移除（解除管理员与角色的关联关系）
// @Tags 角色
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.RoleRemoveRoleUsersReq true "移除参数（角色ID + 管理员ID列表）"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/roles/remove-role-users [post]
func (h *Handler) RemoveRoleUsers(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取请求参数
	var req input.RoleRemoveRoleUsersReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.RoleLogger,
			"RemoveRoleUsers 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.roleSvc.RemoveRoleUsers(ctx, req.RoleID, req.AdminIds)
	if errCode != 0 {
		handler.ErrorLog(
			logger.RoleLogger,
			"roleSvc.RemoveRoleUsers 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.roleId", req.RoleID),
			zap.Any("req.adminIds", req.AdminIds),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 获取当前管理员权限菜单树
// @Description 获取当前登录管理员在当前角色下的菜单及按钮权限树（已按权限过滤），用于前端动态路由及权限控制
// @Tags 角色
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.RolePermissionsResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/roles/permissions [post]
func (h *Handler) Permissions(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取角色权限内的菜单
	menus, errCode, err := h.roleSvc.RoleMenuTree(ctx, adminInfo.RoleID, adminInfo.RoleCode)
	if errCode != 0 {
		handler.ErrorLog(
			logger.RoleLogger,
			"roleSvc.RoleMenuTree 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.RolePermissionsResp{
		Menu: toRoleMenuItem(menus),
	})
}
