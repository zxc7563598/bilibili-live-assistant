package validation

import "reflect"

// getStructField 从结构体或结构体指针中获取指定字段的反射信息
func getStructField(req any, name string) (reflect.StructField, bool) {
	t := reflect.TypeOf(req)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	field, ok := t.FieldByName(name)
	return field, ok
}
