package handler

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/roomgroup"
)

// 跨模块共用的 DTO 转换。
//
// 只放被两个以上 handler 包用到的转换；单个模块自己的 toXxx 仍放在各模块的
// common.go 里。

// ToRoomGroupOptions 把房间号下拉选项（service 出参）转成响应结构。
//
// 弹幕、礼物、PK 三个列表页的房间筛选共用同一份出参与响应结构（见
// pkg/roomgroup 与 resp.RoomGroupOptionsResp），转换也在这里共用一份。
func ToRoomGroupOptions(list []roomgroup.Option) []resp.RoomGroupOptionItem {
	res := make([]resp.RoomGroupOptionItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.RoomGroupOptionItem{
			Label: v.Label,
			Value: v.Value,
		})
	}
	return res
}
