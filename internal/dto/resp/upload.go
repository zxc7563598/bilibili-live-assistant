package resp

// UploadPathResp 图片上传 / OSS 同步返回
type UploadPathResp struct {
	// 上传后可直接访问的图片路径（本地为 /uploads/ 相对路径，同步 OSS 后为 http(s) 地址）
	Path string `json:"path" example:"/uploads/login_bg/1724716800123456789_a1b2c3d4.png"`
}
