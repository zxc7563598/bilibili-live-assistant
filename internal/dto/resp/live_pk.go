package resp

// LivePkLogListPageResp 分页查询 PK 对战记录返回
type LivePkLogListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []LivePkLogListPageItem `json:"pageData"`
	// 统计信息，与列表受相同筛选条件影响
	Stats LivePkLogListPageStats `json:"stats"`
}

// LivePkLogListPageItem PK 对战记录列表中的单条记录
type LivePkLogListPageItem struct {
	// 记录ID
	ID int64 `json:"id" example:"1"`
	// 本直播间真实房间号
	RoomID int64 `json:"room_id" example:"1972873316"`
	// PK ID
	PkID int64 `json:"pk_id" example:"399983699"`
	// PK 状态：101=即将开始 401=正常结束 404=异常结束
	PkStatus int64 `json:"pk_status" example:"401"`
	// 对战类型：2=经典PK 6=大乱斗
	BattleType int64 `json:"battle_type" example:"2"`
	// 匹配类型
	MatchType int64 `json:"match_type" example:"1"`
	// 对方UID
	RivalUID int64 `json:"rival_uid" example:"3493111237970284"`
	// 对方用户名
	RivalUname string `json:"rival_uname" example:"霜煜喵Yumiao"`
	// 对方房间ID
	RivalRoomID int64 `json:"rival_room_id" example:"1727071466"`
	// 我方最终PK值
	SelfVotes int64 `json:"self_votes" example:"6160"`
	// 对方最终PK值
	RivalVotes int64 `json:"rival_votes" example:"10"`
	// 我方胜负，库中原始值：2=获胜 -1=落败 0=未能定位本直播间
	SelfResult int64 `json:"self_result" example:"2"`
	// 对方胜负，库中原始值：2=获胜 -1=落败
	RivalResult int64 `json:"rival_result" example:"-1"`
	// PK 开始时间，未开始为空串
	StartAt string `json:"start_at" example:"2025-01-02 12:22:22"`
	// PK 结算时间，未收到结算事件时为空串
	SettleAt string `json:"settle_at" example:"2025-01-02 12:27:22"`
}

// LivePkLogListPageStats PK 记录的汇总统计（与列表受相同筛选条件影响）
type LivePkLogListPageStats struct {
	// PK 场数（self_result 为 0 的记录既不算胜也不算负，所以场数不一定等于胜 + 负）
	TotalNum int64 `json:"total_num"`
	// 我方胜利场数
	WinNum int64 `json:"win_num"`
	// 我方失败场数
	LoseNum int64 `json:"lose_num"`
}

// LivePkFetchRoomGroupsResp 获取全部房间ID返回
type LivePkFetchRoomGroupsResp struct {
	Option []LivePkFetchRoomGroupsItem `json:"option"`
}

type LivePkFetchRoomGroupsItem struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}
