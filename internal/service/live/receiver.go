package live

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// parseMessageData 根据 Cmd 类型调用对应的 Extract 函数解密消息
//
// 返回值是各 Extract 函数返回的指针类型（如 *live.DanmuMsgInfo），调用方需通过
// MessageProcessor.Process 内部的 cmd switch 做类型断言。
func parseMessageData(cmd live.Cmd, raw string) (any, error) {
	switch cmd {
	case live.CmdLiveStart:
		return live.ExtractLive(raw)
	case live.CmdLiveCutOff:
		return live.ExtractCutOff(raw)
	case live.CmdLiveRoomLock:
		return live.ExtractRoomLock(raw)
	case live.CmdLiveEnd:
		return live.ExtractPreparing(raw)
	case live.CmdSendGift:
		return live.ExtractSendGift(raw)
	case live.CmdSendGiftV2:
		return live.ExtractSendGiftV2(raw)
	case live.CmdGuardBuy:
		return live.ExtractGuardBuy(raw)
	case live.CmdSuperDanmuMsg:
		return live.ExtractSuperChatMessage(raw)
	case live.CmdInteractWord:
		return live.ExtractInteractWordV2(raw)
	case live.CmdDanmuMsg:
		return live.ExtractDanmuMsg(raw)
	case live.CmdPkStart:
		return live.ExtractPkBattlePreNew(raw)
	default:
		return nil, fmt.Errorf("未知的消息类型: %s", cmd)
	}
}

// processMessages 在后台 goroutine 中运行，从 listener 消费消息并更新统计
//
// 当 ctx 被取消（主动停止）时退出。如果检测到 msgCh 关闭且 ctx 未被取消（意外断连），
// 自动调用 reconnect 进行指数退避重连，成功后用新 listener 继续消息循环。
// 退出时关闭 done channel。
func (s *Service) processMessages(ctx context.Context, listener *live.Listener, done chan struct{}, rawLogger *logger.RawMessageLogger) {
	defer close(done)
	for {
		// 每次外层循环使用当前 listener 的 msgCh
		msgCh := listener.Messages()
		s.mu.Lock()
		s.listener = listener
		s.mu.Unlock()
		// 内层消息循环：读取消息直到 channel 关闭或主动停止
		disconnected := false
		for {
			select {
			case msg, ok := <-msgCh:
				if !ok {
					// msgCh 关闭 = listener 已停止
					disconnected = true
					goto checkReconnect
				}
				// 原始消息日志（独立目录 logs/直播间监听原始信息/）
				rawLogger.Log(string(msg.Cmd), msg.Raw)
				// 解密消息并进行处理
				data, err := parseMessageData(msg.Cmd, string(msg.Raw))
				if err != nil {
					log.Printf("[live.Receiver] [%s] 信息解密失败: %v", msg.Cmd, err)
					continue
				}
				// 更新统计
				s.mu.Lock()
				s.stats.msgCount++
				switch msg.Cmd {
				case live.CmdDanmuMsg:
					s.stats.danmuCount++
				case live.CmdSendGift, live.CmdSendGiftV2, live.CmdGuardBuy, live.CmdSuperDanmuMsg:
					s.stats.giftCount++
				}
				s.mu.Unlock()
				// 广播到所有已连接的前端 WebSocket 客户端
				jsonBytes, err := json.Marshal(data)
				if err != nil {
					log.Printf("[live.Receiver] [全站广播] JSON序列化失败: %v", err)
				} else {
					s.hub.Broadcast(string(msg.Cmd), jsonBytes)
				}
				// 分发给业务处理器（异步，不阻塞消息循环）
				go s.dispatcher.dispatch(ctx, msg.Cmd, data, s.roomID)
			case <-ctx.Done():
				return // 主动停止，退出整个函数
			}
		}
	checkReconnect:
		if !disconnected {
			return
		}
		// ctx.Err() != nil 说明是主动停止（cancel 触发了 ctx.Done() 但 msgCh 关闭先被 select 到了），不重连
		if ctx.Err() != nil {
			return
		}
		// 意外断连：停止弹幕发送队列并清空待发送消息
		s.mu.Lock()
		queue := s.queue
		s.mu.Unlock()
		if queue != nil {
			queue.Stop()
			queue.Clear()
			log.Printf("[live.Service] B站 WebSocket 断连，弹幕发送队列已停止并清空")
		}
		// 尝试重连
		newListener := s.reconnect(ctx)
		if newListener == nil {
			// 重连失败（超过最大重试次数），清理状态
			s.mu.Lock()
			s.listener = nil
			s.listenerCtx = nil
			s.listenerCancel = nil
			s.queue = nil
			s.mu.Unlock()
			// TODO: 发送邮件通知用户连接已断开且重连失败
			// s.notifyConnectionLost()
			log.Printf("[live.Service] B站 WebSocket 重连已达到最大重试次数（%d 次），已放弃监听，房间号: %d", 10, s.roomID)
			return
		}
		// 重连成功，重新启动弹幕发送队列
		if queue != nil {
			queue.Start(ctx)
			log.Printf("[live.Service] 弹幕发送队列已随重连重新启动")
		}
		listener = newListener // 用新 listener 继续外循环
	}
}

// reconnect 使用指数退避策略尝试重连 B站 WebSocket
//
// 最多重试 10 次（累计约 5 分钟），退避间隔从 1s 开始翻倍，封顶 60s。
// 成功返回新 listener，失败（超过上限或 ctx 被取消）返回 nil。
func (s *Service) reconnect(ctx context.Context) *live.Listener {
	const (
		maxBackoff  = 60 * time.Second
		maxAttempts = 10
	)
	backoff := 1 * time.Second
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
		s.mu.Lock()
		roomID := s.roomID
		s.mu.Unlock()
		log.Printf("[live.Service] B站 WebSocket 断连，正在重连... 第 %d/%d 次，房间号: %d，等待间隔: %v",
			attempt, maxAttempts, roomID, backoff)
		listener, err := live.NewListenerFromClient(
			context.Background(), s.client, roomID,
			live.WithPingInterval(30*time.Second),
			live.WithMsgChannelSize(512),
		)
		if err != nil {
			log.Printf("[live.Service] 重连失败：无法创建 Listener（第 %d 次），错误: %v", attempt, err)
			backoff = min(backoff*2, maxBackoff)
			continue
		}
		if err := listener.Connect(ctx); err != nil {
			log.Printf("[live.Service] 重连失败：WebSocket 连接失败（第 %d 次），错误: %v", attempt, err)
			backoff = min(backoff*2, maxBackoff)
			continue
		}
		log.Printf("[live.Service] B站 WebSocket 重连成功（第 %d 次），房间号: %d", attempt, roomID)
		return listener
	}
	// 超过最大重试次数
	log.Printf("[live.Service] B站 WebSocket 重连失败，已达到最大重试次数（%d 次）", maxAttempts)
	// TODO: 发送邮件通知用户连接异常
	// s.sendDisconnectEmail()
	return nil
}
