package input

// LiveUserListPageReq 分页查询用户列表请求
type LiveUserListPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=10801" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=10801" example:"20"`
	// 用户UID
	UID *int64 `json:"uid" example:"54272611"`
	// 用户昵称，支持模糊搜索
	Uname *string `json:"uname" example:"哎呀又胖啦"`
}

// LiveUserUserMonthlyAnalysisReq 获取用户每日分析数据请求
type LiveUserUserMonthlyAnalysisReq struct {
	// 用户UID
	UID int64 `json:"uid" binding:"required" err:"required=10801" example:"1"`
	// 年份
	Year int64 `json:"year" binding:"required" err:"required=10801" example:"2025"`
	// 月份
	Month int64 `json:"month" binding:"required" err:"required=10801" example:"2"`
}

// LiveUserUserDanmuAnalysisReq 获取用户弹幕分析请求
type LiveUserUserDanmuAnalysisReq struct {
	// 用户UID
	UID int64 `json:"uid" binding:"required" err:"required=10801" example:"1"`
}

// LiveUserExistsAccountReq 判断用户账号是否存在请求
type LiveUserExistsAccountReq struct {
	// 用户账号(UID)
	Account int64 `json:"account" binding:"required" err:"required=10801" example:"1"`
}

// LiveUserLoginReq 用户登录请求
type LiveUserLoginReq struct {
	// 用户账号(UID)
	Account int64 `json:"account" binding:"required" err:"required=10801" example:"1"`
	// 用户密码
	Password string `json:"password" binding:"required" err:"required=10801" example:"1"`
}

// LiveUserRefreshReq 刷新登录凭证请求
type LiveUserRefreshReq struct {
	// refresh token
	Token string `json:"token" binding:"required" err:"required=10801" example:"Bearer xxxxxxxxxx"`
}
