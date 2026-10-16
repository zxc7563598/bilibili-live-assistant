package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)

const (
	latestVersionAPI = "https://tools.api.hejunjie.life/bilibilidanmu-api/get-last-version"
	// 远端请求超时时间
	requestTimeout = 5 * time.Second
	// 检查成功的缓存时长
	successCacheTTL = time.Hour
	// 检查失败的缓存时长，比成功短，网络恢复后能尽快提示
	failureCacheTTL = time.Minute
)

// 检查结果的进程内缓存
var (
	cacheMu     sync.Mutex
	cacheLatest string
	cacheErr    error
	cacheExpire time.Time
)

// CheckUpdateCached 检查是否有新版本，结果带缓存
// 返回: (最新版本号, 是否需要更新, error)
func CheckUpdateCached() (string, bool, error) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if time.Now().Before(cacheExpire) {
		return cacheResult()
	}
	latest, err := fetchLatest()
	ttl := successCacheTTL
	if err != nil {
		// 只在真正发起检查时记一次
		logger.VersionLogger.Warn("检查更新失败", zap.Error(err))
		ttl = failureCacheTTL
	}
	cacheLatest, cacheErr, cacheExpire = latest, err, time.Now().Add(ttl)
	return cacheResult()
}

// cacheResult 按缓存内容返回，失败时不能让空版本号参与比较
func cacheResult() (string, bool, error) {
	if cacheErr != nil {
		return "", false, cacheErr
	}
	return cacheLatest, cacheLatest != Version, nil
}

// fetchLatest 请求远端版本接口
func fetchLatest() (string, error) {
	client := &http.Client{Timeout: requestTimeout}
	resp, err := client.Post(latestVersionAPI, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("请求版本接口失败: %w", err)
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
		return "", fmt.Errorf("解析版本信息失败: %w", err)
	}
	// 检查返回码
	if result.Code != 0 {
		return "", fmt.Errorf("获取版本失败: %s", result.Message)
	}
	return result.Data.Version, nil
}
