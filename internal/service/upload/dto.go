package upload

import (
	"mime/multipart"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
)

// UploadImageReq 图片上传入参：上传场景 + multipart 文件头
type UploadImageReq struct {
	// 上传场景（决定落盘子目录，取值见 common.go 的 uploadSceneAllow）
	Scene string
	// 待落盘的上传文件
	File *multipart.FileHeader
}

// UploadPathResp 图片上传 / OSS 同步出参：可直接访问的图片路径
type UploadPathResp struct {
	// 本地落盘返回 /uploads/ 相对路径，或同步 OSS 后返回的 http(s) 地址
	Path string
}

// ImportDataReq 数据导入入参：待导入的 .gz 导出文件
type ImportDataReq struct {
	// 旧版导出功能产出的 .gz 数据文件
	File *multipart.FileHeader
	// 发起导入的管理员 ID，仅用于日志溯源
	AdminID int64
}

// ImportTaskResp 数据导入任务出参：任务标识，用于查询进度
type ImportTaskResp struct {
	TaskID string
}

// ImportProgressResp 数据导入进度出参
type ImportProgressResp struct {
	// 任务标识
	TaskID string
	// 任务状态
	Status enum.ImportStatus
	// 当前正在导入的表名
	CurrentTable string
	// 已导入的弹幕条数
	DanmuCount int64
	// 已导入的礼物条数
	GiftCount int64
	// 已导入的用户条数
	UserCount int64
	// 已补录的资产流水条数
	CreditLogCount int64
	// 因 uid 非法或重复而被跳过的用户条数
	SkippedUserCount int64
	// 已读取的压缩字节数
	BytesProcessed int64
	// 压缩文件总字节数
	BytesTotal int64
	// 已读取字节百分比
	Percent float64
	// 失败时的错误码，0 表示无错误
	ErrorCode int
	// 任务开始时间（unix 秒）
	StartedAt int64
	// 任务结束时间（unix 秒），未结束时为 0
	FinishedAt int64
}
