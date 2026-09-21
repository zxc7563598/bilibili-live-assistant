package input

import "encoding/json"

// ExportTicketReq 领取导出下载凭证请求
type ExportTicketReq struct {
	// 导出模块标识，与前端 MeCrud 的 export-module 一致
	Module string `json:"module" binding:"required" err:"required=11703" example:"livedanmu"`
	// 期望导出的列 key，顺序即 CSV 列顺序，须为模块声明过的列
	Columns []string `json:"columns" binding:"required,min=1" err:"required=11702,min=11702" example:"uid,uname,msg"`
	// 筛选条件，字段与各模块列表接口同口径；为空表示不筛
	Filters json.RawMessage `json:"filters" swaggertype:"object"`
}
