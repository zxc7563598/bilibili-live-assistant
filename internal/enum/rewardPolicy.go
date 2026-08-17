package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type RewardPolicy int

const (
	RewardPolicyNoReward RewardPolicy = iota
	RewardPolicyBatteryReward
	RewardPolicyVipReward
)

func (r RewardPolicy) Key() string {
	switch r {
	case RewardPolicyNoReward:
		return "reward_policy.no_reward"
	case RewardPolicyBatteryReward:
		return "reward_policy.battery_reward"
	case RewardPolicyVipReward:
		return "reward_policy.vip_reward"
	default:
		return "unknown"
	}
}

func (r RewardPolicy) Text(lang string) string {
	return i18n.T(lang, r.Key())
}

func (r RewardPolicy) IsValid() bool {
	switch r {
	case RewardPolicyNoReward, RewardPolicyBatteryReward, RewardPolicyVipReward:
		return true
	default:
		return false
	}
}
