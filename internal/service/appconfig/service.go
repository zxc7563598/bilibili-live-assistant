package appconfig

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/app_config"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/imagetype"
)

// 站点基础配置键（与 internal/migrate/seed.go seedAppConfigs 保持一致）
const (
	keySiteName            = "site_name"
	keySiteDescription     = "site_description"
	keySiteBackgroundColor = "site_background_color"
	keySiteThemeColor      = "site_theme_color"
	keySiteIcon            = "site_icon"
)

type Service struct {
	appConfigCache *appconfig.Cache
	appConfigRepo  app_config.Repository
}

func New(appConfigCache *appconfig.Cache, appConfigRepo app_config.Repository) *Service {
	return &Service{
		appConfigCache: appConfigCache,
		appConfigRepo:  appConfigRepo,
	}
}

// Manifest 组装 PWA manifest 配置
//
// 配置缺失时不强校验、直接返回空值：
//   - 文本字段（站点名、描述、颜色）允许为空，由前端自行兜底；
//   - 站点图标未配置时跳过 MIME 检测，避免整份 manifest 因缺图标而失败；
//   - 图标已配置但无法识别时返回 10901，便于定位配置问题。
func (s *Service) Manifest() (ManifestResp, int, error) {
	resp := ManifestResp{
		Name:            s.configValue(keySiteName),
		Description:     s.configValue(keySiteDescription),
		BackgroundColor: s.configValue(keySiteBackgroundColor),
	}
	icon := s.configValue(keySiteIcon)
	if icon == "" {
		return resp, 0, nil
	}
	mimeType, err := imagetype.GetMimeTypeSimple(icon)
	if err != nil {
		return ManifestResp{}, 10901, err
	}
	resp.Icon = icon
	resp.IconType = mimeType
	return resp, 0, nil
}
