package live

import (
	"sync"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/room"
)

// RoomState 直播间信息缓存，线程安全
type RoomState struct {
	mu   sync.RWMutex
	info *room.RealRoomInfo
}

// Update 用 API 返回的完整信息替换缓存
func (rs *RoomState) Update(info *room.RealRoomInfo) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	copy := *info
	rs.info = &copy
}

// LiveStatus 返回当前直播状态
// 缓存为空时返回 -1
func (rs *RoomState) LiveStatus() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if rs.info == nil {
		return -1
	}
	return rs.info.LiveStatus
}

// UID 返回当前房间的主播 UID
// 缓存为空时返回 0
func (rs *RoomState) UID() int64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if rs.info == nil {
		return 0
	}
	return rs.info.UID
}

// SetLiveStatus 仅更新直播状态字段，用于 WebSocket 上下播事件。
func (rs *RoomState) SetLiveStatus(status int) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.info == nil {
		rs.info = &room.RealRoomInfo{LiveStatus: status}
	} else {
		rs.info.LiveStatus = status
	}
}
