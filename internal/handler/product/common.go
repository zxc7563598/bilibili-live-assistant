package product

import (
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
		})
	}
	return res
}

// toProductDetailResp 将 Service 层商品详情转换为响应结构
func toProductDetailResp(detail product.DetailsResp) resp.ProductDetailResp {
	return resp.ProductDetailResp{
		ID:         detail.ID,
		Name:       detail.Name,
		Cover:      detail.Cover,
		Price:      detail.Price,
		CreditType: detail.CreditType,
		Sold:       detail.Sold,
		Stock:      detail.Stock,
		Tags:       detail.Tags,
		Describe:   detail.Describe,
		Skus:       toSkuItems(detail.Skus),
		Specs:      toSpecItems(detail.Specs),
		Images:     toImageItems(detail.Images),
	}
}

func toSkuItems(list []product.SkuItem) []resp.SkuItem {
	res := make([]resp.SkuItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.SkuItem{
			ID:             v.ID,
			Price:          v.Price,
			Stock:          v.Stock,
			SpecProperties: v.SpecProperties,
		})
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
			Type:      v.Type,
		})
	}
	return res
}
