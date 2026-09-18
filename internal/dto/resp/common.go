package resp

// RoomGroupOptionsResp 按房间号筛选的下拉选项返回
//
// 弹幕、礼物、PK 三个列表页的「房间」筛选框共用这一份结构。
type RoomGroupOptionsResp struct {
	// 房间下拉选项
	Option []RoomGroupOptionItem `json:"option"`
}

// RoomGroupOptionItem 房间下拉选项中的一项
type RoomGroupOptionItem struct {
	// 房间号文本（前端下拉要求 label 为字符串）
	Label string `json:"label" example:"22384516"`
	// 房间号
	Value int64 `json:"value" example:"22384516"`
}
