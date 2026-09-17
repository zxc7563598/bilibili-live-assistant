package product

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/product"
)

// toProductListItems 将 Service 层商品列表转换为响应结构
func toProductListItems(list []product.ListPageItem) []resp.ProductListPageItem {
	res := make([]resp.ProductListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.ProductListPageItem{
			ID:         v.ID,
			Name:       v.Name,
			Cover:      v.Cover,
			Price:      v.Price,
			CreditType: v.CreditType,
			Sold:       v.Sold,
			Stock:      v.Stock,
			Tags:       v.Tags,
			Describe:   v.Describe,
			SortOrder:  v.SortOrder,
			Enable:     v.Enable,
		})
	}
	return res
}

// toProductDetailResp 将 Service 层商品详情转换为响应结构
//
// withCostPrice 为 false 时不下发成本价（商城端场景），避免后台成本泄露给终端用户。
func toProductDetailResp(detail product.DetailsResp, withCostPrice bool) resp.ProductDetailResp {
	return resp.ProductDetailResp{
		ID:          detail.ID,
		Name:        detail.Name,
		Cover:       detail.Cover,
		Price:       detail.Price,
		CreditType:  detail.CreditType,
		Sold:        detail.Sold,
		Stock:       detail.Stock,
		Tags:        detail.Tags,
		Describe:    detail.Describe,
		SortOrder:   detail.SortOrder,
		Enable:      detail.Enable,
		ProductType: detail.ProductType,
		Skus:        toSkuItems(detail.Skus, withCostPrice),
		Specs:       toSpecItems(detail.Specs),
		Images:      toImageItems(detail.Images),
	}
}

func toSkuItems(list []product.SkuItem, withCostPrice bool) []resp.SkuItem {
	res := make([]resp.SkuItem, 0, len(list))
	for _, v := range list {
		item := resp.SkuItem{
			ID:             v.ID,
			Price:          v.Price,
			Stock:          v.Stock,
			SpecProperties: v.SpecProperties,
		}
		if withCostPrice {
			costPrice := v.CostPrice
			item.CostPrice = &costPrice
		}
		res = append(res, item)
	}
	return res
}

func toSpecItems(list []product.SpecItem) []resp.SpecItem {
	res := make([]resp.SpecItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.SpecItem{
			ID:      v.ID,
			KeyName: v.KeyName,
			Values:  toSpecValues(v.Values),
		})
	}
	return res
}

func toSpecValues(list []product.SpecValue) []resp.SpecValue {
	res := make([]resp.SpecValue, 0, len(list))
	for _, v := range list {
		res = append(res, resp.SpecValue{
			ID:        v.ID,
			ValueName: v.ValueName,
		})
	}
	return res
}

func toImageItems(list []product.ImageItem) []resp.ImageItem {
	res := make([]resp.ImageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.ImageItem{
			ID:        v.ID,
			ImagePath: v.ImagePath,
			SortOrder: v.SortOrder,
			Type:      v.Type,
			Enable:    v.Enable,
		})
	}
	return res
}

// toSaveReq 将保存请求转换为 Service 入参
func toSaveReq(req input.ProductSaveReq) product.SaveReq {
	return product.SaveReq{
		ID:          req.ID,
		Name:        req.Name,
		Cover:       req.Cover,
		Price:       req.Price,
		CreditType:  req.CreditType,
		ProductType: req.ProductType,
		Sold:        req.Sold,
		Tags:        req.Tags,
		Describe:    req.Describe,
		SortOrder:   req.SortOrder,
		Enable:      req.Enable != nil && *req.Enable,
		Specs:       toSaveSpecs(req.Specs),
		Skus:        toSaveSkus(req.Skus),
		Images:      toSaveImages(req.Images),
	}
}

func toSaveSpecs(list []input.ProductSaveSpecReq) []product.SaveSpec {
	res := make([]product.SaveSpec, 0, len(list))
	for _, v := range list {
		values := make([]product.SaveSpecValue, 0, len(v.Values))
		for _, value := range v.Values {
			values = append(values, product.SaveSpecValue{ValueName: value.ValueName})
		}
		res = append(res, product.SaveSpec{KeyName: v.KeyName, Values: values})
	}
	return res
}

func toSaveSkus(list []input.ProductSaveSkuReq) []product.SaveSku {
	res := make([]product.SaveSku, 0, len(list))
	for _, v := range list {
		res = append(res, product.SaveSku{
			Price:     v.Price,
			CostPrice: v.CostPrice,
			// Stock 为 nil 表示管理员未改动库存，交由 Service 保留库里的最新值
			Stock:          v.Stock,
			SpecProperties: v.SpecProperties,
		})
	}
	return res
}

func toSaveImages(list []input.ProductSaveImageReq) []product.SaveImage {
	res := make([]product.SaveImage, 0, len(list))
	for _, v := range list {
		res = append(res, product.SaveImage{
			ImagePath: v.ImagePath,
			SortOrder: v.SortOrder,
			Type:      v.Type,
			Enable:    v.Enable,
		})
	}
	return res
}
