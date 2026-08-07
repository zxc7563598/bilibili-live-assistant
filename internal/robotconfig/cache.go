// Package robotconfig 提供机器人配置的内存缓存
//
// 项目启动时全量加载 robot_configs 表到内存，按 group_name 分组。
// 作为基础设施组件，与 *gorm.DB / *redis.Client 同级，
// 任意 service 通过构造函数注入后直接读缓存，无需查数据库。
package robotconfig

import (
	"context"
	"fmt"
	"sync"

	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/robot_config"
)

// Cache 机器人配置内存缓存
type Cache struct {
	mu    sync.RWMutex
	repo  robot_config.Repository
	cache map[string]map[string]string // group_name -> config_key -> config_value
}

// New 创建配置缓存（仅创建实例，需调用 Init 加载数据）
func New(repo robot_config.Repository) *Cache {
	return &Cache{
		repo:  repo,
		cache: make(map[string]map[string]string),
	}
}

// Init 从数据库全量加载配置到内存
func (c *Cache) Init(ctx context.Context) error {
	configs, err := c.repo.FindAll(ctx, nil)
	if err != nil {
		return fmt.Errorf("加载机器人配置失败: %w", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]map[string]string, len(configs))
	for _, cfg := range configs {
		if c.cache[cfg.GroupName] == nil {
			c.cache[cfg.GroupName] = make(map[string]string)
		}
		c.cache[cfg.GroupName][cfg.ConfigKey] = cfg.ConfigValue
	}
	return nil
}

// Get 获取指定分组下的单条配置值
func (c *Cache) Get(groupName, configKey string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	group, ok := c.cache[groupName]
	if !ok {
		return "", false
	}
	val, ok := group[configKey]
	return val, ok
}

// GetGroup 获取指定分组下的全部配置
func (c *Cache) GetGroup(groupName string) (map[string]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	group, ok := c.cache[groupName]
	if !ok {
		return nil, false
	}
	result := make(map[string]string, len(group))
	for k, v := range group {
		result[k] = v
	}
	return result, ok
}

// Reload 从数据库全量重新加载配置
func (c *Cache) Reload(ctx context.Context) error {
	return c.Init(ctx)
}
