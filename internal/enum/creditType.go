package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type CreditType int

const (
	CreditTypeStars CreditType = iota
	CreditTypePoints
)

func (c CreditType) Key() string {
	switch c {
	case CreditTypeStars:
		return "credit_type.stars"
	case CreditTypePoints:
		return "credit_type.points"
	default:
		return "unknown"
	}
}

func (c CreditType) Text(lang string) string {
	return i18n.T(lang, c.Key())
}

func (c CreditType) IsValid() bool {
	switch c {
	case CreditTypeStars, CreditTypePoints:
		return true
	default:
		return false
	}
}
