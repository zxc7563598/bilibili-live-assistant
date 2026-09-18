package product

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/product"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/pagination"
	"go.uber.org/zap"
)

// @Summary 后台获取主页商品分页列表
// @Description 用于后台商品列表的展示
// @Tags 商品管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.ProductAdminListPageReq true "分页参数"
// @Success 200 {object} response.Response{data=resp.ProductListPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/product/list [post]
func (h *Handler) AdminListPage(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.ProductAdminListPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.ProductLogger, "AdminListPage 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.productSvc.ListPage(ctx, product.ListPageReq{
		PageResp: pagination.PageResp{
			PageNo:    req.PageNo,
			PageSize:  req.PageSize,
			SortField: req.SortField,
			SortOrder: req.SortOrder,
		},
		Name:       req.Name,
		CreditType: req.CreditType,
		Enable:     req.Enable,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.ProductLogger,
			"productSvc.ListPage 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Any("req.pageNo", req.PageNo),
			zap.Any("req.pageSize", req.PageSize),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.ProductListPageResp{
		Total:    svcResp.Total,
		PageData: toProductListItems(svcResp.PageData),
	})
}

// @Summary 后台变更商品是否启用
// @Description 用于后台快速变更商品上架/下架
// @Tags 商品管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.ProductUpdateEnableReq true "请求参数"
// @Success 200 {object} response.Response "统一响应（code=0成功，其它失败）"
// @Router /api/admin/product/enable [post]
func (h *Handler) AdminUpdateEnable(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.ProductUpdateEnableReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.ProductLogger, "AdminUpdateEnable 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	errCode, err := h.productSvc.UpdateEnable(ctx, req.ID, *req.Enable)
	if errCode != 0 {
		handler.ErrorLog(
			logger.ProductLogger,
			"productSvc.UpdateEnable 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Any("req.id", req.ID),
			zap.Any("req.enable", req.Enable),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, nil)
}

// @Summary 后台创建或变更商品
// @Description 不传 ID 或 ID 为 0 时新增商品；否则变更对应商品。规格、规格值、SKU、图片为全量覆盖，请求中未出现的即被软删除
// @Tags 商品管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.ProductSaveReq true "商品保存参数"
// @Success 200 {object} response.Response{data=resp.ProductSaveResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/product/save [post]
func (h *Handler) AdminSave(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.ProductSaveReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.ProductLogger, "AdminSave 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.productSvc.Save(ctx, toSaveReq(req))
	if errCode != 0 {
		handler.ErrorLog(
			logger.ProductLogger,
			"productSvc.Save 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Int64("req.id", req.ID),
			zap.String("req.name", req.Name),
			zap.Int("req.specs", len(req.Specs)),
			zap.Int("req.skus", len(req.Skus)),
			zap.Int("req.images", len(req.Images)),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, resp.ProductSaveResp{ID: svcResp.ID})
}

// @Summary 后台获取商品详情
// @Description 用于后台获取商品信息
// @Tags 商品管理
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.ProductDetailReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.ProductDetailResp} "统一响应（code=0成功，其它失败）"
// @Router /api/admin/product/details [post]
func (h *Handler) AdminDetails(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取管理员ID
	adminInfo, ok := handler.GetAdminInfo(c)
	if !ok {
		response.Error(c, lang, i18n.CodeTokenExpired)
		return
	}
	// 获取参数
	var req input.ProductDetailReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.ProductLogger, "AdminDetails 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.productSvc.Details(ctx, req.ID, true, true)
	if errCode != 0 {
		handler.ErrorLog(
			logger.ProductLogger,
			"productSvc.Details 调用失败",
			errCode,
			err,
			zap.Any("adminInfo", adminInfo),
			zap.Any("req.id", req.ID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, toProductDetailResp(svcResp, true))
}
