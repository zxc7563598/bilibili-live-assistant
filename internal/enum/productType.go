package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type ProductType int

const (
	// 商品类型与收货地址类型（AddressType）语义一致（0 虚拟 / 1 实体）。
	// 商城端会把 product_type 直接当地址类型查询（确认页取对应类型的地址），
	// 因此数值显式引用 AddressType 常量，避免两条独立 iota 枚举日后漂移导致类型错配。
	ProductTypeVirtual ProductType = ProductType(AddressTypeVirtual)
	ProductTypeActual  ProductType = ProductType(AddressTypeActual)
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
