package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type MuteStatus int

const (
	MuteStatusMuted MuteStatus = iota
	MuteStatusUnmuteFailed
	MuteStatusNotFound
	MuteStatusUnmuted
)

func (m MuteStatus) Key() string {
	switch m {
	case MuteStatusMuted:
		return "mute_status.muted"
	case MuteStatusUnmuteFailed:
		return "mute_status.unmute_failed"
	case MuteStatusNotFound:
		return "mute_status.not_found"
	case MuteStatusUnmuted:
		return "mute_status.unmuted"
	default:
		return "unknown"
	}
}

func (m MuteStatus) IsValid() bool {
	switch m {
	case MuteStatusMuted, MuteStatusUnmuteFailed, MuteStatusNotFound, MuteStatusUnmuted:
		return true
	default:
		return false
	}
}

func (m MuteStatus) Text(lang string) string {
	return i18n.T(lang, m.Key())
}
