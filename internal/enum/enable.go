package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type Enable int

const (
	EnableDisable Enable = iota
	EnableEnable
)

func (e Enable) Key() string {
	switch e {
	case EnableDisable:
		return "enable.disable"
	case EnableEnable:
		return "enable.enable"
	default:
		return "unknown"
	}
}

func (e Enable) Text(lang string) string {
	return i18n.T(lang, e.Key())
}

func (e Enable) IsValid() bool {
	switch e {
	case EnableDisable, EnableEnable:
		return true
	default:
		return false
	}
}

func BoolToEnable(b bool) Enable {
	v := EnableDisable
	if b {
		v = EnableEnable
	}
	return v
}

func BoolToEnablePtr(b *bool) *Enable {
	if b == nil {
		return nil
	}
	v := EnableDisable
	if *b {
		v = EnableEnable
	}
	return &v
}
