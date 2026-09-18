package validation

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
)

// Parse 将 validator 的验证错误解析为对应的错误码
// 如果无法解析则返回默认错误码 i18n.CodeParamInvalid。
func Parse(err error, req any) int {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return i18n.CodeParamInvalid
	}
	e := ve[0]
	field, ok := getStructField(req, e.Field())
	if !ok {
		return i18n.CodeParamInvalid
	}
	return parseFieldError(e, field)
}
