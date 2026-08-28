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

// 数据库暂时需要实现的方法：
// 暂时没想到什么其他的定的，根据其他几个product相关的表帮我做一些常用的查询/变更方法吧
