package resp

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// UploadPathResp 图片上传 / OSS 同步返回
type UploadPathResp struct {
	// 上传后可直接访问的图片路径（本地为 /uploads/ 相对路径，同步 OSS 后为 http(s) 地址）
	Path string `json:"path" example:"/uploads/login_bg/1724716800123456789_a1b2c3d4.png"`
}

// UploadImportTaskResp 数据导入任务返回
type UploadImportTaskResp struct {
	// 导入任务标识，用于查询导入进度
	TaskID string `json:"task_id" example:"8f14e45fceea167a5a36dedd4bea2543"`
}

// UploadImportProgressResp 数据导入进度返回
type UploadImportProgressResp struct {
	// 导入任务标识
	TaskID string `json:"task_id" example:"8f14e45fceea167a5a36dedd4bea2543"`
	// 任务状态（0等待中，1导入中，2已完成，3已失败）
	Status enum.ImportStatus `json:"status" example:"1" enums:"0,1,2,3"`
	// 当前正在导入的表名
	CurrentTable string `json:"current_table" example:"bl_danmu_logs"`
	// 已导入的弹幕条数
	DanmuCount int64 `json:"danmu_count" example:"198138"`
	// 已导入的礼物条数
	GiftCount int64 `json:"gift_count" example:"17697"`
	// 已导入的用户条数
	UserCount int64 `json:"user_count" example:"1624"`
	// 已补录的资产流水条数
	CreditLogCount int64 `json:"credit_log_count" example:"301"`
	// 因 uid 非法或重复被跳过的用户条数
	SkippedUserCount int64 `json:"skipped_user_count" example:"30"`
	// 已读取的压缩字节数
	BytesProcessed int64 `json:"bytes_processed" example:"3250419"`
	// 压缩文件总字节数
	BytesTotal int64 `json:"bytes_total" example:"6506835"`
	// 已读取字节百分比
	Percent float64 `json:"percent" example:"49.96"`
	// 失败时的错误码，0 表示无错误
	ErrorCode int `json:"error_code" example:"0"`
	// 任务开始时间
	StartedAt string `json:"started_at" example:"2026-09-24 16:28:19"`
	// 任务结束时间
	FinishedAt string `json:"finished_at" example:"2026-09-24 16:30:19"`
}
