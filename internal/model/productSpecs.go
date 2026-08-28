package model

// 商品规格字典表
type ProductSpec struct {
	ID        int64  `gorm:"primaryKey"`
	ProductID int64  `gorm:"not null;comment:商品ID"`
	KeyName   string `gorm:"type:varchar(100);not null;comment:规格名称"`
	BaseModel
}

func (ProductSpec) TableName() string {
	return "product_specs"
}
