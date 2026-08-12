package robotconfig

// SignConfig 签到配置
type SignConfig struct {
	Enabled      string `config:"enabled"`       // 是否启用
	Keyword      string `config:"keyword"`       // 签到关键词
	QueryKeyword string `config:"query_keyword"` // 查询关键词
	Requirement  string `config:"requirement"`   // 签到条件
	RewardAmount string `config:"reward_amount"` // 奖励数量
	RewardType   string `config:"reward_type"`   // 奖励类型
	Scene        string `config:"scene"`         // 场景
	QueryReply   []string `config:"query_reply"`   // 查询回复
	RepeatReply  []string `config:"repeat_reply"`  // 重复签到回复
	FailReply    []string `config:"fail_reply"`    // 签到失败回复
	SuccessReply []string `config:"success_reply"` // 签到成功回复
}
