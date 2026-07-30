package enum

import "github.com/zxc7563598/GoAdminKit/internal/i18n"

type Gender int

const (
	GenderUnknown Gender = iota
	GenderMale
	GenderFemale
)

func (g Gender) Key() string {
	switch g {
	case GenderFemale:
		return "gender.female"
	case GenderMale:
		return "gender.male"
	default:
		return "unknown"
	}
}

func (g Gender) Text(lang string) string {
	return i18n.T(lang, g.Key())
}

func (g Gender) IsValid() bool {
	switch g {
	case GenderUnknown, GenderMale, GenderFemale:
		return true
	default:
		return false
	}
}
