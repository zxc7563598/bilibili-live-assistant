package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisLogger 把 go-redis 自身的日志转发到标准库 log
//
// go-redis 的默认 logger 自建 log.New(os.Stderr, "redis: ", ...)，绕过 log.SetOutput，
// 连接与连接池异常会在宝塔面板日志里留下游离输出。转一道手之后，
// 它随标准库输出一起落进 <日志根目录>/stdlog/。
type redisLogger struct{}

func (redisLogger) Printf(_ context.Context, format string, v ...interface{}) {
	log.Printf("[redis] "+format, v...)
}

// InitRedis 根据配置初始化redis
func InitRedis(cfg *Config) (*redis.Client, error) {
	// 未配置redis
	if cfg.Redis.Host == "" || cfg.Redis.Port == 0 {
		return nil, nil
	}
	// go-redis 的 logger 是包级全局，在这里覆盖一次即可
	redis.SetLogger(redisLogger{})
	// 连接redis
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis连接失败: %w", err)
	}
	return rdb, nil
}
