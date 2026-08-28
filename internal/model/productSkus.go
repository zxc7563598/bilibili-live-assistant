package model

// 商品SKU表
type ProductSku struct {
	ID             int64  `gorm:"primaryKey"`
	ProductID      int64  `gorm:"not null;comment:商品ID"`
	Price          int64  `gorm:"not null;default:0;comment:商品价格"`
	CostPrice      int64  `gorm:"not null;default:0;comment:成本价，用来核算成本，金额分"`
	Stock          int64  `gorm:"not null;default:0;comment:库存数量"`
	SpecProperties string `gorm:"type:varchar(1000);not null;comment:规格快照，[{key_name:value_name}]"`
	BaseModel
}

func (ProductSku) TableName() string {
	return "product_skus"
}

// 数据库暂时需要实现的方法：
// 暂时没想到什么其他的定的，根据其他几个product相关的表帮我做一些常用的查询/变更方法吧
