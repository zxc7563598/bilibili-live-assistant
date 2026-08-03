package live

import (
	"context"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/room"
)

// danmuSender 实现 live.Sender 接口，通过 B站 API 实际发送弹幕
//
// 每次 Send 调用会实时从 client.CSRF() 获取 CSRF Token，
// 确保 token 始终有效（不缓存可能过期的 token）
type danmuSender struct {
	roomSvc *room.Service
	client  *bilibili.Client
}

// Send 发送一条弹幕到指定直播间
//
// 内部调用 room.Service.SendDanmu，该接口会先获取弹幕发送权限再发送
func (s *danmuSender) Send(ctx context.Context, roomID int64, message string) error {
	csrfToken, err := s.client.CSRF()
	if err != nil {
		return fmt.Errorf("获取 CSRF Token 失败: %w", err)
	}
	return s.roomSvc.SendDanmu(ctx, roomID, message, csrfToken)
}
