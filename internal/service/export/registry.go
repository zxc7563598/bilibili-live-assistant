package export

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

// Source 由一个希望支持导出的业务模块实现
type Source interface {
	// Module 模块标识，与前端 MeCrud 的 export-module 取值一致
	Module() string
	// Columns 本模块允许导出的列。顺序仅作默认顺序，实际列序以调用方传入的 keys 为准
	Columns() []Column
	// Normalize 解析并校验前端传来的筛选条件，返回规范化后的 JSON
	Normalize(filters json.RawMessage) (json.RawMessage, int, error)
	// Count 统计命中行数。limit > 0 时实现方须保证「超过 limit 即可提前停止」，避免在大表上做全量 COUNT
	Count(ctx context.Context, filters json.RawMessage, limit int) (int64, error)
	// FetchChunk 按有索引的唯一列（约定主键）严格递减取一块：主键 < afterID（0 不限）且满足 filters，最多 limit 行，返回最后一行主键作 nextID 游标；必须如此排序，否则会全表扫描或跳行重复。
	FetchChunk(ctx context.Context, filters json.RawMessage, afterID int64, keys []string, limit int) (rows [][]any, nextID int64, err error)
}

// Registry 已注册的导出模块。
type Registry struct {
	sources map[string]Source
}

// NewRegistry 构建注册表。module 重复注册会 panic —— 属装配期错误，早失败好过静默覆盖。
func NewRegistry(sources ...Source) *Registry {
	r := &Registry{sources: make(map[string]Source, len(sources))}
	for _, s := range sources {
		if s == nil {
			continue
		}
		name := s.Module()
		if _, dup := r.sources[name]; dup {
			panic("export: 模块重复注册: " + name)
		}
		r.sources[name] = s
	}
	return r
}

// Get 按模块名取 Source
func (r *Registry) Get(module string) (Source, bool) {
	s, ok := r.sources[module]
	return s, ok
}

// Modules 返回已注册的模块名（升序，便于日志与排查）
func (r *Registry) Modules() []string {
	names := make([]string, 0, len(r.sources))
	for name := range r.sources {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// resolveColumns 把调用方传来的 key 列表解析为列声明
func resolveColumns(src Source, keys []string) ([]Column, int, error) {
	if len(keys) == 0 {
		return nil, CodeColumnEmpty, nil
	}
	// 上限只是防呆：正常页面不会接近这个数量
	const maxColumns = 50
	if len(keys) > maxColumns {
		return nil, CodeColumnInvalid, nil
	}
	declared := make(map[string]Column)
	for _, c := range src.Columns() {
		declared[c.Key] = c
	}
	cols := make([]Column, 0, len(keys))
	seen := make(map[string]bool, len(keys))
	for _, key := range keys {
		col, ok := declared[key]
		if !ok {
			return nil, CodeColumnInvalid, fmt.Errorf("模块 %s 未声明的导出列: %q", src.Module(), key)
		}
		if seen[key] {
			continue // 同一个 key 重复出现只导出一次
		}
		seen[key] = true
		cols = append(cols, col)
	}
	return cols, 0, nil
}
