package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type AddressType int

const (
	AddressTypeVirtual AddressType = iota
	AddressTypeActual
)

func (a AddressType) Key() string {
	switch a {
	case AddressTypeVirtual:
		return "address_type.virtual"
	case AddressTypeActual:
		return "address_type.actual"
	default:
		return "unknown"
	}
}

func (a AddressType) Text(lang string) string {
	return i18n.T(lang, a.Key())
}

func (a AddressType) IsValid() bool {
	switch a {
	case AddressTypeVirtual, AddressTypeActual:
		return true
	default:
		return false
	}
}
