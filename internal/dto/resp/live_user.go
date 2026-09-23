package resp

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// LiveUserListPageResp 分页查询用户列表返回
type LiveUserListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []LiveUserListPageItem `json:"pageData"`
}

// LiveUserListPageItem 用户列表中的单条用户
type LiveUserListPageItem struct {
	// 用户ID
	ID int64 `json:"id" example:"1"`
	// 用户UID
	UID int64 `json:"uid" example:"27461511"`
	// 用户名称
	Uname string `json:"uname" example:"哎呀又胖啦"`
	// 用户积分
	Points int64 `json:"points" example:"100"`
	// 用户星光
	Stars int64 `json:"stars" example:"100"`
	// 累计发送弹幕数
	TotalDanmuCount int64 `json:"total_danmu_count" example:"100"`
	// 累计赠送礼物金额(分)
	TotalGiftAmount int64 `json:"total_gift_amount" example:"100"`
	// 会员类型
	VipType enum.BadgeType `json:"vip_type" example:"0"`
}

// LiveUserUserMonthlyAnalysisResp 获取用户某月分析数据返回
//
// 四个字段都是以「当月第几天」(1-31) 为 key 的稀疏映射：
// 当天没有数据的键不会出现，前端按月历渲染时需要自己按缺失处理。
// 注意 JSON 序列化后 map 的 key 会变成字符串。
type LiveUserUserMonthlyAnalysisResp struct {
	// 每天的弹幕条数
	DanmuCount map[int64]int64 `json:"danmu_count"`
	// 每天的礼物个数
	GiftCount map[int64]int64 `json:"gift_count"`
	// 每天的礼物金额（分）
	GiftAmount map[int64]int64 `json:"gift_amount"`
	// 当天是否开播（key 为当月第几天）
	LiveDays map[int64]bool `json:"live_days"`
}

// LiveUserUserDanmuAnalysisResp 获取用户弹幕词频分析返回
//
// 对同一批弹幕内容做不同粒度的切分，四种粒度各自独立统计、互不叠加。
type LiveUserUserDanmuAnalysisResp struct {
	// 单词数据
	Words []LiveUserWordFrequency `json:"words"`
	// 双词数据
	Bigrams []LiveUserWordFrequency `json:"bigrams"`
	// 三词数据
	Trigrams []LiveUserWordFrequency `json:"trigrams"`
	// 短句数据
	Messages []LiveUserWordFrequency `json:"messages"`
}

// LiveUserWordFrequency 词频统计项
type LiveUserWordFrequency struct {
	// 内容
	Word string `json:"word" example:"xxx"`
	// 出现次数
	Count int64 `json:"count" example:"32"`
}

// LiveUserExistsAccountResp 判断用户账号是否存在返回
type LiveUserExistsAccountResp struct {
	// 是否存在
	Exist bool `json:"exist"  example:"false"`
}

// LiveUserLoginResp 用户登录返回
type LiveUserLoginResp struct {
	// access token
	AccessToken string `json:"access_token" example:"Bearer xxxxxxxxxx"`
	// refresh token
	RefreshToken string `json:"refresh_token" example:"Bearer xxxxxxxxxx"`
}

// LiveUserUserInfoResp 获取用户基本信息返回
type LiveUserUserInfoResp struct {
	// 用户uid
	UID int64 `json:"uid" example:"4325051"`
	// 用户头像
	Avatar string `json:"avatar" example:"https://xxxxxx.xxx.com"`
	// 用户昵称
	Name string `json:"name" example:"哎呀又胖啦"`
	// 用户剩余积分
	Points int64 `json:"points" example:"30"`
	// 用户剩余星光
	Stars int64 `json:"stars" example:"50"`
	// 会员类型
	VipType enum.BadgeType `json:"vip_type" example:"0"`
}

// LiveUserGetRoomIDResp 获取当前用户绑定的直播间房间号返回
type LiveUserGetRoomIDResp struct {
	// 直播间真实房间号
	RoomID int64 `json:"room_id" example:"22384516"`
}

// LiveUserAssetsPageResp 分页查询账户变动记录返回
//
// 商城端查的是当前登录用户自己，管理端按 user_id 查指定用户，响应结构相同。
type LiveUserAssetsPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []LiveUserAssetsPageItem `json:"pageData"`
}

// LiveUserAssetsPageItem 账户变动流水中的单条记录
type LiveUserAssetsPageItem struct {
	// ID
	ID int64 `json:"id" example:"1"`
	// 备注/原因说明
	Remark string `json:"remark" example:"xxxxxxxxx"`
	// 变动类型
	ChangeType enum.ChangeType `json:"change_type" example:"1" enums:"0,1"`
	// 变动数值
	ChangeAmount int64 `json:"change_amount" example:"100"`
	// 积分类型
	CreditType enum.CreditType `json:"credit_type" example:"1" enums:"0,1"`
	// 发生时间
	CreatedAt string `json:"created_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 变动后数值
	AfterValue int64 `json:"after_value" example:"100"`
}
