package input

// UploadSyncOSSReq 同步图片到 OSS 入参请求
type UploadSyncOSSReq struct {
	// 待同步图片的本地访问路径（由 /api/admin/upload/image 上传后返回的 /uploads/ 相对路径）
	Path string `json:"path" binding:"required" err:"required=11406" example:"/uploads/site_icon/1724716800123456789_a1b2c3d4.png"`
}
