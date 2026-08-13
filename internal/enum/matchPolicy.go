package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type MatchPolicy int

const (
	MatchPolicyMatchAny MatchPolicy = iota
	MatchPolicyMatchAll
)

func (m MatchPolicy) Key() string {
	switch m {
	case MatchPolicyMatchAny:
		return "menu_policy_type.match_any"
	case MatchPolicyMatchAll:
		return "menu_policy_type.match_all"
	default:
		return "unknown"
	}
}

func (m MatchPolicy) IsValid() bool {
	switch m {
	case MatchPolicyMatchAny, MatchPolicyMatchAll:
		return true
	default:
		return false
	}
}

func (m MatchPolicy) Text(lang string) string {
	return i18n.T(lang, m.Key())
}
