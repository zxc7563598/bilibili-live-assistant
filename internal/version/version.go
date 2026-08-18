package version

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)

// CheckUpdate 检查是否有新版本
// 返回: (最新版本号, 是否需要更新, error)
func CheckUpdate() (string, bool, error) {
	// 请求远程版本
	resp, err := http.Post("https://tools.api.hejunjie.life/bilibilidanmu-api/get-last-version", "application/json", nil)
	if err != nil {
		return "", false, fmt.Errorf("请求版本接口失败: %w", err)
	}
	defer resp.Body.Close()
	// 解析 JSON
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Version string `json:"version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", false, fmt.Errorf("解析版本信息失败: %w", err)
	}
	// 检查返回码
	if result.Code != 0 {
		return "", false, fmt.Errorf("获取版本失败: %s", result.Message)
	}
	remoteVersion := result.Data.Version
	needUpdate := remoteVersion != Version
	return remoteVersion, needUpdate, nil
}
