package input

// LivePkLogListPageReq 分页查询 PK 对战记录请求
type LivePkLogListPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=11501" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=11501" example:"20"`
	// 排序字段
	SortField *string `json:"sortField" example:"start_at"`
	// 排序方向 ascend/descend
	SortOrder *string `json:"sortOrder" example:"descend" enums:"ascend,descend"`
	// 本直播间真实房间号，精确匹配
	RoomID *int64 `json:"room_id" example:"1972873316"`
	// 对方 UID，精确匹配
	RivalUID *int64 `json:"rival_uid" example:"3493111237970284"`
	// 对方昵称，模糊搜索
	RivalUname *string `json:"rival_uname" example:"霜煜喵Yumiao"`
	// 我方胜负，与库中 self_result 同口径：-1=我方落败 2=我方获胜
	Result *int `json:"result" example:"2" enums:"-1,2"`
	// PK 开始时间区间，毫秒级时间戳数组，取首尾两天（含当天 23:59:59）
	StartAt *[]int64 `json:"start_at" example:"1789056000000,1789315199999"`
}
