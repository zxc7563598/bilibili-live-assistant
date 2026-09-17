package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type OrderStatus int

const (
	OrderStatusPendingPayment OrderStatus = iota
	OrderStatusPendingShipment
	OrderStatusPendingReceipt
	OrderStatusCompleted
	OrderStatusCancelled
	OrderStatusAfterSales
)

func (o OrderStatus) Key() string {
	switch o {
	case OrderStatusPendingPayment:
		return "order_status.pending_payment"
	case OrderStatusPendingShipment:
		return "order_status.pending_shipment"
	case OrderStatusPendingReceipt:
		return "order_status.pending_receipt"
	case OrderStatusCompleted:
		return "order_status.completed"
	case OrderStatusCancelled:
		return "order_status.cancelled"
	case OrderStatusAfterSales:
		return "order_status.after_sales"
	default:
		return "unknown"
	}
}

func (o OrderStatus) Text(lang string) string {
	return i18n.T(lang, o.Key())
}

func (o OrderStatus) IsValid() bool {
	switch o {
	case OrderStatusPendingPayment, OrderStatusPendingShipment, OrderStatusPendingReceipt, OrderStatusCompleted, OrderStatusCancelled, OrderStatusAfterSales:
		return true
	default:
		return false
	}
}
