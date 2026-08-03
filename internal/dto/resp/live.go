package resp

// LiveQRCodeResp 扫码登录二维码信息
type LiveQRCodeResp struct {
	// 用以生成二维码的URL
	URL string `json:"url" example:"https://api.bilibili.com/x/..."`
	// 扫码登录密钥
	QrcodeKey string `json:"qrcodeKey" example:"xxxxxxxxxxxxxxxxxxxx"`
}

// LivePollQRCodeResp 扫码状态轮询结果
type LivePollQRCodeResp struct {
	// 扫码code：0-扫码登录成功，86038-二维码已失效，86090-二维码已扫码未确认，86101-未扫码
	Status int `json:"status" example:"86101"`
	// 返回的提示信息
	Message string `json:"message" example:"等待扫描"`
	// 是否已扫描
	IsScanned bool `json:"isScanned" example:"false"`
	// 是否登录成功
	IsSuccess bool `json:"isSuccess" example:"false"`
	// 是否已过期
	IsExpired bool `json:"isExpired" example:"false"`
}

// LiveLoginStatusResp 机器人登录状态
type LiveLoginStatusResp struct {
	// 是否登录
	IsLoggedIn bool `json:"isLoggedIn" example:"true"`
	// 机器人账号UID
	UID int64 `json:"uid" example:"123456789"`
	// 账号名称
	Username string `json:"username" example:"你的B站昵称"`
	// 账号头像URL
	Face string `json:"face" example:"https://i2.hdslb.com/bfs/face/99a3f6360dff7882059cced5f1912c51cb3dbd71.jpg"`
	// 账号Buvid3
	Buvid string `json:"buvid" example:"XX-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"`
}

// LiveListenerStatusResp 监听器状态与统计
type LiveListenerStatusResp struct {
	// 是否正在监听
	IsRunning bool `json:"isRunning" example:"true"`
	// 房间号
	RoomID int64 `json:"roomId" example:"22384516"`
	// 开始监听事件
	StartTime string `json:"startTime" example:"2026-07-30 12:00:00"`
	// 已监听时长
	Uptime string `json:"uptime" example:"1h30m0s"`
	// 监听到消息数量
	MsgCount int64 `json:"msgCount" example:"1234"`
	// 监听到弹幕数量
	DanmuCount int64 `json:"danmuCount" example:"1000"`
	// 监听到礼物数量
	GiftCount int64 `json:"giftCount" example:"50"`
}
