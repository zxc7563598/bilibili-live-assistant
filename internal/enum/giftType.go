package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type GiftType int

const (
	GiftTypeNormal GiftType = iota
	GiftTypeGuard
	GiftTypeSuperChat
)

func (g GiftType) Key() string {
	switch g {
	case GiftTypeNormal:
		return "gift_type.normal"
	case GiftTypeGuard:
		return "gift_type.guard"
	case GiftTypeSuperChat:
		return "gift_type.super_chat"
	default:
		return "unknown"
	}
}

func (g GiftType) Text(lang string) string {
	return i18n.T(lang, g.Key())
}

func (g GiftType) IsValid() bool {
	switch g {
	case GiftTypeNormal, GiftTypeGuard, GiftTypeSuperChat:
		return true
	default:
		return false
	}
}
