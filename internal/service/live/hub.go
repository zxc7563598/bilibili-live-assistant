package live

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub 管理所有前端 WebSocket 连接，负责将 B站 直播间消息广播给已连接的前端客户端
//
// Hub 在 Service.New() 时创建并启动，独立于 Listener 生命周期：
// 前端可以在监听器未启动时连接等待，有消息就推送，没消息就保持连接。
type Hub struct {
	// clients 当前已连接的客户端集合
	clients map[*Client]bool
	// register 注册新客户端
	register chan *Client
	// unregister 注销客户端
	unregister chan *Client
	// broadcast 广播消息（已序列化为 JSON []byte）
	broadcast chan []byte
	// mu 保护 clients map 的并发访问
	mu sync.RWMutex
}

// Client 封装单个前端 WebSocket 连接
//
// 每个 Client 有独立的写 goroutine（writePump），通过 send channel 接收要发送的消息。
// 读 goroutine（readPump）仅用于检测客户端断开连接。
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte // 缓冲发送 channel
}

const (
	// writeWait 写超时时间
	writeWait = 10 * time.Second
	// pingPeriod 服务端发送 ping 的间隔（必须小于 writeWait）
	pingPeriod = 30 * time.Second
	// sendBufferSize 每个客户端的发送缓冲区大小
	sendBufferSize = 256
	// broadcastBufferSize Hub 广播 channel 缓冲区大小
	broadcastBufferSize = 512
)

// newHub 创建一个新的 Hub 实例
func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, broadcastBufferSize),
	}
}

// Run 启动 Hub 的主循环，在独立的 goroutine 中运行
//
// 处理客户端注册、注销和消息广播。
// ctx 取消时关闭所有客户端连接并退出。
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.closeAllClients()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[live.Hub] 前端客户端已连接，当前连接数: %d", h.ClientCount())

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[live.Hub] 前端客户端已断开，当前连接数: %d", h.ClientCount())

		case data := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					// 客户端 send channel 已满，跳过（避免慢客户端拖慢整体）
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast 将一条 B站 消息序列化为 JSON 并广播给所有已连接的前端客户端
//
// 非阻塞：如果广播 channel 已满，消息会被丢弃。
// 消息格式：{"cmd":"...", "data":{...处理过的json数据...}}
func (h *Hub) Broadcast(cmd string, data json.RawMessage) {
	data, err := marshalWSMessage(cmd, data)
	if err != nil {
		log.Printf("[live.Hub] 序列化消息失败: %v", err)
		return
	}
	select {
	case h.broadcast <- data:
	default:
		// 广播 channel 满，丢弃本条消息
	}
}

// RegisterClient 将新客户端注册到 Hub
func (h *Hub) RegisterClient(c *Client) {
	h.register <- c
}

// UnregisterClient 从 Hub 注销客户端
func (h *Hub) UnregisterClient(c *Client) {
	h.unregister <- c
}

// ClientCount 返回当前已连接的客户端数量
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// closeAllClients 关闭所有客户端连接（在 Hub 停止时调用）
func (h *Hub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		client.conn.Close()
	}
}

// =========================================================================
// Client 方法
// =========================================================================

// NewClient 创建一个新的 WebSocket 客户端
//
// 创建后需调用 Hub.RegisterClient 注册，然后启动 ReadPump 和 WritePump goroutine。
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, sendBufferSize),
	}
}

// ReadPump 从 WebSocket 读取消息，仅用于检测客户端断开连接
//
// 客户端不会主动发送业务消息，所以收到任何消息（或错误）都视为断开。
// 断开时自动从 Hub 注销。
func (c *Client) ReadPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()
	// 客户端不发送消息，只等待连接关闭
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
		// 收到消息也断开（客户端不应发送消息）
		break
	}
}

// WritePump 从 send channel 读取消息并写入 WebSocket
//
// 运行在独立的 goroutine 中。
// 支持 ping/pong 保活机制，定期发送 ping 检测连接是否存活。
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return

		case message, ok := <-c.send:
			if !ok {
				// send channel 已关闭
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// =========================================================================
// 内部工具函数
// =========================================================================

// wsMessage 是向前端推送的 WebSocket 消息格式
type wsMessage struct {
	Cmd  string          `json:"cmd"`
	Data json.RawMessage `json:"data"`
}

// marshalWSMessage 将 B站 live.Message 序列化为前端 WebSocket 消息的 JSON
func marshalWSMessage(cmd string, data json.RawMessage) ([]byte, error) {
	m := wsMessage{
		Cmd:  cmd,
		Data: data,
	}
	return json.Marshal(m)
}
