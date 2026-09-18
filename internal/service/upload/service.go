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
		return UploadPathResp{}, CodeSceneInvalid, fmt.Errorf("未知上传场景: %q", req.Scene)
	}
	if req.File == nil || req.File.Size == 0 {
		return UploadPathResp{}, CodeFileRequired, errors.New("上传文件为空")
	}
	if !sniffImage(req.File) {
		return UploadPathResp{}, CodeNotImage, errors.New("上传内容不是有效图片")
	}
	uploadPath, err := fileutil.SaveUploadedFile(req.File, subDir)
	switch {
	case err == nil:
		return UploadPathResp{Path: uploadPath}, 0, nil
	case errors.Is(err, fileutil.ErrEmpty):
		return UploadPathResp{}, CodeFileRequired, err
	case errors.Is(err, fileutil.ErrTooLarge):
		return UploadPathResp{}, CodeFileTooLarge, err
	case errors.Is(err, fileutil.ErrInvalidDir):
		// 白名单已保证子目录合法，走到这里属开发期传参问题
		return UploadPathResp{}, CodeSceneInvalid, err
	default:
		return UploadPathResp{}, CodeSaveFailed, fmt.Errorf("保存上传文件失败: %w", err)
	}
}

// SyncOSS 把已本地落盘的图片同步到阿里云 OSS，返回可直接公开访问的 URL
func (s *Service) SyncOSS(_ context.Context, path string) (UploadPathResp, int, error) {
	localPath, err := fileutil.ResolveLocalPath(path)
	if err != nil {
		return UploadPathResp{}, CodePathInvalid, fmt.Errorf("OSS 同步的图片路径不合法 %q: %w", path, err)
	}
	// objectKey 由访问路径推导，不用本地路径：落盘目录可配置为绝对路径，
	// 直接拿本地路径当 key 会把服务器目录结构写进 OSS 对象名
	objectKey, err := fileutil.ObjectKey(path)
	if err != nil {
		return UploadPathResp{}, CodePathInvalid, fmt.Errorf("OSS 同步的图片路径不合法 %q: %w", path, err)
	}
	cfg, ok := s.ossConfigFromCache()
	if !ok {
		return UploadPathResp{}, CodeOSSNotConfigured, errors.New("OSS 未配置：Endpoint/AccessKey ID/AccessKey Secret/bucket 需完整填写")
	}
	client, err := oss.New(cfg)
	if err != nil {
		return UploadPathResp{}, CodeOSSInitFailed, fmt.Errorf("OSS 初始化失败: %w", err)
	}
	// 本地文件必须真实存在，OSS 无源可传；已被清理的文件需重新上传
	if _, err := os.Stat(localPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return UploadPathResp{}, CodeLocalFileNotFound, fmt.Errorf("OSS 同步的本地文件不存在: %s", localPath)
		}
		return UploadPathResp{}, CodeReadFailed, fmt.Errorf("OSS 同步读取本地文件状态失败: %w", err)
	}
	url, err := client.UploadFile(localPath, objectKey)
	if err != nil {
		return UploadPathResp{}, CodeOSSUploadFailed, fmt.Errorf("OSS 上传 %s 失败: %w", objectKey, err)
	}
	return UploadPathResp{Path: url}, 0, nil
}
