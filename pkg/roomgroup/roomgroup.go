// Package roomgroup 提供「按房间号筛选」下拉框的通用选项结构。
//
// 原先 livedanmu / livegift / livepk 三个 service 包各自声明了一份完全相同的
// FetchRoomGroupsResp，各自写了一份一模一样的 id→选项 转换，三个 handler 包又
// 各自复制了一份对应的响应结构与转换函数。这里收敛为一份。
package roomgroup

import "strconv"

// Option 下拉选项：Value 是房间号本身，Label 是它的十进制文本。
//
// Label 与 Value 内容相同但类型不同（前端下拉组件要求 label 为字符串、
// value 为数值），所以两个字段都要给。
type Option struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// FromIDs 把房间号列表转成下拉选项，保持传入顺序。
func FromIDs(ids []int64) []Option {
	options := make([]Option, 0, len(ids))
	for _, id := range ids {
		options = append(options, Option{
			Label: strconv.FormatInt(id, 10),
			Value: id,
		})
	}
	return options
}
