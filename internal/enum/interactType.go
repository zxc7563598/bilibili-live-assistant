package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type InteractType int64

const (
	InteractTypeEnter InteractType = iota + 1
	InteractTypeFollow
	InteractTypeShare
)

func (t InteractType) Key() string {
	switch t {
	case InteractTypeEnter:
		return "interact_type.enter"
	case InteractTypeFollow:
		return "interact_type.follow"
	case InteractTypeShare:
		return "interact_type.share"
	default:
		return "unknown"
	}
}

func (t InteractType) Text(lang string) string {
	return i18n.T(lang, t.Key())
}

func (t InteractType) IsValid() bool {
	switch t {
	case InteractTypeEnter, InteractTypeFollow, InteractTypeShare:
		return true
	default:
		return false
	}
}
