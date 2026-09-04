package region

import _ "embed"

// regions.json 为 shop/src/data/regions.json 的副本，随二进制一起嵌入，
// 保证单文件运行时也能独立校验地区编码。
//
//go:embed regions.json
var regionsJSON []byte
