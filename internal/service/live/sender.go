package live

import (
	"context"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/room"
)

// danmuSender 实现 live.Sender 接口，通过 B站 API 实际发送弹幕
//
// 每次 Send 调用会实时从 client.CSRF() 获取 CSRF Token，
// 确保 token 始终有效（不缓存可能过期的 token）
type danmuSender struct {
	roomSvc     *room.Service
	client      *bilibili.Client
	danmuLogger *logger.DanmuSendLogger // 弹幕发送日志记录器
	testUIDs    map[int64]struct{}      // 测试机器人 UID 白名单，命中则仅记录日志不真正发送
}

// Send 发送弹幕到指定直播间。
//
// 当 message 超过当前账号允许的单条弹幕长度（rune 数）时，只发送其前缀，
// 并把未发完的部分通过返回值交还 Queue，由 Queue 在后续 tick 中继续调用
// Send 发送剩余部分，直到整条消息发完。
func (s *danmuSender) Send(ctx context.Context, roomID int64, message string) (string, error) {
	sess := s.client.Session()
	if sess == nil {
		return "", nil
	}
	if _, isTest := s.testUIDs[sess.UID]; isTest {
		// 测试模式仅记录日志不真正发送，无需按长度拆分
		s.danmuLogger.Log(roomID, message)
		return "", nil
	}
	csrfToken, err := s.client.CSRF()
	if err != nil {
		return "", fmt.Errorf("获取 CSRF Token 失败: %w", err)
	}
	perm, err := s.roomSvc.GetBarragePermissionCached(ctx, roomID)
	if err != nil {
		return "", fmt.Errorf("获取弹幕发送权限失败: %w", err)
	}
	// 按 rune 拆分而非字节，避免把中文等 UTF-8 多字节字符拦腰截断
	runes := []rune(message)
	if perm.Length <= 0 || len(runes) <= perm.Length {
		return "", s.roomSvc.SendDanmuWithPermission(ctx, roomID, message, csrfToken, perm)
	}
	head := string(runes[:perm.Length])
	rest := string(runes[perm.Length:])
	return rest, s.roomSvc.SendDanmuWithPermission(ctx, roomID, head, csrfToken, perm)
}
