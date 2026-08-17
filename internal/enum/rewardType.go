package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type RewardType int

const (
	RewardTypeStars RewardType = iota
	RewardTypePoints
)

func (r RewardType) Key() string {
	switch r {
	case RewardTypeStars:
		return "reward_type.stars"
	case RewardTypePoints:
		return "reward_type.points"
	default:
		return "unknown"
	}
}

func (r RewardType) Text(lang string) string {
	return i18n.T(lang, r.Key())
}

func (r RewardType) IsValid() bool {
	switch r {
	case RewardTypeStars, RewardTypePoints:
		return true
	default:
		return false
	}
}
