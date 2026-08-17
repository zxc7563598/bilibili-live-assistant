package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type NameTrimMode int

const (
	NameTrimModeTrimEnd NameTrimMode = iota
	NameTrimModeTrimStart
)

func (n NameTrimMode) Key() string {
	switch n {
	case NameTrimModeTrimEnd:
		return "name_trim_mode.trim_end"
	case NameTrimModeTrimStart:
		return "name_trim_mode.trim_start"
	default:
		return "unknown"
	}
}

func (n NameTrimMode) IsValid() bool {
	switch n {
	case NameTrimModeTrimEnd, NameTrimModeTrimStart:
		return true
	default:
		return false
	}
}

func (n NameTrimMode) Text(lang string) string {
	return i18n.T(lang, n.Key())
}
