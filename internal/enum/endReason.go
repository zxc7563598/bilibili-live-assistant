package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type EndReason int

const (
	EndReasonNormal EndReason = iota
	EndReasonForced
	EndReasonBanned
)

func (e EndReason) Key() string {
	switch e {
	case EndReasonNormal:
		return "end_reason.normal"
	case EndReasonForced:
		return "end_reason.forced"
	case EndReasonBanned:
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
	case EndReasonNormal, EndReasonForced, EndReasonBanned:
		return true
	default:
		return false
	}
}
