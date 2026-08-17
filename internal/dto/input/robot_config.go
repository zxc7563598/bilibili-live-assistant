package input

// RoomConfigReq 房间模块配置请求
type RoomConfigReq struct {
	// 是否默认监听直播间, 0-否, 1-是
	IsListening string `json:"is_listening" binding:"required" err:"required=10501" example:"1"`
	// 用户名最大长度, 超过此长度则裁剪
	MaxNameLength string `json:"max_name_length" binding:"required" err:"required=10501" example:"8"`
	// 裁剪方式, 0-省略后面, 1-省略前面
	NameTrimMode string `json:"name_trim_mode" binding:"required" err:"required=10501" example:"0"`
	// 用户消费发放奖励, 0-不发放, 1-按消费电池发放, 2-按开通航海类型发放
	ConsumeRewardEnabled string `json:"consume_reward_enabled" binding:"required" err:"required=10501" example:"0"`
	// 奖励类型, 0-星光, 1-积分
	RewardType string `json:"reward_type" binding:"required" err:"required=10501" example:"1"`
	// 消费电池转换倍率, 奖励设置为按消费电池发放时生效
	ConsumeBatteryRate string `json:"consume_battery_rate" binding:"required" err:"required=10501" example:"0"`
	// 开通舰长奖励数量, 奖励设置为按开通航海类型发放时生效
	CaptainRewardAmount string `json:"captain_reward_amount" binding:"required" err:"required=10501" example:"0"`
	// 开通提督奖励数量, 奖励设置为按开通航海类型发放时生效
	CommanderRewardAmount string `json:"commander_reward_amount" binding:"required" err:"required=10501" example:"0"`
	// 开通总督奖励数量, 奖励设置为按开通航海类型发放时生效
	GovernorRewardAmount string `json:"governor_reward_amount" binding:"required" err:"required=10501" example:"0"`
}

// SignConfigReq 签到模块配置请求
type SignConfigReq struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" binding:"required" err:"required=10501" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" binding:"required" err:"required=10501" example:"1"`
	// 触发门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" binding:"required" err:"required=10501" example:"0"`
	// 奖励类型, 0-星光, 1-积分
	RewardType string `json:"reward_type" binding:"required" err:"required=10501" example:"0"`
	// 奖励数量, 正整数
	RewardAmount string `json:"reward_amount" binding:"required" err:"required=10501" example:"10"`
	// 签到关键词
	Keyword string `json:"keyword" binding:"required" err:"required=10501" example:"#签到"`
	// 查询关键词
	QueryKeyword string `json:"query_keyword" binding:"required" err:"required=10501" example:"#查询"`
	// 签到成功回复（JSON 数组字符串）
	SuccessReply []string `json:"success_reply" binding:"required" err:"required=10501"`
	// 签到失败回复（JSON 数组字符串）
	FailReply []string `json:"fail_reply" binding:"required" err:"required=10501"`
	// 重复签到回复（JSON 数组字符串）
	RepeatReply []string `json:"repeat_reply" binding:"required" err:"required=10501"`
	// 查询成功回复（JSON 数组字符串）
	QueryReply []string `json:"query_reply" binding:"required" err:"required=10501"`
}

// AdConfigReq 定时广告模块配置请求
type AdConfigReq struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" binding:"required" err:"required=10501" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" binding:"required" err:"required=10501" example:"1"`
	// 发送间隔, 秒
	Interval string `json:"interval" binding:"required" err:"required=10501" example:"62"`
	// 发送方式, 0-随机发送, 1-顺序发送
	SendMode string `json:"send_mode" binding:"required" err:"required=10501" example:"0"`
	// 发送内容（JSON 数组字符串）
	Content []string `json:"content" binding:"required" err:"required=10501"`
}

// GiftConfigReq 礼物答谢模块配置请求
type GiftConfigReq struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" binding:"required" err:"required=10501" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" binding:"required" err:"required=10501" example:"1"`
	// 答谢门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" binding:"required" err:"required=10501" example:"0"`
	// 展示数量, 0-禁用, 1-启用
	ShowCount string `json:"show_count" binding:"required" err:"required=10501" example:"1"`
	// 礼物合并, 0-禁用, 1-启用
	MergeGift string `json:"merge_gift" binding:"required" err:"required=10501" example:"1"`
	// 盲盒统计, 0-禁用, 1-启用
	IncludeBlindbox string `json:"include_blindbox" binding:"required" err:"required=10501" example:"1"`
	// 起始感谢电池数
	MinBattery string `json:"min_battery" binding:"required" err:"required=10501" example:"10"`
	// 感谢内容（JSON 数组字符串）
	Content []string `json:"content" binding:"required" err:"required=10501"`
}

// PkConfigReq PK播报模块配置请求
type PkConfigReq struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" binding:"required" err:"required=10501" example:"1"`
	// 发送内容（JSON 数组字符串）
	Content []string `json:"content" binding:"required" err:"required=10501"`
}

// WelcomeConfigReq 进房欢迎模块配置请求
type WelcomeConfigReq struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" binding:"required" err:"required=10501" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" binding:"required" err:"required=10501" example:"1"`
	// 欢迎门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" binding:"required" err:"required=10501" example:"0"`
	// 欢迎内容（JSON 数组字符串）
	Content []string `json:"content" binding:"required" err:"required=10501"`
}

// FollowConfigReq 感谢关注模块配置请求
type FollowConfigReq struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" binding:"required" err:"required=10501" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" binding:"required" err:"required=10501" example:"1"`
	// 感谢门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" binding:"required" err:"required=10501" example:"0"`
	// 感谢内容（JSON 数组字符串）
	Content []string `json:"content" binding:"required" err:"required=10501"`
}

// ShareConfigReq 感谢分享模块配置请求
type ShareConfigReq struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" binding:"required" err:"required=10501" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" binding:"required" err:"required=10501" example:"1"`
	// 感谢门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" binding:"required" err:"required=10501" example:"0"`
	// 感谢内容（JSON 数组字符串）
	Content []string `json:"content" binding:"required" err:"required=10501"`
}

// ReplyConfigReq 自动回复模块配置请求
type ReplyConfigReq struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" binding:"required" err:"required=10501" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" binding:"required" err:"required=10501" example:"1"`
	// 触发门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" binding:"required" err:"required=10501" example:"0"`
	// 回复内容（JSON 配置）
	Content []ReplyItem `json:"content" binding:"required" err:"required=10501"`
}

// ReplyItem 自动回复规则项
type ReplyItem struct {
	// 触发关键词列表
	Keyword []string `json:"keyword" example:"你好,在吗"`
	// 关键词匹配策略
	KeywordMatchPolicy string `json:"keyword_match_policy" example:"0"`
	// 安全词列表
	SafeWord []string `json:"safe_word" example:"骗子,哎呀"`
	// 安全词匹配策略
	SafeWordMatchPolicy string `json:"safe_word_match_policy" example:"0"`
	// 是否禁言发送者, 0-否, 1-是
	MuteSender string `json:"mute_sender" example:"0"`
	// 禁言时长（分钟）, 0 表示永久
	MuteDuration string `json:"mute_duration" example:"10"`
	// 赎回金额（解除禁言需要赠送的电池数）, 0 表示不可赎回
	RansomAmount string `json:"ransom_amount" example:"1000"`
	// 回复内容列表
	ReplyContent []string `json:"reply_content" example:"欢迎来到直播间,在哦在哦"`
}
