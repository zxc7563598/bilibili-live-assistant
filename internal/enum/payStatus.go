package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type PayStatus int

const (
	PayStatusUnpaid PayStatus = iota
	PayStatusPaid
	PayStatusRefunded
)

func (p PayStatus) Key() string {
	switch p {
	case PayStatusUnpaid:
		return "pay_status.unpaid"
	case PayStatusPaid:
		return "pay_status.paid"
	case PayStatusRefunded:
		return "pay_status.refunded"
	default:
		return "unknown"
	}
}

func (p PayStatus) Text(lang string) string {
	return i18n.T(lang, p.Key())
}

func (p PayStatus) IsValid() bool {
	switch p {
	case PayStatusUnpaid, PayStatusPaid, PayStatusRefunded:
		return true
	default:
		return false
	}
}
