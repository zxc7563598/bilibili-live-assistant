package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type MinMemberLevel int

const (
	MinMemberLevelUnlimited MinMemberLevel = iota
	MinMemberLevelL1
	MinMemberLevelL2
	MinMemberLevelL3
)

func (m MinMemberLevel) Key() string {
	switch m {
	case MinMemberLevelUnlimited:
		return "unlimited"
	case MinMemberLevelL1:
		return "min_member_level.l1"
	case MinMemberLevelL2:
		return "min_member_level.l2"
	case MinMemberLevelL3:
		return "min_member_level.l3"
	default:
		return "unknown"
	}
}

func (m MinMemberLevel) IsValid() bool {
	switch m {
	case MinMemberLevelUnlimited, MinMemberLevelL1, MinMemberLevelL2, MinMemberLevelL3:
		return true
	default:
		return false
	}
}

func (m MinMemberLevel) Text(lang string) string {
	return i18n.T(lang, m.Key())
}

// MinMemberLevelFromBadge 用户身份（BadgeType：1 总督 / 2 提督 / 3 舰长）换算成限购档位口径（1 舰长 / 2 提督 / 3 总督）
func MinMemberLevelFromBadge(badge BadgeType) MinMemberLevel {
	switch badge {
	case BadgeTypeL1:
		return MinMemberLevelL1
	case BadgeTypeL2:
		return MinMemberLevelL2
	case BadgeTypeL3:
		return MinMemberLevelL3
	default:
		return MinMemberLevelUnlimited
	}
}
