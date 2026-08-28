package model

// 商品规格数值表
type ProductSpecValue struct {
	ID            int64  `gorm:"primaryKey"`
	ProductID     int64  `gorm:"not null;comment:商品ID"`
	ProductSpecID int64  `gorm:"not null;comment:规格ID"`
	ValueName     string `gorm:"type:varchar(100);not null;comment:数值名称"`
	BaseModel
}

func (ProductSpecValue) TableName() string {
	return "product_spec_values"
}

// 数据库暂时需要实现的方法：
// 暂时没想到什么其他的定的，根据其他几个product相关的表帮我做一些常用的查询/变更方法吧
