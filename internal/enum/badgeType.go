package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type BadgeType int

const (
	BadgeTypeL0 BadgeType = iota
	BadgeTypeL3
	BadgeTypeL2
	BadgeTypeL1
)

func (b BadgeType) Key() string {
	switch b {
	case BadgeTypeL0:
		return "badge_type.l0"
	case BadgeTypeL1:
		return "badge_type.l1"
	case BadgeTypeL2:
		return "badge_type.l2"
	case BadgeTypeL3:
		return "badge_type.l3"
	default:
		return "unknown"
	}
}

func (b BadgeType) Text(lang string) string {
	return i18n.T(lang, b.Key())
}

func (b BadgeType) IsValid() bool {
	switch b {
	case BadgeTypeL0, BadgeTypeL1, BadgeTypeL2, BadgeTypeL3:
		return true
	default:
		return false
	}
}
