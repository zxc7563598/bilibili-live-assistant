package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type ImportStatus int

const (
	ImportStatusPending ImportStatus = iota
	ImportStatusRunning
	ImportStatusFinished
	ImportStatusFailed
)

func (i ImportStatus) Key() string {
	switch i {
	case ImportStatusPending:
		return "import_status.pending"
	case ImportStatusRunning:
		return "import_status.running"
	case ImportStatusFinished:
		return "import_status.finished"
	case ImportStatusFailed:
		return "import_status.failed"
	default:
		return "unknown"
	}
}

func (i ImportStatus) Text(lang string) string {
	return i18n.T(lang, i.Key())
}

func (i ImportStatus) IsValid() bool {
	switch i {
	case ImportStatusPending, ImportStatusRunning, ImportStatusFinished, ImportStatusFailed:
		return true
	default:
		return false
	}
}
