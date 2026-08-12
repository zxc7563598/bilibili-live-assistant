package resp

// RoomConfigResp 房间模块配置返回
type RoomConfigResp struct {
	// 是否默认监听直播间, 0-否, 1-是
	IsListening string `json:"isListening" example:"1"`
	// 用户名最大长度, 超过此长度则裁剪
	MaxNameLength string `json:"maxNameLength" example:"8"`
	// 裁剪方式, 0-省略后面, 1-省略前面
	NameTrimMode string `json:"nameTrimMode" example:"0"`
}

// SignConfigResp 签到模块配置返回
type SignConfigResp struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" example:"1"`
	// 触发门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" example:"0"`
	// 奖励类型, 0-星光, 1-积分
	RewardType string `json:"reward_type" example:"0"`
	// 奖励数量, 正整数
	RewardAmount string `json:"reward_amount" example:"10"`
	// 签到关键词
	Keyword string `json:"keyword" example:"#签到"`
	// 查询关键词
	QueryKeyword string `json:"query_keyword" example:"#查询"`
	// 签到成功回复
	SuccessReply []string `json:"success_reply"`
	// 签到失败回复
	FailReply []string `json:"fail_reply"`
	// 重复签到回复
	RepeatReply []string `json:"repeat_reply"`
	// 查询成功回复
	QueryReply []string `json:"query_reply"`
}

// AdConfigResp 定时广告模块配置返回
type AdConfigResp struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" example:"1"`
	// 发送间隔, 秒
	Interval string `json:"interval" example:"62"`
	// 发送方式, 0-随机发送, 1-顺序发送
	SendMode string `json:"send_mode" example:"0"`
	// 发送内容
	Content []string `json:"content"`
}

// GiftConfigResp 礼物答谢模块配置返回
type GiftConfigResp struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" example:"1"`
	// 答谢门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" example:"0"`
	// 展示数量, 0-禁用, 1-启用
	ShowCount string `json:"show_count" example:"1"`
	// 礼物合并, 0-禁用, 1-启用
	MergeGift string `json:"merge_gift" example:"1"`
	// 盲盒统计, 0-禁用, 1-启用
	IncludeBlindbox string `json:"include_blindbox" example:"1"`
	// 起始感谢电池数
	MinBattery string `json:"min_battery" example:"10"`
	// 感谢内容
	Content []string `json:"content"`
}

// PkConfigResp PK播报模块配置返回
type PkConfigResp struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" example:"1"`
	// 发送内容
	Content []string `json:"content"`
}

// WelcomeConfigResp 进房欢迎模块配置返回
type WelcomeConfigResp struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" example:"1"`
	// 欢迎门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" example:"0"`
	// 欢迎内容
	Content []string `json:"content"`
}

// FollowConfigResp 感谢关注模块配置返回
type FollowConfigResp struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" example:"1"`
	// 感谢门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" example:"0"`
	// 感谢内容
	Content []string `json:"content"`
}

// ShareConfigResp 感谢分享模块配置返回
type ShareConfigResp struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" example:"1"`
	// 感谢门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" example:"0"`
	// 感谢内容
	Content []string `json:"content"`
}

// ReplyConfigResp 自动回复模块配置返回
type ReplyConfigResp struct {
	// 是否启用, 0-禁用, 1-启用
	Enabled string `json:"enabled" example:"1"`
	// 可用场景, 0-不限制, 1-直播中, 2-非直播中
	Scene string `json:"scene" example:"1"`
	// 触发门槛, 0-不限制, 1-带本直播间牌子, 2-带本直播间大航海牌子
	Requirement string `json:"requirement" example:"0"`
	// 回复内容
	Content []ReplyItem `json:"content"`
}

// ReplyItem 自动回复规则项返回
type ReplyItem struct {
	// 触发关键词列表
	Keyword []string `json:"keyword"`
	// 安全词列表
	SafeWord []string `json:"safe_word"`
	// 是否禁言发送者, 0-否, 1-是
	MuteSender string `json:"mute_sender"`
	// 禁言时长（分钟）, 0 表示永久
	MuteDuration string `json:"mute_duration"`
	// 赎回金额（解除禁言需要赠送的电池数）, 0 表示不可赎回
	RansomAmount string `json:"ransom_amount"`
	// 回复内容列表
	ReplyContent []string `json:"reply_content"`
}
