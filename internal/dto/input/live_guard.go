package input

// LiveGuardListPageReq 大航海名单查询请求
type LiveGuardListPageReq struct {
	PageNo   int `json:"pageNo" binding:"required" err:"required=11801" example:"1"`
	PageSize int `json:"pageSize" binding:"required" err:"required=11801" example:"20"`
	// 大航海等级, 1-总督, 2-提督, 3-舰长
	GuardLevel int `json:"guard_level" binding:"required" err:"required=11801" example:"3"`
	// 查询月份, 格式 2006-01
	Month string `json:"month" binding:"required" err:"required=11801" example:"2026-07"`
	// 用户uid
	UID *int64 `json:"uid" example:"123456"`
	// 用户昵称, 模糊匹配
	Uname *string `json:"uname" example:"张三"`
	// 排序列, 可选 uid / uname, 其它值静默回退默认排序
	SortField *string `json:"sortField" example:"uname"`
	SortOrder *string `json:"sortOrder" binding:"omitempty,oneof=ascend descend" err:"oneof=11801" enums:"ascend,descend" example:"descend"`
}

// LiveGuardBackfillReq 大航海开通补录请求
//
// 视作系统漏记了一条开通记录，补上后会一并补齐礼物记录、累计消费金额、
// 到期时间、开通台账与积分/星光奖励，但不会补发弹幕、不走黑名单赎回。
//
// 礼物ID与价格不在这里传：三个航海礼物是固定档位，按 guard_level 取。
type LiveGuardBackfillReq struct {
	// 用户uid
	UID int64 `json:"uid" binding:"required" err:"required=11801" example:"123456"`
	// 用户昵称
	Uname string `json:"uname" binding:"required" err:"required=11801" example:"张三"`
	// 相关房间号
	RoomID int64 `json:"room_id" binding:"required" err:"required=11801" example:"22384516"`
	// 大航海等级, 1-总督, 2-提督, 3-舰长
	GuardLevel int `json:"guard_level" binding:"required,oneof=1 2 3" err:"required=41801,oneof=41801" example:"3"`
	// 开通时间, 秒级时间戳
	OpenAt int64 `json:"open_at" binding:"required" err:"required=11801" example:"1783000000"`
	// 开通数量
	Num int64 `json:"num" binding:"required,min=1" err:"required=11801,min=11801" example:"1"`
}
