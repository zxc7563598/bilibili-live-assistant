package appconfig

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/app_config"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/imagetype"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
)

// 站点基础配置键
const (
	keySiteName            = "site_name"
	keySiteDescription     = "site_description"
	keySiteBackgroundColor = "site_background_color"
	keySiteThemeColor      = "site_theme_color"
	keySiteIcon            = "site_icon"
	keyRegister            = "register"
	keyLogo                = "logo"
	keyLoginBg             = "login_bg"
	keyTitle               = "login_title"
	keySlogan              = "login_slogan"
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
//   - 文本字段（站点名、描述、颜色）允许为空，由前端自行兜底
//   - 站点图标未配置时跳过 MIME 检测，避免整份 manifest 因缺图标而失败
//   - 图标已配置但无法识别时返回 10901，便于定位配置问题
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

// ThemeColor 获取网站主题色
func (s *Service) ThemeColor() (string, int, error) {
	return s.configValue(keySiteThemeColor), 0, nil
}

// LoginConfig 获取登录页配置
func (s *Service) LoginConfig() (LoginConfig, int, error) {
	resp := LoginConfig{
		Logo:     s.configValue(keyLogo),
		LoginBg:  s.configValue(keyLoginBg),
		Title:    s.configValue(keyTitle),
		Slogan:   s.configValue(keySlogan),
		Register: false,
	}
	register := ptr.ParseEnumInt[enum.YesNo](s.configValue(keyRegister))
	if register == enum.Yes {
		resp.Register = true
	}
	return resp, 0, nil
}
