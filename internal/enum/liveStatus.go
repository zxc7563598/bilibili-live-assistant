package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type LiveStatus int

const (
	LiveStatusOffline LiveStatus = iota
	LiveStatusLive
	LiveStatusLooping
)

func (l LiveStatus) Key() string {
	switch l {
	case LiveStatusOffline:
		return "live_status.offline"
	case LiveStatusLive:
		return "live_status.live"
	case LiveStatusLooping:
		return "live_status.looping"
	default:
		return "unknown"
	}
}

func (l LiveStatus) Text(lang string) string {
	return i18n.T(lang, l.Key())
}

func (l LiveStatus) IsValid() bool {
	switch l {
	case LiveStatusOffline, LiveStatusLive, LiveStatusLooping:
		return true
	default:
		return false
	}
}
