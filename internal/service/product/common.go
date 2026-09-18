package product

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"unicode/utf8"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"gorm.io/gorm"
)

// 子表字段长度上限，与 model 层的列宽保持一致
const (
	maxSpecNameLen       = 100
	maxSpecValueLen      = 100
	maxImagePathLen      = 1000
	maxSpecPropertiesLen = 1000
)

// maxSkuCount 单个商品允许的 SKU 数量上限，与前端规格组合数上限一致
const maxSkuCount = 200

// errProductNotFound 商品不存在或已被删除
var errProductNotFound = errors.New("商品不存在")

// specDefinition 商品规格定义，用于校验 SKU 规格快照
type specDefinition struct {
	// keys 规格名，按请求中的定义顺序
	keys []string
	// values 规格名 -> 规格值集合
	values map[string]map[string]struct{}
}

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
		ID:          p.ID,
		Name:        p.Name,
		Cover:       p.Cover,
		Price:       p.Price,
		CreditType:  int(p.CreditType),
		Sold:        p.Sold,
		Stock:       p.Stock,
		Tags:        p.Tags,
		Describe:    p.Describe,
		SortOrder:   p.SortOrder,
		Enable:      p.Enable == enum.EnableEnable,
		ProductType: int(p.ProductType),
		Skus:        make([]SkuItem, 0, len(skus)),
		Specs:       make([]SpecItem, 0, len(specs)),
		Images:      make([]ImageItem, 0, len(images)),
	}
	for _, s := range skus {
		resp.Skus = append(resp.Skus, SkuItem{
			ID:             s.ID,
			Price:          s.Price,
			CostPrice:      s.CostPrice,
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
			Enable:    img.Enable == enum.EnableEnable,
		})
	}
	return resp
}

// validateSaveReq 校验保存请求里 binding 标签表达不了的嵌套规则
func validateSaveReq(req SaveReq) (int, error) {
	if len(req.Skus) == 0 {
		return 11012, errors.New("未上架任何规格组合")
	}
	if len(req.Skus) > maxSkuCount {
		return 11013, errors.New("规格组合数量超限")
	}
	def, errCode, err := validateSaveSpecs(req.Specs)
	if errCode != 0 {
		return errCode, err
	}
	// 规格快照落库前先比对，避免同一组合出现两条 SKU 导致库存被重复汇总
	seenCombos := make(map[string]struct{}, len(req.Skus))
	for _, sku := range req.Skus {
		if sku.Price < 0 {
			return 11022, errors.New("SKU 价格非法")
		}
		if sku.CostPrice < 0 {
			return 11028, errors.New("SKU 成本价非法")
		}
		if sku.Stock != nil && *sku.Stock < 0 {
			return 11023, errors.New("SKU 库存非法")
		}
		if errCode, err := validateSkuProps(def, sku.SpecProperties); errCode != 0 {
			return errCode, err
		}
		specProperties, err := marshalSpecProperties(def.keys, sku.SpecProperties)
		if err != nil {
			return 11021, err
		}
		if utf8.RuneCountInString(specProperties) > maxSpecPropertiesLen {
			return 11029, errors.New("SKU 规格快照过长")
		}
		comboKey := canonicalSpecProperties(specProperties)
		if _, ok := seenCombos[comboKey]; ok {
			return 11030, errors.New("SKU 规格组合重复")
		}
		seenCombos[comboKey] = struct{}{}
	}
	for _, img := range req.Images {
		if img.ImagePath == "" {
			return 11024, errors.New("图片地址为空")
		}
		if utf8.RuneCountInString(img.ImagePath) > maxImagePathLen {
			return 11025, errors.New("图片地址过长")
		}
		if !enum.ProductImageType(img.Type).IsValid() {
			return 11026, errors.New("图片类型非法")
		}
	}
	return 0, nil
}

// validateSaveSpecs 校验规格与规格值，返回规格定义
func validateSaveSpecs(specs []SaveSpec) (specDefinition, int, error) {
	def := specDefinition{
		keys:   make([]string, 0, len(specs)),
		values: make(map[string]map[string]struct{}, len(specs)),
	}
	for _, spec := range specs {
		if spec.KeyName == "" {
			return def, 11015, errors.New("规格名称为空")
		}
		if utf8.RuneCountInString(spec.KeyName) > maxSpecNameLen {
			return def, 11016, errors.New("规格名称过长")
		}
		// 重名规格会在规格快照里互相覆盖，必须拦掉
		if _, ok := def.values[spec.KeyName]; ok {
			return def, 11017, errors.New("规格名称重复")
		}
		def.keys = append(def.keys, spec.KeyName)

		if len(spec.Values) == 0 {
			return def, 11018, errors.New("规格下没有规格值")
		}
		valueSet := make(map[string]struct{}, len(spec.Values))
		for _, value := range spec.Values {
			if value.ValueName == "" {
				return def, 11018, errors.New("规格值为空")
			}
			if utf8.RuneCountInString(value.ValueName) > maxSpecValueLen {
				return def, 11019, errors.New("规格值过长")
			}
			if _, ok := valueSet[value.ValueName]; ok {
				return def, 11020, errors.New("规格值重复")
			}
			valueSet[value.ValueName] = struct{}{}
		}
		def.values[spec.KeyName] = valueSet
	}
	return def, 0, nil
}

// validateSkuProps 校验 SKU 规格快照与商品规格定义完全一致
func validateSkuProps(def specDefinition, pairs []map[string]string) (int, error) {
	props, err := flattenSpecProperties(pairs)
	if err != nil {
		return 11021, err
	}
	if len(props) != len(def.keys) {
		return 11021, errors.New("SKU 规格快照与商品规格数量不一致")
	}
	for _, key := range def.keys {
		value, ok := props[key]
		if !ok || value == "" {
			return 11021, errors.New("SKU 规格快照缺少规格或规格值为空")
		}
		if _, ok := def.values[key][value]; !ok {
			return 11021, errors.New("SKU 规格快照引用了未定义的规格值")
		}
	}
	return 0, nil
}

// flattenSpecProperties 把规格快照拍平成「规格名 -> 规格值」
func flattenSpecProperties(pairs []map[string]string) (map[string]string, error) {
	props := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		for key, value := range pair {
			if _, ok := props[key]; ok {
				return nil, errors.New("SKU 规格快照存在重复的规格")
			}
			props[key] = value
		}
	}
	return props, nil
}

// marshalSpecProperties 按规格定义顺序把规格快照序列化成落库字符串
func marshalSpecProperties(specKeys []string, pairs []map[string]string) (string, error) {
	props, err := flattenSpecProperties(pairs)
	if err != nil {
		return "", err
	}
	ordered := make([]map[string]string, 0, len(specKeys))
	for _, key := range specKeys {
		ordered = append(ordered, map[string]string{key: props[key]})
	}
	b, err := json.Marshal(ordered)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// parseSpecProperties marshalSpecProperties 的逆操作
func parseSpecProperties(specProperties string) (map[string]string, error) {
	var pairs []map[string]string
	if err := json.Unmarshal([]byte(specProperties), &pairs); err != nil {
		return nil, err
	}
	return flattenSpecProperties(pairs)
}

// canonicalSpecProperties 把落库的规格快照转成与规格顺序无关的比较键
func canonicalSpecProperties(specProperties string) string {
	props, err := parseSpecProperties(specProperties)
	if err != nil {
		return specProperties
	}
	keys := make([]string, 0, len(props))
	for key := range props {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	ordered := make([]map[string]string, 0, len(keys))
	for _, key := range keys {
		ordered = append(ordered, map[string]string{key: props[key]})
	}
	b, err := json.Marshal(ordered)
	if err != nil {
		return specProperties
	}
	return string(b)
}

// toProductEntity 用请求数据填充商品实体，新增与变更共用
func toProductEntity(req SaveReq) model.Product {
	return model.Product{
		Name:        req.Name,
		Cover:       req.Cover,
		Price:       req.Price,
		CreditType:  enum.CreditType(req.CreditType),
		ProductType: enum.ProductType(req.ProductType),
		Sold:        req.Sold,
		Tags:        req.Tags,
		Describe:    req.Describe,
		SortOrder:   req.SortOrder,
		Enable:      enum.BoolToEnable(req.Enable),
	}
}

// syncRelations 按请求同步商品的规格 / 规格值 / SKU / 图片
func (s *Service) syncRelations(ctx context.Context, tx *gorm.DB, productID int64, req SaveReq) error {
	specKeys, err := s.syncSpecs(ctx, tx, productID, req.Specs)
	if err != nil {
		return err
	}
	if err := s.syncSkus(ctx, tx, productID, specKeys, req.Skus); err != nil {
		return err
	}
	return s.syncImages(ctx, tx, productID, req.Images)
}

// syncSpecs 重建商品的规格与规格值，返回按请求顺序排列的规格名
func (s *Service) syncSpecs(ctx context.Context, tx *gorm.DB, productID int64, specs []SaveSpec) ([]string, error) {
	if err := s.productSpecValueRepo.DeleteByProductID(ctx, tx, productID); err != nil {
		return nil, err
	}
	if err := s.productSpecRepo.DeleteByProductID(ctx, tx, productID); err != nil {
		return nil, err
	}
	specKeys := make([]string, 0, len(specs))
	for _, spec := range specs {
		created, err := s.productSpecRepo.Create(ctx, tx, &model.ProductSpec{
			ProductID: productID,
			KeyName:   spec.KeyName,
		})
		if err != nil {
			return nil, err
		}
		specKeys = append(specKeys, spec.KeyName)
		// 规格值需要拿到落库后的规格 ID 才能写入
		values := make([]model.ProductSpecValue, 0, len(spec.Values))
		for _, value := range spec.Values {
			values = append(values, model.ProductSpecValue{
				ProductID:     productID,
				ProductSpecID: created.ID,
				ValueName:     value.ValueName,
			})
		}
		if err := s.productSpecValueRepo.CreateBatch(ctx, tx, values); err != nil {
			return nil, err
		}
	}
	return specKeys, nil
}

// syncSkus 差量同步商品的 SKU
func (s *Service) syncSkus(ctx context.Context, tx *gorm.DB, productID int64, specKeys []string, skus []SaveSku) error {
	existing, err := s.productSkuRepo.ListByProductID(ctx, tx, productID)
	if err != nil {
		return err
	}
	unmatched := make(map[string]*model.ProductSku, len(existing))
	dupes := make([]int64, 0)
	for i := range existing {
		key := canonicalSpecProperties(existing[i].SpecProperties)
		if _, ok := unmatched[key]; ok {
			// 历史脏数据：同一规格组合落了两行，保留先出现的一行，多余的随后清掉
			dupes = append(dupes, existing[i].ID)
			continue
		}
		unmatched[key] = &existing[i]
	}
	created := make([]model.ProductSku, 0, len(skus))
	for _, sku := range skus {
		specProperties, err := marshalSpecProperties(specKeys, sku.SpecProperties)
		if err != nil {
			return err
		}
		comboKey := canonicalSpecProperties(specProperties)
		old, ok := unmatched[comboKey]
		if !ok {
			stock := int64(0)
			if sku.Stock != nil {
				stock = *sku.Stock
			}
			created = append(created, model.ProductSku{
				ProductID:      productID,
				Price:          sku.Price,
				CostPrice:      sku.CostPrice,
				Stock:          stock,
				SpecProperties: specProperties,
			})
			continue
		}
		delete(unmatched, comboKey)
		old.Price = sku.Price
		old.CostPrice = sku.CostPrice
		// Stock 为 nil 表示管理员没改过这个组合的库存，留库里的最新值，避免覆盖保存期间被订单扣减的库存
		if sku.Stock != nil {
			old.Stock = *sku.Stock
		}
		old.SpecProperties = specProperties
		if err := s.productSkuRepo.Save(ctx, tx, old); err != nil {
			return err
		}
	}
	removed := make([]int64, 0, len(unmatched)+len(dupes))
	removed = append(removed, dupes...)
	for _, left := range unmatched {
		removed = append(removed, left.ID)
	}
	if err := s.productSkuRepo.DeleteByIDs(ctx, tx, removed); err != nil {
		return err
	}
	return s.productSkuRepo.CreateBatch(ctx, tx, created)
}

// syncImages 重建商品的轮播图与详情图
func (s *Service) syncImages(ctx context.Context, tx *gorm.DB, productID int64, images []SaveImage) error {
	if err := s.productImageRepo.DeleteByProductID(ctx, tx, productID); err != nil {
		return err
	}
	list := make([]model.ProductImage, 0, len(images))
	for _, img := range images {
		list = append(list, model.ProductImage{
			ProductID: productID,
			ImagePath: img.ImagePath,
			SortOrder: img.SortOrder,
			Type:      enum.ProductImageType(img.Type),
			Enable:    enum.BoolToEnable(img.Enable),
		})
	}
	return s.productImageRepo.CreateBatch(ctx, tx, list)
}
