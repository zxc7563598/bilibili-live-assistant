// Package wbi 提供 B站 WBI 签名工具。
// 纯算法实现，无任何外部依赖，不发起 HTTP 请求。
//
// WBI 签名用于以下 B站 API：
//   - getDanmuInfo（获取 WebSocket 连接信息）
//   - getStreamerInfo（获取用户空间信息）
//
// 使用流程：
//  1. 从 /x/web-interface/nav 接口获取 wbi_img.img_url 和 wbi_img.sub_url
//  2. 提取 URL 中文件名（不含扩展名）作为 imgKey 和 subKey
//  3. 调用 wbi.Sign(params, imgKey, subKey) 得到签名后的 query string
package wbi

import (
	"crypto/md5"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// mixinKeyTable 是 B站 WBI 算法固定的 64 位乱序索引表。
var mixinKeyTable = [64]int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35,
	27, 43, 5, 49, 33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13,
	37, 48, 7, 16, 24, 55, 40, 61, 26, 17, 0, 1, 60, 51, 30, 4,
	22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11, 36, 20, 34, 44, 52,
}

// getMixinKey 从 imgKey + subKey 拼接后的字符串中取出 64 位 mixin key。
func getMixinKey(combinedKey string) string {
	var sb strings.Builder
	for _, idx := range mixinKeyTable {
		sb.WriteByte(combinedKey[idx])
	}
	return sb.String()[:32]
}

// ExtractKeys 从 nav 接口返回的 wbi_img URL 中提取 imgKey 和 subKey。
// imgURL 和 subURL 格式如 "https://i0.hdslb.com/bfs/wbi/653657f524dd2f7...png"
// 提取文件名中不含扩展名的部分作为 key。
func ExtractKeys(imgURL, subURL string) (imgKey, subKey string) {
	imgKey = extractKeyFromURL(imgURL)
	subKey = extractKeyFromURL(subURL)
	return
}

func extractKeyFromURL(rawURL string) string {
	// 获取 URL path 的最后一段（文件名）
	idx := strings.LastIndex(rawURL, "/")
	if idx < 0 {
		return rawURL
	}
	filename := rawURL[idx+1:]
	// 去掉扩展名
	dotIdx := strings.LastIndex(filename, ".")
	if dotIdx >= 0 {
		return filename[:dotIdx]
	}
	return filename
}

// Sign 对参数进行 WBI 签名。
//
// params 为需要签名的原始参数（不含 wts），
// imgKey 和 subKey 由 ExtractKeys 从 nav 接口提取。
//
// 返回完整的 query string（已包含 wts 和 w_rid 参数），
// 直接拼接到 URL 后面即可：url + "?" + signedQuery
func Sign(params map[string]string, imgKey, subKey string) string {
	mixinKey := getMixinKey(imgKey + subKey)
	params["wts"] = strconv.FormatInt(time.Now().Unix(), 10)
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	reg := regexp.MustCompile(`[!'()*]`)
	var parts []string
	for _, k := range keys {
		value := reg.ReplaceAllString(params[k], "")

		parts = append(parts,
			url.QueryEscape(k)+"="+url.QueryEscape(value),
		)
	}
	query := strings.Join(parts, "&")
	hash := md5.Sum([]byte(query + mixinKey))
	wrid := fmt.Sprintf("%x", hash)
	return query + "&w_rid=" + wrid
}
