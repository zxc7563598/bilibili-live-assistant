package region

import (
	"encoding/json"
	"strings"
)

// Node 对应 regions.json 中的节点：
// value 为 6 位行政区划 code，label 为中文名，level 0=省 1=市 2=区县，children 为下级。
type Node struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Level    int    `json:"level"`
	Children []Node `json:"children"`
}

// provinces 在包初始化时解析嵌入的省节点列表（只读，进程内共享）。
var provinces []Node

func init() {
	// 嵌入数据为编译期静态内容，解析失败属于构建错误，直接崩溃暴露，避免运行时所有地区校验静默失败
	if err := json.Unmarshal(regionsJSON, &provinces); err != nil {
		panic("region: 解析嵌入的 regions.json 失败: " + err.Error())
	}
}

// Resolve 校验 codes 是否为 regions.json 中一条「省 → … → 下级区域」的合法子链（至少 2 段），
// 合法则返回各节点 label 空格拼接的地区文案。
//
//	Resolve([]string{"370000", "370100", "370116"}) // "山东省 济南市 莱芜区", true
//	Resolve([]string{"440000", "441900"})           // "广东省 东莞市", true（无区县的地级市仅 2 段）
func Resolve(codes []string) (string, bool) {
	if len(codes) < 2 {
		return "", false
	}

	roots := provinces
	var cur *Node
	labels := make([]string, 0, len(codes))

	for _, code := range codes {
		var next *Node
		if cur == nil {
			// 首段必须在省节点中找到
			for i := range roots {
				if roots[i].Value == code {
					next = &roots[i]
					break
				}
			}
		} else {
			// 后续段必须是上一节点的直接子级
			for i := range cur.Children {
				if cur.Children[i].Value == code {
					next = &cur.Children[i]
					break
				}
			}
		}
		if next == nil {
			return "", false
		}
		cur = next
		labels = append(labels, next.Label)
	}
	return strings.Join(labels, " "), true
}
