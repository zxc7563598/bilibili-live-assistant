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

// Send 发送一条弹幕到指定直播间
func (s *danmuSender) Send(ctx context.Context, roomID int64, message string) error {
	if sess := s.client.Session(); sess != nil {
		if _, isTest := s.testUIDs[sess.UID]; isTest {
			s.danmuLogger.Log(roomID, message)
		} else {
			csrfToken, err := s.client.CSRF()
			if err != nil {
				return fmt.Errorf("获取 CSRF Token 失败: %w", err)
			}
			return s.roomSvc.SendDanmu(ctx, roomID, message, csrfToken)
		}
	}
	return nil
}
