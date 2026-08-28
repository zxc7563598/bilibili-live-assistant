package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type StockChangeType int

const (
	StockChangeTypeOrderDeduct StockChangeType = iota
	StockChangeTypeReturnRefund
	StockChangeTypeAdminAdd
	StockChangeTypeAdminReduce
)

func (s StockChangeType) Key() string {
	switch s {
	case StockChangeTypeOrderDeduct:
		return "stock_change_type.order_deduct"
	case StockChangeTypeReturnRefund:
		return "stock_change_type.return_refund"
	case StockChangeTypeAdminAdd:
		return "stock_change_type.admin_add"
	case StockChangeTypeAdminReduce:
		return "stock_change_type.admin_reduce"
	default:
		return "unknown"
	}
}

func (s StockChangeType) Text(lang string) string {
	return i18n.T(lang, s.Key())
}

func (s StockChangeType) IsValid() bool {
	switch s {
	case StockChangeTypeOrderDeduct, StockChangeTypeReturnRefund, StockChangeTypeAdminAdd, StockChangeTypeAdminReduce:
		return true
	default:
		return false
	}
}
