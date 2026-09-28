package input

// UploadSyncOSSReq 同步图片到 OSS 入参请求
type UploadSyncOSSReq struct {
	// 待同步图片的本地访问路径（由 /api/admin/upload/image 上传后返回的 /uploads/ 相对路径）
	Path string `json:"path" binding:"required" err:"required=11406" example:"/uploads/site_icon/1724716800123456789_a1b2c3d4.png"`
}

// UploadImportProgressReq 数据导入进度查询入参请求
type UploadImportProgressReq struct {
	// 导入任务标识（由 /api/admin/upload/import 上传后返回）
	TaskID string `json:"task_id" binding:"required" err:"required=11412" example:"8f14e45fceea167a5a36dedd4bea2543"`
}
