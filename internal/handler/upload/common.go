package upload

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	uploadSvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/upload"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// Handler 图片上传 / OSS 同步 / 数据导入 HTTP 接口处理器
type Handler struct {
	uploadSvc *uploadSvc.Service
}

// New 创建 Handler 实例
func New(uploadSvc *uploadSvc.Service) *Handler {
	return &Handler{
		uploadSvc: uploadSvc,
	}
}

// toUploadImportProgressResp 导入进度转换为响应结构
func toUploadImportProgressResp(item uploadSvc.ImportProgressResp) resp.UploadImportProgressResp {
	return resp.UploadImportProgressResp{
		TaskID:           item.TaskID,
		Status:           item.Status,
		CurrentTable:     item.CurrentTable,
		DanmuCount:       item.DanmuCount,
		GiftCount:        item.GiftCount,
		UserCount:        item.UserCount,
		CreditLogCount:   item.CreditLogCount,
		SkippedUserCount: item.SkippedUserCount,
		BytesProcessed:   item.BytesProcessed,
		BytesTotal:       item.BytesTotal,
		Percent:          item.Percent,
		ErrorCode:        item.ErrorCode,
		StartedAt:        timeutil.Format(item.StartedAt),
		FinishedAt:       timeutil.Format(item.FinishedAt),
	}
}
