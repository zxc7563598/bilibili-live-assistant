package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
)

// NewAccountRateLimiter 按账号(UID)限流的内存版固定窗口限流器。
//
// 需挂在已解密请求体的中间件（ShopEncrypt）之后，从明文 JSON 提取 account 字段作为统计 key，
// 用于防止对同一账号暴力撞库 / 频繁探测。
// 纯内存计数、不依赖 Redis（Redis 可选，未配置时也正常工作）；单实例自部署场景够用，
// 多实例负载均衡时各实例独立计数，进程重启后计数清零。
func NewAccountRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	l := &accountRateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string]*accountRateBucket),
	}
	go l.cleanupLoop()
	return l.middleware
}

type accountRateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]*accountRateBucket
}

// accountRateBucket 单个账号的固定窗口计数
type accountRateBucket struct {
	count int       // 窗口内已请求次数
	start time.Time // 窗口起点
}

func (l *accountRateLimiter) middleware(c *gin.Context) {
	account, ok := readAccount(c)
	if !ok {
		// 空/非法请求体提取不到账号 → 不参与限流，交由下游参数校验拒绝
		c.Next()
		return
	}
	if !l.allow(account) {
		response.Error(c, "", 20002)
		c.Abort()
		return
	}
	c.Next()
}

// readAccount 读取并还原请求体，从明文 JSON 中提取 account 字段
func readAccount(c *gin.Context) (string, bool) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", false
	}
	// 还原请求体，供下游 ShouldBindJSON 绑定
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var req struct {
		Account int64 `json:"account"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.Account == 0 {
		return "", false
	}
	return strconv.FormatInt(req.Account, 10), true
}

func (l *accountRateLimiter) allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[key]
	if !ok || now.Sub(b.start) >= l.window {
		// 新窗口：重置计数
		l.buckets[key] = &accountRateBucket{count: 1, start: now}
		return true
	}
	b.count++
	return b.count <= l.limit
}

// cleanupLoop 周期清理过期桶，防止 map 无限增长
func (l *accountRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.window)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-l.window)
		l.mu.Lock()
		for k, b := range l.buckets {
			if b.start.Before(cutoff) {
				delete(l.buckets, k)
			}
		}
		l.mu.Unlock()
	}
}
