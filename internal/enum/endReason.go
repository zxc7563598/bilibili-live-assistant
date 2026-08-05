package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type EndReason int

const (
	Normal EndReason = iota
	Forced
	Banned
)

func (e EndReason) Key() string {
	switch e {
	case Normal:
		return "end_reason.normal"
	case Forced:
		return "end_reason.forced"
	case Banned:
		return "end_reason.banned"
	default:
		return "unknown"
	}
}

func (e EndReason) Text(lang string) string {
	return i18n.T(lang, e.Key())
}

func (e EndReason) IsValid() bool {
	switch e {
	case Normal, Forced, Banned:
		return true
	default:
		return false
	}
}
