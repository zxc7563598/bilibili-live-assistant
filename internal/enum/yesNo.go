package enum

import "github.com/zxc7563598/bilibili-live-assistant/internal/i18n"

type YesNo int

const (
	No YesNo = iota
	Yes
)

func (y YesNo) Key() string {
	switch y {
	case No:
		return "no"
	case Yes:
		return "yes"
	default:
		return "unknown"
	}
}

func (y YesNo) IsValid() bool {
	switch y {
	case No, Yes:
		return true
	default:
		return false
	}
}

func (y YesNo) Text(lang string) string {
	return i18n.T(lang, y.Key())
}

func BoolToYesNo(b bool) YesNo {
	v := No
	if b {
		v = Yes
	}
	return v
}

func BoolToYesNoPtr(b *bool) *YesNo {
	if b == nil {
		return nil
	}
	v := No
	if *b {
		v = Yes
	}
	return &v
}
