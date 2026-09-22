package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type PkStatus int

// PK 状态，取值来自 B 站协议（非连续，所以不能直接用 iota）
const (
	PkStatusStartingSoon PkStatus = 101
	PkStatusCompleted    PkStatus = 401
	PkStatusAborted      PkStatus = 404
)

func (p PkStatus) Key() string {
	switch p {
	case PkStatusStartingSoon:
		return "pk_status.starting_soon"
	case PkStatusCompleted:
		return "pk_status.completed"
	case PkStatusAborted:
		return "pk_status.aborted"
	default:
		return "unknown"
	}
}

func (p PkStatus) Text(lang string) string {
	return i18n.T(lang, p.Key())
}

func (p PkStatus) IsValid() bool {
	switch p {
	case PkStatusStartingSoon, PkStatusCompleted, PkStatusAborted:
		return true
	default:
		return false
	}
}
