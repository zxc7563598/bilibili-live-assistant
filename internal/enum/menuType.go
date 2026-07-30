package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type MenuType string

const (
	MenuTypeButton MenuType = "BUTTON"
	MenuTypeMenu   MenuType = "MENU"
)

func (m MenuType) Key() string {
	switch m {
	case MenuTypeButton:
		return "menu_type.button"
	case MenuTypeMenu:
		return "menu_type.menu"
	default:
		return "unknown"
	}
}

func (m MenuType) IsValid() bool {
	switch m {
	case MenuTypeButton, MenuTypeMenu:
		return true
	default:
		return false
	}
}

func (m MenuType) Text(lang string) string {
	return i18n.T(lang, m.Key())
}
