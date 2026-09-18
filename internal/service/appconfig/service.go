package appconfig

import (
	"context"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/app_config"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/imagetype"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/oss"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
)

// 站点基础配置键
// OSS 相关键由 internal/appconfig 统一提供（上传模块也要读），见 appconfig.KeyOss*
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

// GetManifest 组装 PWA manifest 配置
//
// 配置缺失时不强校验、直接返回空值：
//   - 文本字段（站点名、描述、颜色）允许为空，由前端自行兜底
//   - 站点图标未配置时跳过 MIME 检测，避免整份 manifest 因缺图标而失败
//   - 图标已配置但无法识别时返回 CodeImageInvalid，便于定位配置问题
func (s *Service) GetManifest() (ManifestResp, int, error) {
	resp := ManifestResp{
		Name:            s.appConfigCache.GetValue(keySiteName),
		Description:     s.appConfigCache.GetValue(keySiteDescription),
		BackgroundColor: s.appConfigCache.GetValue(keySiteBackgroundColor),
	}
	icon := s.appConfigCache.GetValue(keySiteIcon)
	if icon == "" {
		return resp, 0, nil
	}
	mimeType, err := imagetype.GetMimeTypeSimple(icon)
	if err != nil {
		return ManifestResp{}, CodeImageInvalid, err
	}
	resp.Icon = icon
	resp.IconType = mimeType
	return resp, 0, nil
}

// GetThemeColor 获取网站主题色
func (s *Service) GetThemeColor() (string, int, error) {
	return s.appConfigCache.GetValue(keySiteThemeColor), 0, nil
}

// GetLoginConfig 获取登录页配置
func (s *Service) GetLoginConfig() (LoginConfig, int, error) {
	resp := LoginConfig{
		Logo:     s.appConfigCache.GetValue(keyLogo),
		LoginBg:  s.appConfigCache.GetValue(keyLoginBg),
		Title:    s.appConfigCache.GetValue(keyTitle),
		Slogan:   s.appConfigCache.GetValue(keySlogan),
		Register: false,
	}
	register := ptr.ParseEnumInt[enum.YesNo](s.appConfigCache.GetValue(keyRegister))
	if register == enum.Yes {
		resp.Register = true
	}
	return resp, 0, nil
}

// GetConfigData 获取全部配置信息
func (s *Service) GetConfigData() (ConfigDataResp, int, error) {
	data := s.appConfigCache.GetAll()
	return ConfigDataResp{
		SiteName:            data[keySiteName],
		SiteDescription:     data[keySiteDescription],
		SiteBackgroundColor: data[keySiteBackgroundColor],
		SiteThemeColor:      data[keySiteThemeColor],
		SiteIcon:            data[keySiteIcon],
		Register:            data[keyRegister],
		Logo:                data[keyLogo],
		LoginBg:             data[keyLoginBg],
		LoginTitle:          data[keyTitle],
		LoginSlogan:         data[keySlogan],
		OssEndpoint:         data[appconfig.KeyOssEndpoint],
		OssAccessKeyId:      data[appconfig.KeyOssAccessKeyId],
		OssAccessKeySecret:  data[appconfig.KeyOssAccessKeySecret],
		OssBucket:           data[appconfig.KeyOssBucket],
	}, 0, nil
}

// SaveConfig 保存基础配置
func (s *Service) SaveConfig(ctx context.Context, data SaveConfigReq) (int, error) {
	values := map[string]string{
		keySiteName:            data.SiteName,
		keySiteDescription:     data.SiteDescription,
		keySiteBackgroundColor: data.SiteBackgroundColor,
		keySiteThemeColor:      data.SiteThemeColor,
		keySiteIcon:            data.SiteIcon,
		keyRegister:            data.Register,
		keyLogo:                data.Logo,
		keyLoginBg:             data.LoginBg,
		keyTitle:               data.LoginTitle,
		keySlogan:              data.LoginSlogan,
	}
	if err := s.appConfigRepo.SaveValues(ctx, nil, values); err != nil {
		return CodeSaveFailed, fmt.Errorf("保存 App 配置失败: %w", err)
	}
	// 落库成功后刷新缓存
	reloadCtx := context.WithoutCancel(ctx)
	if err := s.appConfigCache.Reload(reloadCtx); err != nil {
		return CodeCacheReloadFailed, fmt.Errorf("刷新 App 配置缓存失败: %w", err)
	}
	return 0, nil
}

// SaveOssConfig 保存OSS配置
func (s *Service) SaveOssConfig(ctx context.Context, data SaveOssConfigReq) (int, error) {
	// 验证配置是否可靠
	oss, err := oss.New(oss.Config{
		Endpoint:        data.OssEndpoint,
		AccessKeyID:     data.OssAccessKeyId,
		AccessKeySecret: data.OssAccessKeySecret,
		Bucket:          data.OssBucket,
	})
	if err != nil {
		return CodeOSSInitFailed, fmt.Errorf("OSS 初始化失败: %w", err)
	}
	if err := oss.CheckConfig(); err != nil {
		return CodeOSSConfigFailed, fmt.Errorf("OSS 验证失败: %w", err)
	}
	// 存储信息
	values := map[string]string{
		appconfig.KeyOssEndpoint:        data.OssEndpoint,
		appconfig.KeyOssAccessKeyId:     data.OssAccessKeyId,
		appconfig.KeyOssAccessKeySecret: data.OssAccessKeySecret,
		appconfig.KeyOssBucket:          data.OssBucket,
	}
	if err := s.appConfigRepo.SaveValues(ctx, nil, values); err != nil {
		return CodeSaveFailed, fmt.Errorf("保存 App 配置失败: %w", err)
	}
	// 落库成功后刷新缓存
	reloadCtx := context.WithoutCancel(ctx)
	if err := s.appConfigCache.Reload(reloadCtx); err != nil {
		return CodeCacheReloadFailed, fmt.Errorf("刷新 App 配置缓存失败: %w", err)
	}
	return 0, nil
}
