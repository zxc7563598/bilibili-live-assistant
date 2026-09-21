package resp

// ExportTicketResp 领取导出下载凭证响应
type ExportTicketResp struct {
	// 下载地址（含凭证），前端据此触发浏览器原生下载
	URL string `json:"url" example:"/api/admin/export/download?ticket=xxx"`
	// 建议的文件名
	Filename string `json:"filename" example:"弹幕记录_20260921_1530.csv"`
	// 本次导出的命中行数
	Total int64 `json:"total" example:"1024"`
	// 凭证过期时间（unix 秒）
	ExpiresAt int64 `json:"expires_at" example:"1790000000"`
}
