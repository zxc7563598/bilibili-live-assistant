package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type Requirement int

const (
	RequirementUnlimited Requirement = iota
	RequirementHasBadge
	RequirementHasSailBadge
)

func (r Requirement) Key() string {
	switch r {
	case RequirementUnlimited:
		return "unlimited"
	case RequirementHasBadge:
		return "requirement.has_badge"
	case RequirementHasSailBadge:
		return "requirement.has_sail_badge"
	default:
		return "unknown"
	}
}

func (r Requirement) Text(lang string) string {
	return i18n.T(lang, r.Key())
}

func (r Requirement) IsValid() bool {
	switch r {
	case RequirementUnlimited, RequirementHasBadge, RequirementHasSailBadge:
		return true
	default:
		return false
	}
}
