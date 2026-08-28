package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// 商品主表
type Product struct {
	ID         int64           `gorm:"primaryKey"`
	Name       string          `gorm:"type:varchar(255);not null;comment:商品名称"`
	Cover      string          `gorm:"type:varchar(255);not null;comment:商品封面图"`
	Price      int64           `gorm:"not null;default:0;comment:商品展示价格"`
	CreditType enum.CreditType `gorm:"type:smallint;not null;default:0;comment:积分类型"`
	Sold       int64           `gorm:"not null;default:0;comment:销售数量"`
	Stock      int64           `gorm:"not null;default:0;comment:库存数量"`
	Tags       string          `gorm:"type:varchar(255);not null;comment:商品标签，英文逗号隔开"`
	Describe   string          `gorm:"type:varchar(1000);not null;comment:商品描述"`
	SortOrder  int             `gorm:"type:int;not null;comment:排序，越大越靠前"`
	Enable     enum.Enable     `gorm:"type:smallint;not null;default:0;comment:是否启用"`
	BaseModel
}

func (Product) TableName() string {
	return "products"
}

// ProductListPageQuery 商品分页查询入参，不对应数据库表
type ProductListPageQuery struct {
	Name   *string
	Offset int
	Limit  int
}
