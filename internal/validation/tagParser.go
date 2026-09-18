package validation

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
)

// parseFieldError 将验证错误映射为自定义的错误码
func parseFieldError(e validator.FieldError, field reflect.StructField) int {
	errTag := field.Tag.Get("err")
	// err:"required=10104,min=10105"
	parts := strings.Split(errTag, ",")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}
		if kv[0] == e.Tag() {
			code, err := strconv.Atoi(kv[1])
			if err == nil {
				return code
			}
		}
	}
	return i18n.CodeParamInvalid
}
