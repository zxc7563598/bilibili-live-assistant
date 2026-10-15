package enum

// 大航海礼物名称，B站 固定的三个名字
const (
	GuardNameCaptain  = "舰长"
	GuardNameAdmiral  = "提督"
	GuardNameGovernor = "总督"
)

// BadgeTypeByGuardName 大航海礼物名称 → 等级，非航海名称返回 false
//
// 航海礼物记录不写 badge_type，礼物名称是唯一的档位来源。
func BadgeTypeByGuardName(name string) (BadgeType, bool) {
	switch name {
	case GuardNameCaptain:
		return BadgeTypeL1, true
	case GuardNameAdmiral:
		return BadgeTypeL2, true
	case GuardNameGovernor:
		return BadgeTypeL3, true
	default:
		return BadgeTypeL0, false
	}
}

// GuardNameByBadgeType 大航海等级 → 礼物名称，非航海等级返回 false
func GuardNameByBadgeType(level BadgeType) (string, bool) {
	switch level {
	case BadgeTypeL1:
		return GuardNameCaptain, true
	case BadgeTypeL2:
		return GuardNameAdmiral, true
	case BadgeTypeL3:
		return GuardNameGovernor, true
	default:
		return "", false
	}
}
