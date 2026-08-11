package live

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/room"
)

// danmuSender 实现 live.Sender 接口，通过 B站 API 实际发送弹幕
//
// 每次 Send 调用会实时从 client.CSRF() 获取 CSRF Token，
// 确保 token 始终有效（不缓存可能过期的 token）
//
// TODO: 测试阶段，弹幕不真实发送，仅记录到日志文件
type danmuSender struct {
	roomSvc     *room.Service
	client      *bilibili.Client
	danmuLogger *logger.DanmuSendLogger // 弹幕发送日志记录器
}

// Send 发送一条弹幕到指定直播间
//
// 测试阶段：不实际发送弹幕，仅将内容记录到日志文件 logs/弹幕发送记录/
// 正式上线时取消注释 SendDanmu 调用即可恢复
func (s *danmuSender) Send(ctx context.Context, roomID int64, message string) error {
	s.danmuLogger.Log(roomID, message)
	// csrfToken, err := s.client.CSRF()
	// if err != nil {
	// 	return fmt.Errorf("获取 CSRF Token 失败: %w", err)
	// }
	// return s.roomSvc.SendDanmu(ctx, roomID, message, csrfToken)
	return nil
}
