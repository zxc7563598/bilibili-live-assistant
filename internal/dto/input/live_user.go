package input

// LiveUserListPageReq 分页查询用户列表请求
type LiveUserListPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=10801" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=10801" example:"20"`
	// 排序字段
	SortField *string `json:"sortField" example:"points"`
	// 排序方向 ascend/descend
	SortOrder *string `json:"sortOrder" example:"descend" enums:"ascend,descend"`
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

// LiveUserDetailsReq 获取用户详细信息请求
type LiveUserDetailsReq struct {
	// user_id（用户表主键，非 B站 UID）
	UserID int64 `json:"user_id" binding:"required" err:"required=10801" example:"1"`
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
	Password string `json:"password" binding:"required,min=6" err:"required=10801,min=10802" example:"1"`
}

// LiveUserRefreshReq 刷新登录凭证请求
type LiveUserRefreshReq struct {
	// refresh token
	Token string `json:"token" binding:"required" err:"required=10801" example:"Bearer xxxxxxxxxx"`
}

// LiveUserChangePasswordReq 修改用户密码请求
type LiveUserChangePasswordReq struct {
	// 旧密码
	OldPassword string `json:"old_password" binding:"required,min=6" err:"required=10801,min=10802" example:"123456"`
	// 新密码
	NewPassword string `json:"new_password" binding:"required,min=6" err:"required=10801,min=10802" example:"654321"`
}

// LiveUserResetPasswordReq 重置用户密码请求
type LiveUserResetPasswordReq struct {
	// user_id（用户表主键，非 B站 UID）
	UserID int64 `json:"user_id" binding:"required" err:"required=10801" example:"1"`
	// 密码
	Password string `json:"password" binding:"required,min=6" err:"required=10801,min=10802" example:"654321"`
}

// LiveUserAssetsPageReq 用户分页查询账户记录请求
type LiveUserAssetsPageReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=10801" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=10801" example:"20"`
	// 排序字段
	SortField *string `json:"sortField" example:"created_at"`
	// 排序方向 ascend/descend
	SortOrder *string `json:"sortOrder" example:"descend" enums:"ascend,descend"`
	// 余额类型 0-星光，1-积分
	CreditType *int `json:"credit_type" example:"1" enums:"0,1"`
}

// LiveUserAssetsPageByIdReq 管理员根据用户ID分页查询账户记录请求
type LiveUserAssetsPageByIdReq struct {
	// 页码
	PageNo int `json:"pageNo" binding:"required" err:"required=10801" example:"1"`
	// 每页展示条数
	PageSize int `json:"pageSize" binding:"required" err:"required=10801" example:"20"`
	// 排序字段
	SortField *string `json:"sortField" example:"created_at"`
	// 排序方向 ascend/descend
	SortOrder *string `json:"sortOrder" example:"descend" enums:"ascend,descend"`
	// user_id（用户表主键，非 B站 UID）
	UserID int64 `json:"user_id" binding:"required" err:"required=10801" example:"1"`
	// 余额类型 0-星光，1-积分
	CreditType *int `json:"credit_type" example:"1" enums:"0,1"`
}

// LiveUserSaveBalanceReq 保存用户余额变更记录请求
type LiveUserSaveBalanceReq struct {
	// user_id（用户表主键，非 B站 UID）
	UserID int64 `json:"user_id" binding:"required" err:"required=10801" example:"1"`
	// 余额类型 0-星光，1-积分，必须显式传入
	CreditType *int `json:"credit_type" binding:"required" err:"required=10801" example:"1" enums:"0,1"`
	// 变动类型 0-减少，1-增加，必须显式传入
	ChangeType *int `json:"change_type" binding:"required" err:"required=10801" example:"1" enums:"0,1"`
	// 变动数值，传正数，增加或减少由变动类型决定
	ChangeAmount int64 `json:"change_amount" binding:"required" err:"required=10801" example:"100"`
	// 变动说明
	Remark *string `json:"remark" example:"xxxxxxxxxxxxx"`
}
