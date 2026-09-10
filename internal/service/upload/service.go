// Package upload 提供通用的图片上传与 OSS 同步能力
//
// 与业务模块解耦：落盘目录由 scene 白名单决定，OSS 配置从 appconfig 基础设施缓存读取。
// 任意模块把上传接口挂到自己的业务页面上即可复用，无需重复实现落盘与同步逻辑。
package upload

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zxc7563598/bilibili-live-assistant/internal/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/fileutil"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/oss"
)

// MaxRequestSize 上传请求体上限：单文件上限 + 1MB（multipart 头部 / 边界开销）
const MaxRequestSize int64 = fileutil.MaxUploadSize + 1<<20

type Service struct {
	appConfigCache *appconfig.Cache
}

func New(appConfigCache *appconfig.Cache) *Service {
	return &Service{
		appConfigCache: appConfigCache,
	}
}

// UploadImage 校验上传场景与图片内容后落盘，返回可直接访问的相对路径
func (s *Service) UploadImage(_ context.Context, req UploadImageReq) (UploadPathResp, int, error) {
	subDir, ok := resolveScene(req.Scene)
	if !ok {
		return UploadPathResp{}, 11405, fmt.Errorf("未知上传场景: %q", req.Scene)
	}
	if req.File == nil || req.File.Size == 0 {
		return UploadPathResp{}, 11403, errors.New("上传文件为空")
	}
	if !sniffImage(req.File) {
		return UploadPathResp{}, 11402, errors.New("上传内容不是有效图片")
	}
	uploadPath, err := fileutil.SaveUploadedFile(req.File, subDir)
	switch {
	case err == nil:
		return UploadPathResp{Path: uploadPath}, 0, nil
	case errors.Is(err, fileutil.ErrEmpty):
		return UploadPathResp{}, 11403, err
	case errors.Is(err, fileutil.ErrTooLarge):
		return UploadPathResp{}, 11404, err
	case errors.Is(err, fileutil.ErrInvalidDir):
		// 白名单已保证子目录合法，走到这里属开发期传参问题
		return UploadPathResp{}, 11405, err
	default:
		return UploadPathResp{}, 61401, fmt.Errorf("保存上传文件失败: %w", err)
	}
}

// SyncOSS 把已本地落盘的图片同步到阿里云 OSS，返回可直接公开访问的 URL
func (s *Service) SyncOSS(_ context.Context, path string) (UploadPathResp, int, error) {
	localPath, err := fileutil.ResolveLocalPath(path)
	if err != nil {
		return UploadPathResp{}, 11407, fmt.Errorf("OSS 同步的图片路径不合法 %q: %w", path, err)
	}
	cfg, ok := s.ossConfigFromCache()
	if !ok {
		return UploadPathResp{}, 41401, errors.New("OSS 未配置：Endpoint/AccessKey ID/AccessKey Secret/bucket 需完整填写")
	}
	client, err := oss.New(cfg)
	if err != nil {
		return UploadPathResp{}, 61403, fmt.Errorf("OSS 初始化失败: %w", err)
	}
	// 本地文件必须真实存在，OSS 无源可传；已被清理的文件需重新上传
	if _, err := os.Stat(localPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return UploadPathResp{}, 51401, fmt.Errorf("OSS 同步的本地文件不存在: %s", localPath)
		}
		return UploadPathResp{}, 61402, fmt.Errorf("OSS 同步读取本地文件状态失败: %w", err)
	}
	url, err := client.UploadFile(localPath, filepath.ToSlash(localPath))
	if err != nil {
		return UploadPathResp{}, 61404, fmt.Errorf("OSS 上传 %s 失败: %w", localPath, err)
	}
	return UploadPathResp{Path: url}, 0, nil
}
