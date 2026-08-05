package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type EndSource int

const (
	EndSourceEvent EndSource = iota
	EndSourcePolling
	EndSourceManual
)

func (s EndSource) Key() string {
	switch s {
	case EndSourceEvent:
		return "end_source.event"
	case EndSourcePolling:
		return "end_source.polling"
	case EndSourceManual:
		return "end_source.manual"
	default:
		return "unknown"
	}
}

func (s EndSource) Text(lang string) string {
	return i18n.T(lang, s.Key())
}

func (s EndSource) IsValid() bool {
	switch s {
	case EndSourceEvent, EndSourcePolling, EndSourceManual:
		return true
	default:
		return false
	}
}
