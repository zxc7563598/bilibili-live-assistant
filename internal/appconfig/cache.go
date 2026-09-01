// Package appconfig 提供应用配置的内存缓存
//
// 项目启动时全量加载 app_configs 表到内存，按 config_key 平铺存储。
// 作为基础设施组件，与 *gorm.DB / *redis.Client 同级，
// 任意 service 通过构造函数注入后直接读缓存，无需查数据库。
package appconfig

import (
	"context"
	"fmt"
	"sync"

	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/app_config"
)

// Cache 应用配置内存缓存
type Cache struct {
	mu    sync.RWMutex
	repo  app_config.Repository
	cache map[string]string // config_key -> config_value
}

// New 创建配置缓存（仅创建实例，需调用 Init 加载数据）
func New(repo app_config.Repository) *Cache {
	return &Cache{
		repo:  repo,
		cache: make(map[string]string),
	}
}

// Init 从数据库全量加载配置到内存
func (c *Cache) Init(ctx context.Context) error {
	configs, err := c.repo.GetAll(ctx, nil)
	if err != nil {
		return fmt.Errorf("加载应用配置失败: %w", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]string, len(configs))
	for _, cfg := range configs {
		c.cache[cfg.ConfigKey] = cfg.ConfigValue
	}
	return nil
}

// Get 获取指定配置键的配置值
func (c *Cache) Get(configKey string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.cache[configKey]
	return val, ok
}

// GetAll 获取全部配置（返回副本，调用方修改不影响缓存）
func (c *Cache) GetAll() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]string, len(c.cache))
	for k, v := range c.cache {
		result[k] = v
	}
	return result
}

// Reload 从数据库全量重新加载配置
func (c *Cache) Reload(ctx context.Context) error {
	return c.Init(ctx)
}
