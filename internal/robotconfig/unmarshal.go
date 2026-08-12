package robotconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// UnmarshalGroup 将指定分组的配置解析到结构体指针。
// 结构体字段需使用 `config:"key_name"` tag 指定映射的配置键名。
// 未配置 tag 的字段会被跳过，配置中不存在的 key 会保留字段零值。
func (c *Cache) UnmarshalGroup(groupName string, v any) error {
	group, ok := c.GetGroup(groupName)
	if !ok {
		return fmt.Errorf("配置分组 %s 不存在", groupName)
	}
	return unmarshalConfig(group, v)
}

func unmarshalConfig(m map[string]string, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v 必须是非空的指针")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errors.New("v 必须是指向结构体的指针")
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("config")
		if tag == "" {
			continue
		}
		val, ok := m[tag]
		if !ok {
			continue
		}
		fv := rv.Field(i)
		if !fv.CanSet() {
			continue
		}
		if fv.Kind() == reflect.String {
			fv.SetString(val)
		}
		if fv.Kind() == reflect.Slice && fv.Type().Elem().Kind() == reflect.String {
			var s []string
			if err := json.Unmarshal([]byte(val), &s); err != nil {
				return fmt.Errorf("解析字段 %s 失败: %w", field.Name, err)
			}
			fv.Set(reflect.ValueOf(s))
		}
	}
	return nil
}
