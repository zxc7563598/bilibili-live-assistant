package input

// LiveQRCodePollReq 轮询 B站 扫码登录状态
type LiveQRCodePollReq struct {
	// 扫码登录密钥
	QrcodeKey string `json:"qrcodeKey" binding:"required" err:"required=10402" example:"xxxxxxxxxxxxxxxxxxxx"`
}

// LiveRoomUpdateReq 更新监听房间号
type LiveRoomUpdateReq struct {
	// 房间ID
	RoomID int64 `json:"roomId" binding:"required,min=1" err:"required=10403,min=10404" example:"22384516"`
}

// SendDanmuReq 以机器人身份发送一条弹幕请求
type SendDanmuReq struct {
	// 弹幕内容，长度 1-40 个字符
	Message string `json:"message" binding:"required,min=1,max=40" err:"required=10405,min=10406,max=10407" example:"发送弹幕信息"`
}
