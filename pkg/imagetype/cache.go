package imagetype

import (
	"sync"
	"time"
)

// mimeCacheEntry 缓存条目
type mimeCacheEntry struct {
	mime    string
	err     error
	expires time.Time
}

// mimeCache 带 TTL 与条数上限的 MIME 检测结果缓存（成功与失败都缓存）。
// get 仅持读锁、set 仅持写锁，二者互斥且无锁升级，不会死锁。
type mimeCache struct {
	mu    sync.RWMutex
	max   int
	ttl   time.Duration
	items map[string]mimeCacheEntry
}

func newMimeCache(max int, ttl time.Duration) *mimeCache {
	return &mimeCache{
		max:   max,
		ttl:   ttl,
		items: make(map[string]mimeCacheEntry),
	}
}

// get 读取缓存；过期视为未命中（惰性删除，读锁下不删条目）。
func (c *mimeCache) get(key string) (mime string, err error, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expires) {
		return "", nil, false
	}
	return e.mime, e.err, true
}

// set 写入缓存；先清理过期条目，仍超上限则简单驱逐（非精确 LRU）。
func (c *mimeCache) set(key string, mime string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, e := range c.items {
		if now.After(e.expires) {
			delete(c.items, k)
		}
	}
	c.items[key] = mimeCacheEntry{mime: mime, err: err, expires: now.Add(c.ttl)}
	for k := range c.items {
		if len(c.items) <= c.max {
			break
		}
		delete(c.items, k)
	}
}
