package version

import (
	versionsvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/version"
)

// Handler 版本 HTTP 接口处理器
type Handler struct {
	versionSvc *versionsvc.Service
}

// New 创建 Handler 实例
func New(versionSvc *versionsvc.Service) *Handler {
	return &Handler{
		versionSvc: versionSvc,
	}
}
