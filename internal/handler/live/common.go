package live

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	liveSvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/live"
)

// Handler 直播控制 HTTP 接口处理器
type Handler struct {
	liveSvc *liveSvc.Service
	rdb     *redis.Client
}

// New 创建 Handler 实例
func New(liveSvc *liveSvc.Service, rdb *redis.Client) *Handler {
	return &Handler{liveSvc: liveSvc, rdb: rdb}
}

// wsUpgrader 是 HTTP 到 WebSocket 的升级器
//
// CheckOrigin 允许所有来源，因为开发环境前后端运行在不同端口
var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
