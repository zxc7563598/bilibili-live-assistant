package validation

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// Parse 将 validator 的验证错误解析为对应的错误码
// 如果无法解析则返回默认错误码 10001。
func Parse(err error, req any) int {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return 10001
	}
	e := ve[0]
	field, ok := getStructField(req, e.Field())
	if !ok {
		return 10001
	}
	return parseFieldError(e, field)
}
