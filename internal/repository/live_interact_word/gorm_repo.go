package live_interact_word

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"gorm.io/gorm"
)

type gormRepo struct {
	*base.Repo[model.LiveInteractWord]
}

func New(db *gorm.DB) Repository {
	return &gormRepo{Repo: base.NewRepo[model.LiveInteractWord](db)}
}
