package product_image

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type gormRepo struct {
	*base.Repo[model.ProductImage]
}

func New(db *gorm.DB) Repository {
	return &gormRepo{Repo: base.NewRepo[model.ProductImage](db)}
}
