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
