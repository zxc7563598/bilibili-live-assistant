package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/admin"
	altchaSvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/altcha"
	"go.uber.org/zap"
)

type Handler struct {
	adminSvc  *admin.Service
	altchaSvc *altchaSvc.Service
}

func New(adminSvc *admin.Service, altchaSvc *altchaSvc.Service) *Handler {
	return &Handler{
		adminSvc:  adminSvc,
		altchaSvc: altchaSvc,
	}
}

// @Summary 管理员登录
// @Description 管理员通过账号和密码登录系统，登录成功后返回 access_token 和 refresh_token，用于后续接口鉴权
// @Tags 认证
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AdminLoginReq true "登录参数"
// @Success 200 {object} response.Response{data=resp.AdminLoginResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取请求参数
	var req input.AdminLoginReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.AdminLogger,
			"Login 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 验证码校验
	if errCode, err := h.altchaSvc.VerifySolution(ctx, req.Captcha); errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"altchaSvc.VerifySolution 调用失败",
			errCode,
			err,
		)
		response.Error(c, lang, errCode)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.adminSvc.Login(ctx, req.Username, req.Password)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"adminSvc.Login 调用失败",
			errCode,
			err,
			zap.String("uname", req.Username),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回数据
	response.Success(c, lang, resp.AdminLoginResp{
		AccessToken:  svcResp.AccessToken,
		RefreshToken: svcResp.RefreshToken,
	})
}

// @Summary 获取验证码启用状态
// @Description 获取当前系统是否已配置并启用了 altcha 验证码，前端可根据此状态决定是否展示验证码组件
// @Tags 认证
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.AdminCaptchaResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/auth/captcha [post]
func (h *Handler) CaptchaStatus(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	response.Success(c, lang, resp.AdminCaptchaResp{Enabled: h.altchaSvc.IsEnabled()})
}

// @Summary 刷新登录凭证
// @Description 使用 refresh_token 刷新登录状态，获取新的 access_token 和 refresh_token，用于延长会话有效期
// @Tags 认证
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AdminRefreshReq true "刷新凭证参数"
// @Success 200 {object} response.Response{data=resp.AdminLoginResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	// 获取上下文/语言配置
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取请求参数
	var req input.AdminRefreshReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.AdminLogger,
			"Refresh 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.adminSvc.RefreshLogin(ctx, req.Token)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"adminSvc.RefreshLogin 调用失败",
			errCode,
			err,
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.AdminLoginResp{
		AccessToken:  svcResp.AccessToken,
		RefreshToken: svcResp.RefreshToken,
	})
}

// @Summary 管理员退出登录
// @Description 管理员退出当前登录状态，使当前 access_token 失效（通常用于前端主动登出）
// @Tags 认证
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
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
	errCode, err := h.adminSvc.Logout(ctx, adminInfo.AdminID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"adminSvc.Logout 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 切换管理员角色
// @Description 管理员在已登录状态下切换当前角色，根据目标角色重新生成 access_token 和 refresh_token，用于更新权限上下文
// @Tags 认证
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AdminSwitchRoleReq true "切换角色参数"
// @Success 200 {object} response.Response{data=resp.AdminLoginResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/auth/switch-role [post]
func (h *Handler) SwitchRole(c *gin.Context) {
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
	var req input.AdminSwitchRoleReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.AdminLogger,
			"SwitchRole 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.adminSvc.SwitchRole(ctx, adminInfo.AdminID, req.Code)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"adminSvc.SwitchRole 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.String("req.code", req.Code),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.AdminLoginResp{
		AccessToken:  svcResp.AccessToken,
		RefreshToken: svcResp.RefreshToken,
	})
}

// @Summary 修改登录密码
// @Description 管理员在已登录状态下通过旧密码验证后修改登录密码，修改成功后建议重新登录以获取新的凭证
// @Tags 认证
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AdminChangePasswordReq true "修改密码参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/auth/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
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
	var req input.AdminChangePasswordReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.AdminLogger,
			"ChangePassword 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.adminSvc.ChangePassword(ctx, adminInfo.AdminID, req.OldPassword, req.NewPassword)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"adminSvc.ChangePassword 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 获取当前管理员信息
// @Description 获取当前登录管理员的详细信息，包括基础信息、个人资料、拥有角色列表及当前角色信息
// @Tags 认证
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.AdminDetailsResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/auth/detail [post]
func (h *Handler) Details(c *gin.Context) {
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
	svcResp, errCode, err := h.adminSvc.Details(ctx, adminInfo.AdminID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"adminSvc.Details 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.AdminDetailsResp{
		ID:        svcResp.ID,
		Username:  svcResp.Username,
		Enable:    svcResp.Enable,
		CreatedAt: svcResp.CreatedAt,
		UpdatedAt: svcResp.UpdatedAt,
		Profile: resp.AdminDetailsProfileItem{
			ID:       svcResp.Profile.ID,
			Nickname: svcResp.Profile.Nickname,
			Gender:   svcResp.Profile.Gender,
			Avatar:   svcResp.Profile.Avatar,
			Address:  svcResp.Profile.Address,
			Email:    svcResp.Profile.Email,
		},
		Roles: toAdminDetailsRoleItem(svcResp.Roles),
		CurrentRole: resp.AdminDetailsRoleItem{
			ID:     svcResp.CurrentRole.ID,
			Code:   svcResp.CurrentRole.Code,
			Name:   svcResp.CurrentRole.Name,
			Enable: svcResp.CurrentRole.Enable,
		},
	})
}

// @Summary 分页查询管理员列表
// @Description 分页获取管理员列表，支持按用户名、性别、启用状态进行筛选
// @Tags 管理员
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AdminListPageReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.AdminListPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/admin/list [post]
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
	var req input.AdminListPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.AdminLogger,
			"ListPage 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.adminSvc.ListPage(ctx, admin.ListPageReq{
		PageResp: admin.PageResp{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
		},
		Username: req.Username,
		Gender:   req.Gender,
		Enable:   req.Enable,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"adminSvc.ListPage 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int("req.pageNo", req.PageNo),
			zap.Int("req.pageSize", req.PageSize),
			zap.Any("req.username", req.Username),
			zap.Any("req.gender", req.Gender),
			zap.Any("req.enable", req.Enable),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.AdminListPageResp{
		Total:    svcResp.Total,
		PageData: toAdminListItems(svcResp.PageData),
	})
}

// @Summary 创建或更新管理员
// @Description 根据是否传入 ID 创建或更新管理员账号：未传 ID 时为创建，传入 ID 时为更新，同时可设置账号状态及绑定角色
// @Tags 管理员
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AdminSaveReq true "管理员保存参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/admin/save [post]
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
	var req input.AdminSaveReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.AdminLogger,
			"Save 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.adminSvc.Save(ctx, admin.SaveReq{
		ID:       req.ID,
		Enable:   req.Enable,
		Username: req.Username,
		Password: req.Password,
		RoleIds:  req.RoleIds,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"adminSvc.Save 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Any("req.id", req.ID),
			zap.Any("req.username", req.Username),
			zap.Any("req.role_ids", req.RoleIds),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 删除管理员
// @Description 根据管理员 ID 删除管理员账号（通常不允许删除当前登录管理员）
// @Tags 管理员
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AdminDeleteReq true "删除参数（管理员 ID）"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/admin/delete [post]
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
	var req input.AdminDeleteReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.AdminLogger,
			"Delete 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.adminSvc.Delete(ctx, req.ID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"adminSvc.Delete 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 重置管理员密码
// @Description 管理员根据管理员 ID 重置指定账号的登录密码（无需旧密码，通常需要具备管理员管理权限）
// @Tags 管理员
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AdminResetAdminPasswordReq true "重置密码参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/admin/update-password [post]
func (h *Handler) UpdatePassword(c *gin.Context) {
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
	var req input.AdminResetAdminPasswordReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.AdminLogger,
			"UpdatePassword 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.adminSvc.ResetAdminPassword(ctx, req.ID, req.Password)
	if errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"adminSvc.ResetAdminPassword 调用失败",
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

// @Summary 更新管理员个人资料
// @Description 管理员更新个人资料信息（如昵称、头像、性别、邮箱等），支持部分字段更新
// @Tags 管理员
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.AdminUpdateProfileReq true "管理员个人资料参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/admin/update-profile [post]
func (h *Handler) UpdateProfile(c *gin.Context) {
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
	var req input.AdminUpdateProfileReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.AdminLogger,
			"UpdateProfile 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.adminSvc.UpdateProfile(ctx, admin.UpdateProfileReq{
		ID:       req.ID,
		Nickname: req.Nickname,
		Gender:   req.Gender,
		Address:  req.Address,
		Email:    req.Email,
		Avatar:   req.Avatar,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.AdminLogger,
			"adminSvc.UpdateProfile 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.id", req.ID),
			zap.Any("req.nickname", req.Nickname),
			zap.Any("req.gender", req.Gender),
			zap.Any("req.Address", req.Address),
			zap.Any("req.Email", req.Email),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}
