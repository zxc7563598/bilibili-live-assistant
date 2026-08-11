package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type OperatorType int

const (
	OperatorTypeUser OperatorType = iota
	OperatorTypeSystem
	OperatorTypeAdmin
)

func (o OperatorType) Key() string {
	switch o {
	case OperatorTypeUser:
		return "operator_type.user"
	case OperatorTypeSystem:
		return "operator_type.system"
	case OperatorTypeAdmin:
		return "operator_type.admin"
	default:
		return "unknown"
	}
}

func (o OperatorType) Text(lang string) string {
	return i18n.T(lang, o.Key())
}

func (o OperatorType) IsValid() bool {
	switch o {
	case OperatorTypeUser, OperatorTypeSystem, OperatorTypeAdmin:
		return true
	default:
		return false
	}
}
