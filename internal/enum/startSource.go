package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type StartSource int

const (
	StartSourceEvent StartSource = iota
	StartSourcePolling
	StartSourceManual
)

func (s StartSource) Key() string {
	switch s {
	case StartSourceEvent:
		return "start_source.event"
	case StartSourcePolling:
		return "start_source.polling"
	case StartSourceManual:
		return "start_source.manual"
	default:
		return "unknown"
	}
}

func (s StartSource) Text(lang string) string {
	return i18n.T(lang, s.Key())
}

func (s StartSource) IsValid() bool {
	switch s {
	case StartSourceEvent, StartSourcePolling, StartSourceManual:
		return true
	default:
		return false
	}
}
