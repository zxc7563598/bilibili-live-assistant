package feedback

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
)

type Repository interface {
	base.Repository[model.Feedback]
}
