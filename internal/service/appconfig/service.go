package appconfig

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zxc7563598/bilibili-live-assistant/internal/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/app_config"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/fileutil"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/imagetype"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/oss"
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
	keyOssEndpoint         = "oss_endpoint"
	keyOssAccessKeyId      = "oss_access_key_id"
	keyOssAccessKeySecret  = "oss_access_key_secret"
	keyOssBucket           = "oss_bucket"
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

// ConfigData 获取全部配置信息
func (s *Service) ConfigData() (ConfigDataResp, int, error) {
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
		OssEndpoint:         data[keyOssEndpoint],
		OssAccessKeyId:      data[keyOssAccessKeyId],
		OssAccessKeySecret:  data[keyOssAccessKeySecret],
		OssBucket:           data[keyOssBucket],
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
		return 60902, fmt.Errorf("保存 App 配置失败: %w", err)
	}
	// 落库成功后刷新缓存
	reloadCtx := context.WithoutCancel(ctx)
	if err := s.appConfigCache.Reload(reloadCtx); err != nil {
		return 60903, fmt.Errorf("刷新 App 配置缓存失败: %w", err)
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
		return 60905, fmt.Errorf("OSS 初始化失败: %w", err)
	}
	if err := oss.CheckConfig(); err != nil {
		return 60906, fmt.Errorf("OSS 验证失败: %w", err)
	}
	// 存储信息
	values := map[string]string{
		keyOssEndpoint:        data.OssEndpoint,
		keyOssAccessKeyId:     data.OssAccessKeyId,
		keyOssAccessKeySecret: data.OssAccessKeySecret,
		keyOssBucket:          data.OssBucket,
	}
	if err := s.appConfigRepo.SaveValues(ctx, nil, values); err != nil {
		return 60902, fmt.Errorf("保存 App 配置失败: %w", err)
	}
	// 落库成功后刷新缓存
	reloadCtx := context.WithoutCancel(ctx)
	if err := s.appConfigCache.Reload(reloadCtx); err != nil {
		return 60903, fmt.Errorf("刷新 App 配置缓存失败: %w", err)
	}
	return 0, nil
}

// SyncOSS 把已本地落盘的图片同步到阿里云 OSS，返回可直接公开访问的 URL
func (s *Service) SyncOSS(_ context.Context, path string) (string, int, error) {
	localPath, err := fileutil.ResolveLocalPath(path)
	if err != nil {
		return "", 10906, fmt.Errorf("OSS 同步的图片路径不合法 %q: %w", path, err)
	}
	cfg, ok := s.ossConfigFromCache()
	if !ok {
		return "", 40901, errors.New("OSS 未配置：Endpoint/AccessKey ID/AccessKey Secret/bucket 需完整填写")
	}
	client, err := oss.New(cfg)
	if err != nil {
		return "", 60905, fmt.Errorf("OSS 初始化失败: %w", err)
	}
	// 本地文件必须真实存在，OSS 无源可传；已被清理的文件需重新上传
	if _, err := os.Stat(localPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", 50901, fmt.Errorf("OSS 同步的本地文件不存在: %s", localPath)
		}
		return "", 60904, fmt.Errorf("OSS 同步读取本地文件状态失败: %w", err)
	}
	url, err := client.UploadFile(localPath, filepath.ToSlash(localPath))
	if err != nil {
		return "", 60907, fmt.Errorf("OSS 上传 %s 失败: %w", localPath, err)
	}
	return url, 0, nil
}

// ossConfigFromCache 从配置缓存读取 OSS 四项配置，任一缺失视为尚未完整配置
func (s *Service) ossConfigFromCache() (oss.Config, bool) {
	cfg := oss.Config{
		Endpoint:        s.configValue(keyOssEndpoint),
		AccessKeyID:     s.configValue(keyOssAccessKeyId),
		AccessKeySecret: s.configValue(keyOssAccessKeySecret),
		Bucket:          s.configValue(keyOssBucket),
	}
	complete := cfg.Endpoint != "" && cfg.AccessKeyID != "" && cfg.AccessKeySecret != "" && cfg.Bucket != ""
	return cfg, complete
}
