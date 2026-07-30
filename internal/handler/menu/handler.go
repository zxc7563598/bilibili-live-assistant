package menu

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/GoAdminKit/internal/dto/input"
	"github.com/zxc7563598/GoAdminKit/internal/dto/resp"
	"github.com/zxc7563598/GoAdminKit/internal/handler"
	"github.com/zxc7563598/GoAdminKit/internal/i18n"
	"github.com/zxc7563598/GoAdminKit/internal/logger"
	"github.com/zxc7563598/GoAdminKit/internal/response"
	"github.com/zxc7563598/GoAdminKit/internal/service/menu"
	"go.uber.org/zap"
)

type Handler struct {
	menuSvc *menu.Service
}

func New(menuSvc *menu.Service) *Handler {
	return &Handler{
		menuSvc: menuSvc,
	}
}

// @Summary 获取菜单列表
// @Description 获取系统完整菜单树结构（包含菜单及按钮权限），用于前端动态路由生成和权限控制
// @Tags 菜单
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Success 200 {object} response.Response{data=resp.MenuListResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/menu/list [post]
func (h *Handler) List(c *gin.Context) {
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
	svcResp, errCode, err := h.menuSvc.MenuTree(ctx)
	if errCode != 0 {
		handler.ErrorLog(
			logger.MenuLogger,
			"menuSvc.MenuTree 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.MenuListResp{
		Menu: toMenuItem(svcResp),
	})
}

// @Summary 校验菜单路径是否存在
// @Description 根据菜单路径校验菜单是否已存在，常用于前端需要对菜单进行校验时
// @Tags 菜单
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.MenuValidateReq true "校验参数（菜单路径）"
// @Success 200 {object} response.Response{data=resp.MenuValidateResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/menu/validate [post]
func (h *Handler) Validate(c *gin.Context) {
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
	var req input.MenuValidateReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.MenuLogger,
			"Validate 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.menuSvc.MenuExists(ctx, req.Path)
	if errCode != 0 {
		handler.ErrorLog(
			logger.MenuLogger,
			"menuSvc.MenuExists 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.String("req.path", req.Path),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.MenuValidateResp{
		Has: svcResp,
	})
}

// @Summary 获取菜单下的按钮权限
// @Description 根据菜单 ID 获取其下的按钮权限列表（按钮类型菜单），用于前端页面按钮级权限控制
// @Tags 菜单
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.MenuButtonsReq true "查询参数（父级菜单 ID）"
// @Success 200 {object} response.Response{data=resp.MenuButtonsResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/menu/buttons [post]
func (h *Handler) Buttons(c *gin.Context) {
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
	var req input.MenuButtonsReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.MenuLogger,
			"Buttons 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.menuSvc.MenuButtons(ctx, req.ParentID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.MenuLogger,
			"menuSvc.MenuButtons 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Uint64("req.parent_id", req.ParentID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.MenuButtonsResp{
		Total:    int64(len(svcResp)),
		PageData: toMenuItem(svcResp),
	})
}

// @Summary 创建或更新菜单
// @Description 根据是否传入 ID 创建或更新菜单信息：未传 ID 时为创建，传入 ID 时为更新，支持菜单及按钮类型（MENU / BUTTON），用于构建系统菜单结构和权限控制
// @Tags 菜单
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.MenuSaveReq true "菜单保存参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/menu/save [post]
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
	var req input.MenuSaveReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.MenuLogger,
			"Save 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.menuSvc.Save(ctx, menu.SaveReq{
		ID:        req.ID,
		Code:      req.Code,
		Enable:    req.Enable,
		Show:      req.Show,
		KeepAlive: req.KeepAlive,
		Layout:    req.Layout,
		Type:      req.Type,
		ParentID:  req.ParentID,
		Name:      req.Name,
		Icon:      req.Icon,
		Path:      req.Path,
		Component: req.Component,
		Order:     req.Order,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.MenuLogger,
			"menuSvc.Save 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Any("req.id", req.ID),
			zap.String("req.code", req.Code),
			zap.Bool("req.enable", req.Enable),
			zap.Bool("req.show", req.Show),
			zap.Bool("req.keep_alive", req.KeepAlive),
			zap.String("req.layout", req.Layout),
			zap.String("req.type", req.Type),
			zap.Uint64("req.parent_id", req.ParentID),
			zap.String("req.name", req.Name),
			zap.String("req.icon", req.Icon),
			zap.String("req.path", req.Path),
			zap.String("req.component", req.Component),
			zap.Int("req.order", req.Order),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 切换菜单启用状态
// @Description 根据菜单 ID 切换菜单的启用状态（启用/禁用），常用于后台列表中的快速开关操作
// @Tags 菜单
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.MenuToggleReq true "切换参数（菜单 ID）"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/menu/toggle [post]
func (h *Handler) Toggle(c *gin.Context) {
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
	var req input.MenuToggleReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.MenuLogger,
			"Toggle 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.menuSvc.SetMenuEnable(ctx, req.ID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.MenuLogger,
			"menuSvc.SetMenuEnable 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Uint64("req.id", req.ID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 删除菜单
// @Description 根据菜单 ID 删除菜单（通常不允许删除存在子菜单或已绑定权限的菜单）
// @Tags 菜单
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.MenuDeleteReq true "删除参数（菜单 ID）"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/menu/delete [post]
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
	var req input.MenuDeleteReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(
			logger.MenuLogger,
			"Delete 参数异常",
			code,
			err,
		)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.menuSvc.Delete(ctx, req.ID)
	if errCode != 0 {
		handler.ErrorLog(
			logger.MenuLogger,
			"menuSvc.Delete 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Uint64("req.id", req.ID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}
