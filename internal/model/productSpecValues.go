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
