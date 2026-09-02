package product

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
)

func toListPageItems(admins []model.Product) []ListPageItem {
	respList := make([]ListPageItem, 0, len(admins))
	for _, v := range admins {
		item := ListPageItem{
			ID:         v.ID,
			Name:       v.Name,
			Cover:      v.Cover,
			Price:      v.Price,
			CreditType: int(v.CreditType),
			Sold:       v.Sold,
			Stock:      v.Stock,
			Tags:       v.Tags,
			Describe:   v.Describe,
			SortOrder:  v.SortOrder,
			Enable:     v.Enable == enum.EnableEnable,
		}
		respList = append(respList, item)
	}
	return respList
}

// toDetailsResp 组装商品详情响应
func toDetailsResp(p *model.Product, skus []model.ProductSku, specs []model.ProductSpec, specValues []model.ProductSpecValue, images []model.ProductImage) DetailsResp {
	valuesBySpec := make(map[int64][]SpecValue)
	for _, v := range specValues {
		valuesBySpec[v.ProductSpecID] = append(valuesBySpec[v.ProductSpecID], SpecValue{
			ID:        v.ID,
			ValueName: v.ValueName,
		})
	}
	resp := DetailsResp{
		ID:         p.ID,
		Name:       p.Name,
		Cover:      p.Cover,
		Price:      p.Price,
		CreditType: int(p.CreditType),
		Sold:       p.Sold,
		Stock:      p.Stock,
		Tags:       p.Tags,
		Describe:   p.Describe,
		SortOrder:  p.SortOrder,
		Enable:     p.Enable == enum.EnableEnable,
		Skus:       make([]SkuItem, 0, len(skus)),
		Specs:      make([]SpecItem, 0, len(specs)),
		Images:     make([]ImageItem, 0, len(images)),
	}
	for _, s := range skus {
		resp.Skus = append(resp.Skus, SkuItem{
			ID:             s.ID,
			Price:          s.Price,
			Stock:          s.Stock,
			SpecProperties: s.SpecProperties,
		})
	}
	for _, sp := range specs {
		vals := valuesBySpec[sp.ID]
		if vals == nil {
			vals = []SpecValue{}
		}
		resp.Specs = append(resp.Specs, SpecItem{ID: sp.ID, KeyName: sp.KeyName, Values: vals})
	}
	for _, img := range images {
		resp.Images = append(resp.Images, ImageItem{
			ID:        img.ID,
			ImagePath: img.ImagePath,
			SortOrder: img.SortOrder,
			Type:      int(img.Type),
		})
	}
	return resp
}
