package live

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/room"
)

// Listener 是 B站 直播间 WebSocket 监听器
//
// 它在独立的 goroutine 中运行，接收直播间的实时消息（弹幕、礼物、互动等）
// 默认创建后处于未连接状态，需调用 Start 传入房间连接信息后开始监听
//
// 连接成功后的流程：
//   - 发送认证包（操作码 7）
//   - 发送首次心跳（操作码 2）
//   - 触发 onConnected 回调（如果设置）
//   - 启动心跳定时器并持续读取消息投递到 Messages() channel
//
// 典型用法：
//
//	l := live.NewListener(roomID, live.WithAuth(uid, buvid3))
//	l.Start(ctx, danmuInfo)
//
//	for msg := range l.Messages() {
//	    switch msg.Cmd {
//	    case live.CmdDanmuMsg:
//	        // 处理弹幕
//	    case live.CmdSendGift:
//	        // 处理礼物
//	    }
//	}
//
//	l.Stop() // 停止监听
type Listener struct {
	roomID int64
	// 认证信息
	uid    int64
	buvid3 string
	// 状态保护
	mu      sync.Mutex
	running bool
	conn    *websocket.Conn
	cancel  context.CancelFunc
	// msgCh 是接收消息的 channel, Start 后即可从中读取
	msgCh chan *Message
	// done 在所有 goroutine 退出后关闭, 用于 Stop 等待
	done chan struct{}
	// 预获取的连接信息
	danmuInfo *room.DanmuInfo
	wsHeaders http.Header
	// pingInterval 心跳间隔，默认 30 秒
	pingInterval time.Duration
	// msgChSize 消息 channel 缓冲区大小
	msgChSize int
	// onConnected 在认证包发送成功后回调
	onConnected func()
}

// ListenerOption 是 Listener 的函数式配置项
type ListenerOption func(*Listener)

// WithAuth 设置认证所需的用户信息（uid 和 buvid3）
// uid 从 Cookie 中的 DedeUserID 获取
// 有登录态时必须设置，否则认证包中 uid=0, buvid3=""
func WithAuth(uid int64, buvid3 string) ListenerOption {
	return func(l *Listener) {
		l.uid = uid
		l.buvid3 = buvid3
	}
}

// WithPingInterval 设置心跳间隔（默认 30 秒）
func WithPingInterval(d time.Duration) ListenerOption {
	return func(l *Listener) {
		l.pingInterval = d
	}
}

// WithMsgChannelSize 设置消息 channel 缓冲区大小（默认 256）
func WithMsgChannelSize(size int) ListenerOption {
	return func(l *Listener) {
		l.msgChSize = size
	}
}

// WithOnConnected 设置连接成功后的回调（认证包和首次心跳发送后触发）
func WithOnConnected(fn func()) ListenerOption {
	return func(l *Listener) {
		l.onConnected = fn
	}
}

// NewListener 创建一个新的直播间监听器
//
// 创建后监听器处于未连接状态，需调用 Start 传入房间连接信息后开始工作
func NewListener(roomID int64, opts ...ListenerOption) *Listener {
	l := &Listener{
		roomID:       roomID,
		pingInterval: 30 * time.Second,
		msgChSize:    256,
		done:         make(chan struct{}),
	}
	for _, o := range opts {
		o(l)
	}
	l.msgCh = make(chan *Message, l.msgChSize)
	return l
}

// Start 连接到直播间 WebSocket 并开始监听
//
// 连接成功后会依次执行：
//   - 发送认证包（B站 二进制帧，操作码 7）
//   - 发送首次心跳（B站 二进制帧，操作码 2）
//   - 触发 onConnected 回调（如果设置）
//   - 启动心跳定时器 + 消息读取循环
//
// # Start 调用后立即返回，消息通过 Messages() channel 异步投递
//
// 参数：
//   - ctx: 用于控制整个监听生命周期，ctx 被取消时自动断开连接
//   - info: 弹幕 WebSocket 连接信息，通过 room.Service.GetDanmuInfo() 获取
//
// 如果已在运行中，Start 返回错误需要先 Stop 再重新 Start
func (l *Listener) Start(ctx context.Context, info *room.DanmuInfo) error {
	return l.startInternal(ctx, info, nil)
}

// StartWithHeaders 与 Start 相同，但允许在 WebSocket 握手时携带自定义 HTTP Headers
//
// 典型用途：携带 Cookie 以获取需要登录态的直播间消息
func (l *Listener) StartWithHeaders(ctx context.Context, info *room.DanmuInfo, headers http.Header) error {
	return l.startInternal(ctx, info, headers)
}

// Connect 使用预获取的连接信息启动 WebSocket 监听
//
// 仅在通过 NewListenerFromClient 创建 Listener 时可用，
// 因为它内部已自动获取了 danmuInfo 和 wsHeaders
//
// 如果使用 NewListener 手动创建 Listener，请使用 Start / StartWithHeaders
func (l *Listener) Connect(ctx context.Context) error {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return fmt.Errorf("live: 房间 %d 的监听器已在运行中", l.roomID)
	}
	info := l.danmuInfo
	headers := l.wsHeaders
	l.mu.Unlock()

	if info == nil {
		return fmt.Errorf("live: 没有预配置的连接信息；请使用 Start 或 StartWithHeaders 代替")
	}

	return l.startInternal(ctx, info, headers)
}

func (l *Listener) startInternal(ctx context.Context, info *room.DanmuInfo, headers http.Header) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.running {
		return fmt.Errorf("live: 房间 %d 的监听器已在运行中", l.roomID)
	}

	// 过滤掉 gorilla/websocket 自动设置的 WebSocket 握手头，避免 duplicate header 错误
	// gorilla/websocket 会自动设置: Upgrade, Connection, Sec-WebSocket-Key, Sec-WebSocket-Version
	headers = filterWebSocketHeaders(headers)

	wssURL := fmt.Sprintf("wss://%s:%d/sub", info.Host, info.WSSPort)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wssURL, headers)
	if err != nil {
		return fmt.Errorf("live: WebSocket %s 握手/连接失败: %w", wssURL, err)
	}

	ctx, l.cancel = context.WithCancel(ctx)
	l.conn = conn
	l.running = true

	go l.run(ctx, info.Token)

	return nil
}

// Stop 停止监听
//
// 关闭 WebSocket 连接，取消内部 context，等待所有 goroutine 退出后返回
// 如果未在运行中，Stop 是幂等的（无操作）
func (l *Listener) Stop() error {
	l.mu.Lock()
	if !l.running {
		l.mu.Unlock()
		return nil
	}
	if l.cancel != nil {
		l.cancel()
	}
	l.mu.Unlock()

	<-l.done

	l.mu.Lock()
	l.running = false
	l.conn = nil
	l.mu.Unlock()

	return nil
}

// Messages 返回消息接收 channel
//
// Start 之后可从中持续读取收到的直播间消息。Stop 之后 channel 不会被关闭，
// 调用方应通过 context 或 Stop 返回来判断何时停止读取
func (l *Listener) Messages() <-chan *Message {
	return l.msgCh
}

// IsRunning 返回当前是否正在监听
func (l *Listener) IsRunning() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.running
}

// RoomID 返回监听的房间号
func (l *Listener) RoomID() int64 {
	return l.roomID
}

// =========================================================================
// 内部实现
// =========================================================================

// run 是监听主循环，运行在独立的 goroutine 中
//
// 流程：
//   - 发送认证包
//   - 发送首次心跳
//   - 触发 onConnected 回调
//   - 启动心跳定时器和消息读取循环
func (l *Listener) run(ctx context.Context, token string) {
	defer close(l.done)
	defer func() {
		l.mu.Lock()
		if l.conn != nil {
			l.conn.Close()
		}
		l.mu.Unlock()
	}()
	// 发送认证包（操作码 7）
	log.Printf("[live.Listener] 已连接到 WebSocket，房间号: %d", l.roomID)

	authFrame, err := BuildAuthFrame(l.roomID, l.uid, l.buvid3, token)
	if err != nil {
		log.Printf("[live.Listener] 构建认证包失败: %v", err)
		return
	}
	if err := l.conn.WriteMessage(websocket.BinaryMessage, authFrame); err != nil {
		log.Printf("[live.Listener] 发送认证包失败: %v", err)
		return
	}
	log.Printf("[live.Listener] 认证包发送")

	// 发送首次心跳（操作码 2）
	heartbeatFrame := BuildHeartbeatFrame()
	if err := l.conn.WriteMessage(websocket.BinaryMessage, heartbeatFrame); err != nil {
		log.Printf("[live.Listener] 发送首次心跳失败: %v", err)
		return
	}
	log.Printf("[live.Listener] 首次 websocket 心跳发送")

	// 触发 onConnected 回调
	if l.onConnected != nil {
		l.onConnected()
	}

	// 启动心跳定时器
	pingTicker := time.NewTicker(l.pingInterval)
	defer pingTicker.Stop()

	// 启动读取循环（单独 goroutine，读错误通过 channel 汇报）
	readErrCh := make(chan error, 1)
	go func() {
		readErrCh <- l.readLoop(ctx)
	}()

	// 主循环：等待心跳触发、读取错误、或 context 取消
	for {
		select {
		case <-ctx.Done():
			return

		case <-pingTicker.C:
			if err := l.sendHeartbeat(); err != nil {
				log.Printf("[live.Listener] 心跳发送失败: %v", err)
				return
			}

		case err := <-readErrCh:
			if err != nil {
				log.Printf("[live.Listener] 读取消息错误: %v", err)
				// TODO: 实现自动重连逻辑
				return
			}
			return
		}
	}
}

// filterWebSocketHeaders 移除 gorilla/websocket 会自动设置的 WebSocket 握手头
// 调用方如果手动传入这些头会导致 "duplicate header" 错误，这里统一过滤
//
// gorilla/websocket 自动设置的 header:
//   - Upgrade: websocket
//   - Connection: Upgrade
//   - Sec-WebSocket-Key（自动生成随机值）
//   - Sec-WebSocket-Version: 13
//   - Sec-WebSocket-Extensions（根据 Dialer 配置自动生成）
func filterWebSocketHeaders(h http.Header) http.Header {
	if h == nil {
		return nil
	}
	// websocketHeaders 是 gorilla/websocket 内部管理的 header，调用方不应设置
	websocketHeaders := []string{
		"Upgrade",
		"Connection",
		"Sec-Websocket-Key", // http.CanonicalHeaderKey 转换后的格式
		"Sec-Websocket-Version",
		"Sec-Websocket-Extensions",
	}
	for _, key := range websocketHeaders {
		h.Del(key)
	}
	return h
}

// sendHeartbeat 发送心跳帧（操作码 2）
func (l *Listener) sendHeartbeat() error {
	heartbeatFrame := BuildHeartbeatFrame()
	return l.conn.WriteMessage(websocket.BinaryMessage, heartbeatFrame)
}

// readLoop 持续从 WebSocket 读取消息并解析投递到 msgCh
func (l *Listener) readLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		_, raw, err := l.conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("live: 从 WebSocket 接收消息时出错: %w", err)
		}

		// 使用 B站 协议解析响应帧
		messages, err := ParseResponse(raw)
		if err != nil {
			// 无法解析的消息跳过
			continue
		}

		for _, msg := range messages {
			if msg == nil || msg.Cmd == "" {
				continue
			}
			select {
			case l.msgCh <- msg:
			case <-ctx.Done():
				return nil
			default:
				// msgCh 缓冲区满，丢弃本条消息
			}
		}
	}
}
