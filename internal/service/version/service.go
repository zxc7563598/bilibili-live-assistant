package version

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/version"
)

// Service 版本服务
type Service struct{}

// New 创建 Service 实例
func New() *Service {
	return &Service{}
}

// GetVersionInfo 获取当前版本与远端最新版本；检查失败只标记不报错
func (s *Service) GetVersionInfo() VersionInfo {
	latest, needUpdate, err := version.CheckUpdateCached()
	if err != nil {
		return VersionInfo{
			Version:     version.Version,
			Commit:      version.Commit,
			CheckFailed: true,
		}
	}
	return VersionInfo{
		Version:    version.Version,
		Commit:     version.Commit,
		Latest:     latest,
		NeedUpdate: needUpdate,
	}
}
