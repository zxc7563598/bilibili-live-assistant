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
	OrderID      int64                `gorm:"comment:相关联订单ID"`
	OrderSn      string               `gorm:"type:varchar(50);comment:订单号"`
	DraftID      int64                `gorm:"comment:关联草稿ID"`
	UserID       int64                `gorm:"comment:用户ID"`
	Type         enum.StockChangeType `gorm:"type:smallint;not null;default:0;comment:变动类型"`
	BaseModel
}

func (ProductSkuStockLog) TableName() string {
	return "product_sku_stock_logs"
}
