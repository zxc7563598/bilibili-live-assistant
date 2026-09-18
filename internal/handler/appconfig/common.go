package appconfig

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/appconfig"
)

// Handler App 配置 HTTP 接口处理器
type Handler struct {
	appConfigSvc *appconfig.Service
}

// New 创建 Handler 实例
func New(appConfigSvc *appconfig.Service) *Handler {
	return &Handler{
		appConfigSvc: appConfigSvc,
	}
}
