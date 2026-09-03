package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type ProductType int

const (
	ProductTypeVirtual ProductType = iota
	ProductTypeActual
)

func (p ProductType) Key() string {
	switch p {
	case ProductTypeVirtual:
		return "product_type.virtual"
	case ProductTypeActual:
		return "product_type.actual"
	default:
		return "unknown"
	}
}

func (p ProductType) Text(lang string) string {
	return i18n.T(lang, p.Key())
}

func (p ProductType) IsValid() bool {
	switch p {
	case ProductTypeVirtual, ProductTypeActual:
		return true
	default:
		return false
	}
}
