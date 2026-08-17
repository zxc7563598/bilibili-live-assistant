package robotconfig

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

// GetRoomConfig 请求返回
type RoomConfigResp struct {
	IsListening           string `json:"is_listening"`
	MaxNameLength         string `json:"max_name_length"`
	NameTrimMode          string `json:"name_trim_mode"`
	ConsumeRewardEnabled  string `json:"consume_reward_enabled"`
	RewardType            string `json:"reward_type"`
	ConsumeBatteryRate    string `json:"consume_battery_rate"`
	CaptainRewardAmount   string `json:"captain_reward_amount"`
	CommanderRewardAmount string `json:"commander_reward_amount"`
	GovernorRewardAmount  string `json:"governor_reward_amount"`
}

// GetSignConfig 请求返回
type SignConfigResp struct {
	Enabled      string   `json:"enabled"`
	Scene        string   `json:"scene"`
	Requirement  string   `json:"requirement"`
	RewardType   string   `json:"reward_type"`
	RewardAmount string   `json:"reward_amount"`
	Keyword      string   `json:"keyword"`
	QueryKeyword string   `json:"query_keyword"`
	SuccessReply []string `json:"success_reply"`
	FailReply    []string `json:"fail_reply"`
	RepeatReply  []string `json:"repeat_reply"`
	QueryReply   []string `json:"query_reply"`
}

// GetAdConfig 请求返回
type AdConfigResp struct {
	Enabled  string   `json:"enabled"`
	Scene    string   `json:"scene"`
	Interval string   `json:"interval"`
	SendMode string   `json:"send_mode"`
	Content  []string `json:"content"`
}

// GetGiftConfig 请求返回
type GiftConfigResp struct {
	Enabled         string   `json:"enabled"`
	Scene           string   `json:"scene"`
	Requirement     string   `json:"requirement"`
	ShowCount       string   `json:"show_count"`
	MergeGift       string   `json:"merge_gift"`
	IncludeBlindbox string   `json:"include_blindbox"`
	MinBattery      string   `json:"min_battery"`
	Content         []string `json:"content"`
}

// GetPkConfig 请求返回
type PkConfigResp struct {
	Enabled string   `json:"enabled"`
	Content []string `json:"content"`
}

// GetWelcomeConfig 请求返回
type WelcomeConfigResp struct {
	Enabled     string   `json:"enabled"`
	Scene       string   `json:"scene"`
	Requirement string   `json:"requirement"`
	Content     []string `json:"content"`
}

// GetFollowConfig 请求返回
type FollowConfigResp struct {
	Enabled     string   `json:"enabled"`
	Scene       string   `json:"scene"`
	Requirement string   `json:"requirement"`
	Content     []string `json:"content"`
}

// GetShareConfig 请求返回
type ShareConfigResp struct {
	Enabled     string   `json:"enabled"`
	Scene       string   `json:"scene"`
	Requirement string   `json:"requirement"`
	Content     []string `json:"content"`
}

// GetReplyConfig 请求返回
type ReplyConfigResp struct {
	Enabled     string      `json:"enabled"`
	Scene       string      `json:"scene"`
	Requirement string      `json:"requirement"`
	Content     []ReplyItem `json:"content"`
}

// ApplyRoomConfig 请求入参
type RoomConfigReq struct {
	IsListening           string `json:"is_listening"`
	MaxNameLength         string `json:"max_name_length"`
	NameTrimMode          string `json:"name_trim_mode"`
	ConsumeRewardEnabled  string `json:"consume_reward_enabled"`
	RewardType            string `json:"reward_type"`
	ConsumeBatteryRate    string `json:"consume_battery_rate"`
	CaptainRewardAmount   string `json:"captain_reward_amount"`
	CommanderRewardAmount string `json:"commander_reward_amount"`
	GovernorRewardAmount  string `json:"governor_reward_amount"`
}

// ApplySignConfig 请求入参
type SignConfigReq struct {
	Enabled      string   `json:"enabled"`
	Scene        string   `json:"scene"`
	Requirement  string   `json:"requirement"`
	RewardType   string   `json:"reward_type"`
	RewardAmount string   `json:"reward_amount"`
	Keyword      string   `json:"keyword"`
	QueryKeyword string   `json:"query_keyword"`
	SuccessReply []string `json:"success_reply"`
	FailReply    []string `json:"fail_reply"`
	RepeatReply  []string `json:"repeat_reply"`
	QueryReply   []string `json:"query_reply"`
}

// ApplyAdConfig 请求入参
type AdConfigReq struct {
	Enabled  string   `json:"enabled"`
	Scene    string   `json:"scene"`
	Interval string   `json:"interval"`
	SendMode string   `json:"send_mode"`
	Content  []string `json:"content"`
}

// ApplyGiftConfig 请求入参
type GiftConfigReq struct {
	Enabled         string   `json:"enabled"`
	Scene           string   `json:"scene"`
	Requirement     string   `json:"requirement"`
	ShowCount       string   `json:"show_count"`
	MergeGift       string   `json:"merge_gift"`
	IncludeBlindbox string   `json:"include_blindbox"`
	MinBattery      string   `json:"min_battery"`
	Content         []string `json:"content"`
}

// ApplyPkConfig 请求入参
type PkConfigReq struct {
	Enabled string   `json:"enabled"`
	Content []string `json:"content"`
}

// ApplyWelcomeConfig 请求入参
type WelcomeConfigReq struct {
	Enabled     string   `json:"enabled"`
	Scene       string   `json:"scene"`
	Requirement string   `json:"requirement"`
	Content     []string `json:"content"`
}

// ApplyFollowConfig 请求入参
type FollowConfigReq struct {
	Enabled     string   `json:"enabled"`
	Scene       string   `json:"scene"`
	Requirement string   `json:"requirement"`
	Content     []string `json:"content"`
}

// ApplyShareConfig 请求入参
type ShareConfigReq struct {
	Enabled     string   `json:"enabled"`
	Scene       string   `json:"scene"`
	Requirement string   `json:"requirement"`
	Content     []string `json:"content"`
}

// ApplyReplyConfig 请求入参
type ReplyConfigReq struct {
	Enabled     string      `json:"enabled"`
	Scene       string      `json:"scene"`
	Requirement string      `json:"requirement"`
	Content     []ReplyItem `json:"content"`
}
