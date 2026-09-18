package upload

import (
	uploadSvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/upload"
)

// Handler 图片上传 / OSS 同步 HTTP 接口处理器
type Handler struct {
	uploadSvc *uploadSvc.Service
}

// New 创建 Handler 实例
func New(uploadSvc *uploadSvc.Service) *Handler {
	return &Handler{
		uploadSvc: uploadSvc,
	}
}
