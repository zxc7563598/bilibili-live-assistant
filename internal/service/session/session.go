// Package session 承载各 service 共用的会话清理逻辑。
//
// 抽出来的原因：admin.Logout 与 role.logout 原本是两份逐字相同的实现
// （清 Redis 中的 token → 清库中的 token），连返回的错误码都一样。
//
// 这里只返回哨兵 error、不返回错误码：按本项目的错误码约定，
// 每个模块只能使用自己 MM 段的码，所以由调用方把哨兵映射成各自模块的码。
package session

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/admin"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/jwt"
)

var (
	// ErrTokenClearFailed 清理 Redis 中的 token 失败
	ErrTokenClearFailed = errors.New("清理 token 缓存失败")
	// ErrTokenPersistFailed 清理数据库中的 token 失败
	ErrTokenPersistFailed = errors.New("清理 token 记录失败")
)

// Logout 清理指定管理员的登录态，需在数据库事务之外调用，
// 避免事务回滚时 Redis 已清理造成两边不一致。
func Logout(ctx context.Context, rdb *redis.Client, adminRepo admin.Repository, adminID int64) error {
	if rdb != nil {
		if err := rdb.Del(ctx,
			jwt.AdminTokenKey(adminID),
			jwt.AdminRefreshKey(adminID),
		).Err(); err != nil {
			return fmt.Errorf("%w: %v", ErrTokenClearFailed, err)
		}
	}
	if err := adminRepo.UpdateTokenByID(ctx, nil, adminID, nil); err != nil {
		return fmt.Errorf("%w: %v", ErrTokenPersistFailed, err)
	}
	return nil
}
