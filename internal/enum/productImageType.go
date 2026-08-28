package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type ProductImageType int

const (
	ProductImageTypeBanner ProductImageType = iota
	ProductImageTypeDetail
)

func (p ProductImageType) Key() string {
	switch p {
	case ProductImageTypeBanner:
		return "product_image_type.banner"
	case ProductImageTypeDetail:
		return "product_image_type.detail"
	default:
		return "unknown"
	}
}

func (p ProductImageType) Text(lang string) string {
	return i18n.T(lang, p.Key())
}

func (p ProductImageType) IsValid() bool {
	switch p {
	case ProductImageTypeBanner, ProductImageTypeDetail:
		return true
	default:
		return false
	}
}
