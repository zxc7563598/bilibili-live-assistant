package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type DraftStatus int

const (
	DraftStatusActive DraftStatus = iota
	DraftStatusRedeemed
	DraftStatusCancelled
)

func (d DraftStatus) Key() string {
	switch d {
	case DraftStatusActive:
		return "draft_status.active"
	case DraftStatusRedeemed:
		return "draft_status.redeemed"
	case DraftStatusCancelled:
		return "draft_status.cancelled"
	default:
		return "unknown"
	}
}

func (d DraftStatus) Text(lang string) string {
	return i18n.T(lang, d.Key())
}

func (d DraftStatus) IsValid() bool {
	switch d {
	case DraftStatusActive, DraftStatusRedeemed, DraftStatusCancelled:
		return true
	default:
		return false
	}
}
