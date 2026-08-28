package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type ShipStatus int

const (
	ShipStatusPending ShipStatus = iota
	ShipStatusShipped
	ShipStatusDelivered
)

func (s ShipStatus) Key() string {
	switch s {
	case ShipStatusPending:
		return "ship_status.pending"
	case ShipStatusShipped:
		return "ship_status.shipped"
	case ShipStatusDelivered:
		return "ship_status.delivered"
	default:
		return "unknown"
	}
}

func (s ShipStatus) Text(lang string) string {
	return i18n.T(lang, s.Key())
}

func (s ShipStatus) IsValid() bool {
	switch s {
	case ShipStatusPending, ShipStatusShipped, ShipStatusDelivered:
		return true
	default:
		return false
	}
}
