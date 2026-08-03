package live

import "time"

// QRCodeResp 扫码登录二维码信息
type QRCodeResp struct {
	URL       string
	QrcodeKey string
}

// PollQRCodeResp 扫码状态轮询结果
type PollQRCodeResp struct {
	Status    int
	Message   string
	IsScanned bool
	IsSuccess bool
	IsExpired bool
}

// LoginStatusResp B站 登录状态
type LoginStatusResp struct {
	IsLoggedIn bool
	UID        int64
	Username   string
	Face       string
	Buvid      string
}

// ListenerStatusResp 监听器状态与统计
type ListenerStatusResp struct {
	IsRunning  bool
	RoomID     int64
	StartTime  string
	Uptime     string
	MsgCount   int64
	DanmuCount int64
	GiftCount  int64
	UID        int64
	Title      string
	LiveStatus int
	Online     int
	Attention  int
	LiveTime   string
}

// listenerStats 消息统计（内部使用）
type listenerStats struct {
	startTime  time.Time
	msgCount   int64
	danmuCount int64
	giftCount  int64
}
