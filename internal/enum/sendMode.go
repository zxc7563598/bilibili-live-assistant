package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type SendMode int

const (
	SendModeRandom SendMode = iota
	SendModeSequential
)

func (s SendMode) Key() string {
	switch s {
	case SendModeRandom:
		return "send_mode.random"
	case SendModeSequential:
		return "send_mode.sequential"
	default:
		return "unknown"
	}
}

func (s SendMode) Text(lang string) string {
	return i18n.T(lang, s.Key())
}

func (s SendMode) IsValid() bool {
	switch s {
	case SendModeRandom, SendModeSequential:
		return true
	default:
		return false
	}
}
