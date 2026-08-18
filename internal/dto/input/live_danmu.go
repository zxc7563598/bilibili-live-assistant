package input

// LiveDanmuListPageReq 分页查询弹幕列表请求
type LiveDanmuListPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=10601" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=10601" example:"20"`
	// 房间ID
	RoomID *int64 `json:"room_id" example:"22384516"`
	// 用户ID
	UID *int64 `json:"uid" example:"54272611"`
	// 用户昵称，支持模糊搜索
	Uname *string `json:"uname" example:"哎呀又胖啦"`
	// 弹幕内容，支持模糊搜索
	Msg *string `json:"msg" example:"好耶"`
	// 弹幕发送时间区间
	SendAt *[]int64 `json:"send_at" example:"1183135260000,1787024216391"`
}
