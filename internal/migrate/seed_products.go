package migrate

import (
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"gorm.io/gorm"
)

// 测试商品数据标记：写入商品 Tags 字段，用于幂等判断
const seedProductsTag = "测试"

// SeedProducts 填充商城商品测试数据（仅供测试，手动触发）
//
// 该函数不会被项目启动流程调用（启动只执行 Seed），而是通过命令行手动触发：
//
//	go run ./cmd/server/main.go -seed-products
//
// 或
//
//	make seed-products
//
// 用于本地测试商品列表 / 详情 / 下单等流程，避免生产环境启动时误填充。
// 幂等：已存在带 demo 标记的商品时直接跳过，不会重复插入。
func SeedProducts(db *gorm.DB) error {
	// 幂等检查：存在任一 demo 商品则跳过
	var count int64
	if err := db.Model(&model.Product{}).
		Where("tags LIKE ?", "%"+seedProductsTag+"%").
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		fmt.Println("[seed-products] 已存在测试商品，跳过填充")
		return nil
	}

	products := buildTestProducts()
	return db.Transaction(func(tx *gorm.DB) error {
		for _, p := range products {
			if err := createTestProduct(tx, p); err != nil {
				return fmt.Errorf("创建测试商品 %q 失败: %w", p.product.Name, err)
			}
		}
		return nil
	})
}

// testSpec 单个规格及其可选值
type testSpec struct {
	key    string
	values []string
}

// testProductSeed 一个测试商品的完整种子数据（商品 + 规格 + SKU + 图片）
type testProductSeed struct {
	product model.Product
	specs   []testSpec
	skus    []model.ProductSku
	images  []model.ProductImage
}

// createTestProduct 插入商品及其关联数据
// 商品主键使用自增 ID，插入成功后以其 ID 回填到各子表。
func createTestProduct(tx *gorm.DB, seed testProductSeed) error {
	if err := tx.Create(&seed.product).Error; err != nil {
		return err
	}
	pid := seed.product.ID

	// 规格字典 + 规格值
	for _, s := range seed.specs {
		spec := model.ProductSpec{
			ProductID: pid,
			KeyName:   s.key,
		}
		if err := tx.Create(&spec).Error; err != nil {
			return err
		}
		for _, v := range s.values {
			value := model.ProductSpecValue{
				ProductID:     pid,
				ProductSpecID: spec.ID,
				ValueName:     v,
			}
			if err := tx.Create(&value).Error; err != nil {
				return err
			}
		}
	}

	// SKU
	for i := range seed.skus {
		seed.skus[i].ProductID = pid
		if err := tx.Create(&seed.skus[i]).Error; err != nil {
			return err
		}
	}

	// 图片
	for i := range seed.images {
		seed.images[i].ProductID = pid
		if err := tx.Create(&seed.images[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// buildTestProducts 构造测试商品数据
func buildTestProducts() []testProductSeed {
	return []testProductSeed{
		{
			product: model.Product{
				Name:        "横插挂画40*60cm",
				Cover:       "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%A3%81%E7%BA%B8/cover.png",
				Price:       60,
				CreditType:  enum.CreditTypeStars,
				ProductType: enum.ProductTypeVirtual,
				Stock:       1000,
				Tags:        seedProductsTag + ",虚拟商品,壁纸",
				Describe:    "挂画尺寸40*60cm，单面印刷。",
				SortOrder:   100,
				Enable:      enum.EnableEnable,
			},
			specs: []testSpec{
				{key: "类型", values: []string{"婚纱", "未亡人", "调查员", "女鬼"}},
			},
			skus: []model.ProductSku{
				{Price: 60, CostPrice: 0, Stock: 250, SpecProperties: `[{"类型":"婚纱"}]`},
				{Price: 65, CostPrice: 0, Stock: 250, SpecProperties: `[{"类型":"调查员"}]`},
				{Price: 55, CostPrice: 0, Stock: 250, SpecProperties: `[{"类型":"女鬼"}]`},
				{Price: 70, CostPrice: 0, Stock: 250, SpecProperties: `[{"类型":"未亡人"}]`},
			},
			images: []model.ProductImage{
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%A3%81%E7%BA%B8/1.png", SortOrder: 1, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%A3%81%E7%BA%B8/2.png", SortOrder: 2, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%A3%81%E7%BA%B8/3.png", SortOrder: 3, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%A3%81%E7%BA%B8/4.png", SortOrder: 4, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
			},
		},
		{
			product: model.Product{
				Name:        "限定皮肤小粒牌",
				Cover:       "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%B0%8F%E7%B2%92%E7%89%8C/cover.jpg",
				Price:       100,
				CreditType:  enum.CreditTypePoints,
				ProductType: enum.ProductTypeActual,
				Stock:       6000,
				Tags:        seedProductsTag + ",实体商品,立牌",
				Describe:    "下单后1-3天发出哦~",
				SortOrder:   99,
				Enable:      enum.EnableEnable,
			},
			specs: []testSpec{
				{key: "类型", values: []string{"镭射款", "非镭射款"}},
				{key: "形象", values: []string{"wink小蓝", "傲娇小团子", "像素小蓝", "呼呼小狐狸", "智慧小蓝", "猥琐小蓝", "疑惑小狐狸", "舔屏小狐狸", "震惊小团子"}},
			},
			skus: []model.ProductSku{
				{Price: 100, CostPrice: 100, Stock: 500, SpecProperties: `[{"类型":"镭射款"},{"形象":"wink小蓝"}]`},
				{Price: 100, CostPrice: 100, Stock: 500, SpecProperties: `[{"类型":"镭射款"},{"形象":"傲娇小团子"}]`},
				{Price: 100, CostPrice: 100, Stock: 500, SpecProperties: `[{"类型":"镭射款"},{"形象":"像素小蓝"}]`},
				{Price: 110, CostPrice: 110, Stock: 500, SpecProperties: `[{"类型":"镭射款"},{"形象":"呼呼小狐狸"}]`},
				{Price: 105, CostPrice: 106, Stock: 500, SpecProperties: `[{"类型":"镭射款"},{"形象":"智慧小蓝"}]`},
				{Price: 95, CostPrice: 95, Stock: 500, SpecProperties: `[{"类型":"镭射款"},{"形象":"猥琐小蓝"}]`},
				{Price: 80, CostPrice: 80, Stock: 500, SpecProperties: `[{"类型":"非镭射款"},{"形象":"震惊小团子"}]`},
				{Price: 105, CostPrice: 105, Stock: 500, SpecProperties: `[{"类型":"非镭射款"},{"形象":"舔屏小狐狸"}]`},
				{Price: 100, CostPrice: 100, Stock: 500, SpecProperties: `[{"类型":"非镭射款"},{"形象":"疑惑小狐狸"}]`},
				{Price: 100, CostPrice: 100, Stock: 500, SpecProperties: `[{"类型":"非镭射款"},{"形象":"猥琐小蓝"}]`},
				{Price: 130, CostPrice: 130, Stock: 500, SpecProperties: `[{"类型":"非镭射款"},{"形象":"智慧小蓝"}]`},
				{Price: 100, CostPrice: 100, Stock: 500, SpecProperties: `[{"类型":"非镭射款"},{"形象":"呼呼小狐狸"}]`},
			},
			images: []model.ProductImage{
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%B0%8F%E7%B2%92%E7%89%8C/wink%E5%B0%8F%E8%93%9D.png", SortOrder: 1, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%B0%8F%E7%B2%92%E7%89%8C/%E5%82%B2%E5%A8%87%E5%B0%8F%E5%9B%A2%E5%AD%90.png", SortOrder: 2, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%B0%8F%E7%B2%92%E7%89%8C/%E5%83%8F%E7%B4%A0%E5%B0%8F%E8%93%9D.png", SortOrder: 2, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%B0%8F%E7%B2%92%E7%89%8C/%E5%91%BC%E5%91%BC%E5%B0%8F%E7%8B%90%E7%8B%B8.jpg", SortOrder: 2, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%B0%8F%E7%B2%92%E7%89%8C/%E6%99%BA%E6%85%A7%E5%B0%8F%E8%93%9D.jpg", SortOrder: 2, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%B0%8F%E7%B2%92%E7%89%8C/%E7%8C%A5%E7%90%90%E5%B0%8F%E8%93%9D.png", SortOrder: 2, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%B0%8F%E7%B2%92%E7%89%8C/%E7%96%91%E6%83%91%E5%B0%8F%E7%8B%90%E7%8B%B8.jpg", SortOrder: 2, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%B0%8F%E7%B2%92%E7%89%8C/%E8%88%94%E5%B1%8F%E5%B0%8F%E7%8B%90%E7%8B%B8.jpg", SortOrder: 2, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E5%B0%8F%E7%B2%92%E7%89%8C/%E9%9C%87%E6%83%8A%E5%B0%8F%E5%9B%A2%E5%AD%90.png", SortOrder: 2, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/details.png", SortOrder: 1, Type: enum.ProductImageTypeDetail, Enable: enum.EnableEnable},
			},
		},
		{
			product: model.Product{
				Name:        "Q版板绘",
				Cover:       "https://cdn.hejunjie.life/bilibilidanmu/product/%E6%89%8B%E7%BB%98/cover.png",
				Price:       100,
				CreditType:  enum.CreditTypeStars,
				ProductType: enum.ProductTypeVirtual,
				Stock:       5,
				Tags:        seedProductsTag + ",虚拟商品,头像",
				Describe:    "下单的朋友将会在每月固定板绘专场得到对应的板绘，限定5位。",
				SortOrder:   101,
				Enable:      enum.EnableEnable,
			},
			specs: []testSpec{
				{key: "类型", values: []string{"后续与主播沟通"}},
			},
			skus: []model.ProductSku{
				{Price: 100, CostPrice: 0, Stock: 5, SpecProperties: `[{"类型":"后续与主播沟通"}]`},
			},
			images: []model.ProductImage{
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E6%89%8B%E7%BB%98/%E5%B0%8F%E7%81%B0.png", SortOrder: 1, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E6%89%8B%E7%BB%98/%E5%B0%8F%E7%BB%BF.png", SortOrder: 2, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E6%89%8B%E7%BB%98/%E9%9D%93%E4%BB%94.png", SortOrder: 3, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/details.png", SortOrder: 1, Type: enum.ProductImageTypeDetail, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/details.png", SortOrder: 2, Type: enum.ProductImageTypeDetail, Enable: enum.EnableEnable},
			},
		},
		{
			product: model.Product{
				Name:        "最后的摇摇乐",
				Cover:       "https://cdn.hejunjie.life/bilibilidanmu/product/%E6%91%87%E6%91%87%E4%B9%90/cover.jpg",
				Price:       80,
				CreditType:  enum.CreditTypePoints,
				ProductType: enum.ProductTypeActual,
				Stock:       19,
				Tags:        seedProductsTag + ",实体商品,摇摇乐",
				Describe:    "仓库剩下的，没几个了，手慢无",
				SortOrder:   102,
				Enable:      enum.EnableEnable,
			},
			specs: []testSpec{
				{key: "形象", values: []string{"工会小黑", "智慧小蓝", "猥琐小蓝", "肌肉狸狸"}},
			},
			skus: []model.ProductSku{
				{Price: 80, CostPrice: 0, Stock: 1, SpecProperties: `[{"形象":"工会小黑"}]`},
				{Price: 80, CostPrice: 0, Stock: 3, SpecProperties: `[{"形象":"智慧小蓝"}]`},
				{Price: 80, CostPrice: 0, Stock: 10, SpecProperties: `[{"形象":"猥琐小蓝"}]`},
				{Price: 80, CostPrice: 0, Stock: 5, SpecProperties: `[{"形象":"肌肉狸狸"}]`},
			},
			images: []model.ProductImage{
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E6%91%87%E6%91%87%E4%B9%90/%E5%B7%A5%E4%BC%9A%E5%B0%8F%E9%BB%91.png", SortOrder: 1, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E6%91%87%E6%91%87%E4%B9%90/%E6%99%BA%E6%85%A7%E5%B0%8F%E8%93%9D.jpg", SortOrder: 1, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E6%91%87%E6%91%87%E4%B9%90/%E7%8C%A5%E7%90%90%E5%B0%8F%E8%93%9D.jpg", SortOrder: 1, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/%E6%91%87%E6%91%87%E4%B9%90/%E8%82%8C%E8%82%89%E7%8B%B8%E7%8B%B8.jpg", SortOrder: 1, Type: enum.ProductImageTypeBanner, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/details.png", SortOrder: 1, Type: enum.ProductImageTypeDetail, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/avatars/shop.png", SortOrder: 2, Type: enum.ProductImageTypeDetail, Enable: enum.EnableEnable},
				{ImagePath: "https://cdn.hejunjie.life/bilibilidanmu/product/details.png", SortOrder: 3, Type: enum.ProductImageTypeDetail, Enable: enum.EnableEnable},
			},
		},
	}
}
