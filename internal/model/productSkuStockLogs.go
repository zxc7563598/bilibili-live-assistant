package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// 商品SKU库存流水表
type ProductSkuStockLog struct {
	ID           int64                `gorm:"primaryKey"`
	ProductID    int64                `gorm:"not null;comment:商品ID"`
	ProductSkuID int64                `gorm:"not null;comment:商品SKU ID"`
	ChangeNum    int64                `gorm:"not null;default:0;comment:变动数量"`
	BeforeStock  int64                `gorm:"not null;default:0;comment:变动前库存"`
	AfterStock   int64                `gorm:"not null;default:0;comment:变动后库存"`
	OrderID      string               `gorm:"comment:相关联订单ID"`
	Type         enum.StockChangeType `gorm:"type:smallint;not null;default:0;comment:变动类型"`
	BaseModel
}

func (ProductSkuStockLog) TableName() string {
	return "product_sku_stock_logs"
}

// 数据库暂时需要实现的方法：
// 暂时没想到什么其他的定的，根据其他几个product相关的表帮我做一些常用的查询/变更方法吧
