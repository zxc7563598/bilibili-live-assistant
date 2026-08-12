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
	Keyword      []string `json:"keyword"`       // 触发自动回复的关键词列表
	SafeWord     []string `json:"safe_word"`     // 包含此词则不触发回复
	MuteSender   string   `json:"mute_sender"`   // 是否禁言, 0-否, 1-是
	MuteDuration string   `json:"mute_duration"` // 禁言时长（分钟）, 0 表示永久
	RansomAmount string   `json:"ransom_amount"` // 赎回金额（解除禁言需要赠送的电池数）, 0 表示不可赎回
	ReplyContent []string `json:"reply_content"` // 回复内容
}
