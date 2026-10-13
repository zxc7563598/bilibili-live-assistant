package robotconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// snapKey 解析快照的键：配置分组名 + 目标结构体类型
type snapKey struct {
	group string
	typ   reflect.Type
}

// UnmarshalGroup 将指定分组的配置解析到结构体指针。
// 结构体字段需使用 `config:"key_name"` tag 指定映射的配置键名。
// 未配置 tag 的字段会被跳过，配置中不存在的 key 会保留字段零值。
//
// 解析结果按 (分组, 类型) 缓存，同一份配置只反射解析一次，Reload 后重建。
// 取到的是浅拷贝，切片字段与缓存共享底层数组，调用方只能读取。
func (c *Cache) UnmarshalGroup(groupName string, v any) error {
	elem, err := structElem(v)
	if err != nil {
		return err
	}
	key := snapKey{group: groupName, typ: elem.Type()}

	c.mu.RLock()
	cached, hit := c.snap[key]
	group, ok := c.cache[groupName]
	c.mu.RUnlock()
	if hit {
		elem.Set(reflect.ValueOf(cached).Elem())
		return nil
	}
	if !ok {
		return fmt.Errorf("配置分组 %s 不存在", groupName)
	}
	if err := unmarshalConfig(group, v); err != nil {
		return err
	}
	// 存副本：快照与调用方各自的后续改动互不影响（切片底层数组仍共享）
	snapshot := reflect.New(elem.Type())
	snapshot.Elem().Set(elem)
	c.mu.Lock()
	c.snap[key] = snapshot.Interface()
	c.mu.Unlock()
	return nil
}

// structElem 校验 v 是指向结构体的非空指针，返回其指向的结构体值
func structElem(v any) (reflect.Value, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return reflect.Value{}, errors.New("v 必须是非空的指针")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, errors.New("v 必须是指向结构体的指针")
	}
	return rv, nil
}

func unmarshalConfig(m map[string]string, v any) error {
	rv, err := structElem(v)
	if err != nil {
		return err
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
		} else {
			if val == "" {
				continue
			}
			dest := reflect.New(fv.Type())
			if err := json.Unmarshal([]byte(val), dest.Interface()); err != nil {
				return fmt.Errorf("解析字段 %s 失败: %w", field.Name, err)
			}
			fv.Set(dest.Elem())
		}
	}
	return nil
}
