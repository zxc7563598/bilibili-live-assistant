package upload

import "mime/multipart"

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
