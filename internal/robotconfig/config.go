package robotconfig

// SignConfig 签到配置
type SignConfig struct {
	Enabled      string   `config:"enabled"`       // 是否启用
	Keyword      string   `config:"keyword"`       // 签到关键词
	QueryKeyword string   `config:"query_keyword"` // 查询关键词
	Requirement  string   `config:"requirement"`   // 签到条件
	RewardAmount string   `config:"reward_amount"` // 奖励数量
	RewardType   string   `config:"reward_type"`   // 奖励类型
	Scene        string   `config:"scene"`         // 场景
	QueryReply   []string `config:"query_reply"`   // 查询回复
	RepeatReply  []string `config:"repeat_reply"`  // 重复签到回复
	FailReply    []string `config:"fail_reply"`    // 签到失败回复
	SuccessReply []string `config:"success_reply"` // 签到成功回复
}

// ReplyConfig 自动回复配置
type ReplyConfig struct {
	Enabled     string      `config:"enabled"`     // 是否启用
	Scene       string      `config:"scene"`       // 场景
	Requirement string      `config:"requirement"` // 回复条件
	Content     []ReplyItem `config:"content"`     // 配置内容
}

type ReplyItem struct {
	Keyword             []string `json:"keyword"`                // 触发自动回复的关键词列表
	KeywordMatchPolicy  string   `json:"keyword_match_policy"`   // 关键词匹配策略
	SafeWord            []string `json:"safe_word"`              // 包含此词则不触发回复
	SafeWordMatchPolicy string   `json:"safe_word_match_policy"` // 安全词匹配策略
	MuteSender          string   `json:"mute_sender"`            // 是否禁言, 0-否, 1-是
	MuteDuration        string   `json:"mute_duration"`          // 禁言时长（分钟）, 0 表示永久
	RansomAmount        string   `json:"ransom_amount"`          // 赎回金额（解除禁言需要赠送的电池数）, 0 表示不可赎回
	ReplyContent        []string `json:"reply_content"`          // 回复内容
}

// GiftConfig 礼物答谢回复配置
type GiftConfig struct {
	Enabled         string   `config:"enabled"`          // 是否启用
	Scene           string   `config:"scene"`            // 场景
	Requirement     string   `config:"requirement"`      // 签到条件
	ShowCount       string   `config:"show_count"`       // 展示数量
	MergeGift       string   `config:"merge_gift"`       // 礼物合并
	IncludeBlindbox string   `config:"include_blindbox"` // 盲盒统计
	MinBattery      string   `config:"min_battery"`      // 起始感谢电池
	Content         []string `config:"content"`          // 答谢内容
}

// InteractConfig 互动消息（进房欢迎/感谢关注/感谢分享）回复配置
type InteractConfig struct {
	Enabled     string   `config:"enabled"`     // 是否启用
	Scene       string   `config:"scene"`       // 场景
	Requirement string   `config:"requirement"` // 条件
	Content     []string `config:"content"`     // 内容
}

// PkConfig PK信息
type PkConfig struct {
	Enabled string   `config:"enabled"` // 是否启用
	Content []string `config:"content"` // 内容
}

// RoomConfig 房间配置
type RoomConfig struct {
	MaxNameLength         string `config:"max_name_length"`         // 用户名最大长度, 超过此长度则裁剪
	NameTrimMode          string `config:"name_trim_mode"`          // 裁剪方式, 0-省略后面, 1-省略前面
	ConsumeRewardEnabled  string `config:"consume_reward_enabled"`  // 用户消费发放奖励, 0-不发放, 1-按消费电池发放, 2-按开通航海类型发放
	RewardType            string `config:"reward_type"`             // 奖励类型, 0-星光, 1-积分
	ConsumeBatteryRate    string `config:"consume_battery_rate"`    // 消费电池转换倍率, 设置为 2 则代表用户消耗 1 电池会得到 2 点奖励, 奖励设置为按消费电池发放时生效
	CaptainRewardAmount   string `config:"captain_reward_amount"`   // 开通舰长奖励数量, 奖励设置为按开通航海类型发放时生效
	CommanderRewardAmount string `config:"commander_reward_amount"` // 开通提督奖励数量, 奖励设置为按开通航海类型发放时生效
	GovernorRewardAmount  string `config:"governor_reward_amount"`  // 开通总督奖励数量, 奖励设置为按开通航海类型发放时生效
}
