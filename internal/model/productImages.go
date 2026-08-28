package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// 商品图片表
type ProductImage struct {
	ID        int64                 `gorm:"primaryKey"`
	ProductID int64                 `gorm:"not null;comment:商品ID"`
	ImagePath string                `gorm:"type:varchar(1000);not null;comment:图片路径"`
	SortOrder int                   `gorm:"type:int;not null;comment:排序，越大越靠前"`
	Type      enum.ProductImageType `gorm:"type:smallint;not null;default:0;comment:图片类型"`
	Enable    enum.Enable           `gorm:"type:smallint;not null;default:0;comment:是否启用"`
	BaseModel
}

func (ProductImage) TableName() string {
	return "product_images"
}

// 数据库暂时需要实现的方法：
// 暂时没想到什么其他的定的，根据其他几个product相关的表帮我做一些常用的查询/变更方法吧
