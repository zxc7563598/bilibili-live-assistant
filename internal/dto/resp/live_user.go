package resp

// LiveUserListPageResp 分页查询用户列表返回
type LiveUserListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []LiveUserListPageItem `json:"pageData"`
}

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
}

// LiveUserUserMonthlyAnalysisResp 获取用户每日分析数据返回
type LiveUserUserMonthlyAnalysisResp struct {
	// 每日弹幕数量
	DanmuCount map[int64]int64 `json:"danmu_count"`
	// 每日礼物数量
	GiftCount map[int64]int64 `json:"gift_count"`
	// 每日礼物金额
	GiftAmount map[int64]int64 `json:"gift_amount"`
	// 每日是否有开播
	LiveDays map[int64]bool `json:"live_days"`
}

// LiveUserUserDanmuAnalysisResp 获取用户弹幕分析返回
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
	UID    int64  `json:"uid" example:"4325051"`
	Avatar string `json:"avatar" example:"https://xxxxxx.xxx.com"`
	Name   string `json:"name" example:"哎呀又胖啦"`
	Points int64  `json:"points" example:"30"`
	Stars  int64  `json:"stars" example:"50"`
}

// LiveUserGetRoomIDResp 获取直播间房间号返回
type LiveUserGetRoomIDResp struct {
	RoomID int64 `json:"room_id" example:"22384516"`
}
