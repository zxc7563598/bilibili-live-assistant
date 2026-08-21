package live_user_credit_log

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
)

// Repository 接口定义
type Repository interface {
	base.Repository[model.LiveUserCreditLog]
}
