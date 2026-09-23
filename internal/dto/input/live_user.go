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
	SortOrder *string `json:"sortOrder" binding:"omitempty,oneof=ascend descend" err:"oneof=10801" example:"descend" enums:"ascend,descend"`
	// 用户UID
	UID *int64 `json:"uid" example:"54272611"`
	// 用户昵称，支持模糊搜索
	Uname *string `json:"uname" example:"哎呀又胖啦"`
}

// LiveUserUserMonthlyAnalysisReq 获取用户某月弹幕/礼物/开播分析数据请求
//
// 返回的统计以自然月为界，按当月第几天聚合。
type LiveUserUserMonthlyAnalysisReq struct {
	// 用户UID
	UID int64 `json:"uid" binding:"required" err:"required=10801" example:"1"`
	// 年份，1970-2100
	Year int64 `json:"year" binding:"required,gte=1970,lte=2100" err:"required=10801,gte=10801,lte=10801" example:"2025"`
	// 月份，1-12
	Month int64 `json:"month" binding:"required,gte=1,lte=12" err:"required=10801,gte=10801,lte=10801" example:"2"`
}

// LiveUserUserDanmuAnalysisReq 获取用户弹幕分析请求
type LiveUserUserDanmuAnalysisReq struct {
	// 用户UID
	UID int64 `json:"uid" binding:"required" err:"required=10801" example:"1"`
}

// LiveUserDetailsReq 按主键获取单个用户的请求
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
//
// password 只限下限不限上限：这里是拿输入去比对已有密码，
// 加上限会让历史上设过超长密码的用户直接登不上。
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
//
// 旧密码是比对、新密码是落库，所以上限只加在新密码上。
type LiveUserChangePasswordReq struct {
	// 旧密码
	OldPassword string `json:"old_password" binding:"required,min=6" err:"required=10801,min=10802" example:"123456"`
	// 新密码
	NewPassword string `json:"new_password" binding:"required,min=6,max=32" err:"required=10801,min=10802,max=10801" example:"654321"`
}

// LiveUserResetPasswordReq 重置用户密码请求
type LiveUserResetPasswordReq struct {
	// user_id（用户表主键，非 B站 UID）
	UserID int64 `json:"user_id" binding:"required" err:"required=10801" example:"1"`
	// 密码
	Password string `json:"password" binding:"required,min=6,max=32" err:"required=10801,min=10802,max=10801" example:"654321"`
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
	SortOrder *string `json:"sortOrder" binding:"omitempty,oneof=ascend descend" err:"oneof=10801" example:"descend" enums:"ascend,descend"`
	// 余额类型 0-星光，1-积分
	CreditType *int `json:"credit_type" binding:"omitempty,oneof=0 1" err:"oneof=10801" example:"1" enums:"0,1"`
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
	SortOrder *string `json:"sortOrder" binding:"omitempty,oneof=ascend descend" err:"oneof=10801" example:"descend" enums:"ascend,descend"`
	// user_id（用户表主键，非 B站 UID）
	UserID int64 `json:"user_id" binding:"required" err:"required=10801" example:"1"`
	// 余额类型 0-星光，1-积分
	CreditType *int `json:"credit_type" binding:"omitempty,oneof=0 1" err:"oneof=10801" example:"1" enums:"0,1"`
}

// LiveUserUpdateGuardExpireReq 变更用户大航海身份请求
//
// 三个到期时间都是 Unix 秒（取本地当天 0 点），允许为空，传 null 表示清空该档位。
// 身份不落库，由到期时间实时推导，所以这里不限制时间必须晚于当前时间。
type LiveUserUpdateGuardExpireReq struct {
	// user_id（用户表主键，非 B站 UID）
	UserID int64 `json:"user_id" binding:"required" err:"required=10801" example:"1"`
	// 舰长到期时间（Unix 秒），不设置传 null
	CaptainExpireAt *int64 `json:"captain_expire_at" example:"1767225600"`
	// 提督到期时间（Unix 秒），不设置传 null
	AdmiralExpireAt *int64 `json:"admiral_expire_at" example:"1767225600"`
	// 总督到期时间（Unix 秒），不设置传 null
	GovernorExpireAt *int64 `json:"governor_expire_at" example:"1767225600"`
}

// LiveUserSaveBalanceReq 保存用户余额变更记录请求
type LiveUserSaveBalanceReq struct {
	// user_id（用户表主键，非 B站 UID）
	UserID int64 `json:"user_id" binding:"required" err:"required=10801" example:"1"`
	// 余额类型 0-星光，1-积分，必须显式传入
	CreditType *int `json:"credit_type" binding:"required,oneof=0 1" err:"required=10801,oneof=10801" example:"1" enums:"0,1"`
	// 变动类型 0-减少，1-增加，必须显式传入
	ChangeType *int `json:"change_type" binding:"required,oneof=0 1" err:"required=10801,oneof=10801" example:"1" enums:"0,1"`
	// 变动数值，传正数，增加或减少由变动类型决定
	ChangeAmount int64 `json:"change_amount" binding:"required,gt=0" err:"required=10801,gt=10801" example:"100"`
	// 变动说明
	Remark *string `json:"remark" example:"xxxxxxxxxxxxx"`
}
