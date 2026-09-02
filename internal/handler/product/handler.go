package product

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/handler"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/product"
	"go.uber.org/zap"
)

type Handler struct {
	productSvc *product.Service
}

func New(productSvc *product.Service) *Handler {
	return &Handler{productSvc: productSvc}
}

// @Summary 商城端获取主页商品分页列表
// @Description 用于商城端主页商品列表的展示
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.ProductListPageReq true "分页参数"
// @Success 200 {object} response.Response{data=resp.ProductListPageResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/product/list [post]
func (h *Handler) ShopListPage(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取参数
	var req input.ProductListPageReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.ProductLogger, "ListPage 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	enable := int(enum.EnableEnable)
	svcResp, errCode, err := h.productSvc.ListPage(ctx, product.ListPageReq{
		PageResp: product.PageResp{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
		},
		Name:       req.Name,
		CreditType: req.CreditType,
		Enable:     &enable,
	})
	if errCode != 0 {
		handler.ErrorLog(
			logger.ProductLogger,
			"productSvc.ListPage 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
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

// @Summary 商城端获取商品详细信息
// @Description 用于商城端商品详情页的数据展示
// @Tags 移动端
// @Security BearerAuth
// @Param Accept-Language header string false "语言标识（zh: 中文，en: English）" enums(zh,en) default(zh)
// @Param data body input.ProductDetailReq true "请求参数"
// @Success 200 {object} response.Response{data=resp.ProductDetailResp} "统一响应（code=0成功，其它失败）"
// @Router /api/shop/product/detail [post]
func (h *Handler) ShopDetail(c *gin.Context) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	// 获取用户ID
	userInfo, ok := handler.GetUserInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return
	}
	// 获取参数
	var req input.ProductDetailReq
	if code, ok, err := handler.BindAndValidate(c, &req); !ok {
		handler.ErrorLog(logger.ProductLogger, "ShopDetail 参数异常", code, err)
		response.Error(c, lang, code)
		return
	}
	// 执行请求
	svcResp, errCode, err := h.productSvc.Details(ctx, req.ID, false, false)
	if errCode != 0 {
		handler.ErrorLog(
			logger.ProductLogger,
			"productSvc.Details 调用失败",
			errCode,
			err,
			zap.Any("userInfo", userInfo),
			zap.Any("req.id", req.ID),
		)
		response.Error(c, lang, errCode)
		return
	}
	// 返回结果
	response.Success(c, lang, toProductDetailResp(svcResp))
}
