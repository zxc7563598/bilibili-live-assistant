package upload

import (
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/oss"
)

// uploadSceneAllow 图片上传场景白名单
var uploadSceneAllow = map[string]string{
	"login_bg":  "login_bg",         // 登录页背景图（建议 20:9）
	"site_icon": "site_icon",        // 网站图标（建议 1:1，≤512x512）
	"logo":      "logo",             // 网站 logo（建议 1:1，≤512x512）
	"cover":     "product/cover",    // 商品封面（建议 1:1）
	"carousel":  "product/carousel", // 商品轮播图（建议 1:1）
	"details":   "product/details",  // 商品详情图（建议等宽长图）
}

// resolveScene 校验上传场景，返回落盘子目录
func resolveScene(scene string) (string, bool) {
	subDir, ok := uploadSceneAllow[scene]
	return subDir, ok
}

// sniffImage 读取上传文件头部做内容嗅探，确认其真实类型为图片
func sniffImage(file *multipart.FileHeader) bool {
	src, err := file.Open()
	if err != nil {
		return false
	}
	defer src.Close()
	buf := make([]byte, 512)
	n, _ := src.Read(buf)
	if n == 0 {
		return false
	}
	return strings.HasPrefix(http.DetectContentType(buf[:n]), "image/")
}

// ossConfigFromCache 从配置缓存读取 OSS 四项配置，任一缺失视为尚未完整配置
func (s *Service) ossConfigFromCache() (oss.Config, bool) {
	cfg := oss.Config{
		Endpoint:        s.appConfigCache.GetValue(appconfig.KeyOssEndpoint),
		AccessKeyID:     s.appConfigCache.GetValue(appconfig.KeyOssAccessKeyId),
		AccessKeySecret: s.appConfigCache.GetValue(appconfig.KeyOssAccessKeySecret),
		Bucket:          s.appConfigCache.GetValue(appconfig.KeyOssBucket),
	}
	complete := cfg.Endpoint != "" && cfg.AccessKeyID != "" && cfg.AccessKeySecret != "" && cfg.Bucket != ""
	return cfg, complete
}
