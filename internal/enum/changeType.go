package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type ChangeType int

const (
	ChangeTypeReduce ChangeType = iota
	ChangeTypeIncrease
)

func (c ChangeType) Key() string {
	switch c {
	case ChangeTypeReduce:
		return "change_type.reduce"
	case ChangeTypeIncrease:
		return "change_type.increase"
	default:
		return "unknown"
	}
}

func (c ChangeType) Text(lang string) string {
	return i18n.T(lang, c.Key())
}

func (c ChangeType) IsValid() bool {
	switch c {
	case ChangeTypeReduce, ChangeTypeIncrease:
		return true
	default:
		return false
	}
}
