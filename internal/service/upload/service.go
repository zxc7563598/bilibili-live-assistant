// Package upload 提供通用的图片上传、OSS 同步与旧版数据导入能力
//
// 图片上传与业务模块解耦：落盘目录由 scene 白名单决定，OSS 配置从 appconfig 基础设施缓存读取。
// 数据导入把旧版导出的 .gz 文件还原到 live_danmus / live_gifts / live_users，见 import.go。
package upload

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/zxc7563598/bilibili-live-assistant/internal/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/fileutil"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/oss"
	"gorm.io/gorm"
)

// MaxRequestSize 上传请求体上限：单文件上限 + 1MB（multipart 头部 / 边界开销）
const MaxRequestSize int64 = fileutil.MaxUploadSize + 1<<20

type Service struct {
	appConfigCache *appconfig.Cache
	db             *gorm.DB
	liveDanmuRepo  live_danmu.Repository
	liveGiftRepo   live_gift.Repository
	liveUserRepo   live_user.Repository
	liveUserSvc    *liveuser.Service
	importTasks    *importTaskStore
}

func New(appConfigCache *appconfig.Cache, db *gorm.DB, liveDanmuRepo live_danmu.Repository, liveGiftRepo live_gift.Repository, liveUserRepo live_user.Repository, liveUserSvc *liveuser.Service) *Service {
	return &Service{
		appConfigCache: appConfigCache,
		db:             db,
		liveDanmuRepo:  liveDanmuRepo,
		liveGiftRepo:   liveGiftRepo,
		liveUserRepo:   liveUserRepo,
		liveUserSvc:    liveUserSvc,
		importTasks:    newImportTaskStore(),
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

// StartImport 接收旧版导出的 .gz 文件并投递后台导入任务，返回任务标识
func (s *Service) StartImport(ctx context.Context, req ImportDataReq) (ImportTaskResp, int, error) {
	if req.File == nil || req.File.Size == 0 {
		return ImportTaskResp{}, CodeImportFileRequired, errors.New("导入文件为空")
	}
	if req.File.Size > importFileLimit {
		return ImportTaskResp{}, CodeImportFileTooLarge, fmt.Errorf("导入文件大小 %d 超过上限 %d", req.File.Size, importFileLimit)
	}
	// 先抢单飞锁：任务表在内存里，并发请求能立即拿到拒绝原因，不必等数据库
	taskID, ok := s.importTasks.start(req.AdminID, req.File.Filename, req.File.Size)
	if !ok {
		return ImportTaskResp{}, CodeImportRunning, errors.New("已有导入任务正在进行中")
	}
	// 请求结束后 multipart 临时文件会被清理，先落到自己的临时文件
	path, err := copyUploadedFile(req.File)
	if err != nil {
		s.importTasks.cancel(taskID)
		return ImportTaskResp{}, CodeImportFailed, fmt.Errorf("暂存导入文件失败: %w", err)
	}
	// 空表校验放在任务开跑之前，让调用方立即拿到拒绝原因
	if errCode, err := s.checkImportPrecondition(ctx); errCode != 0 {
		s.importTasks.cancel(taskID)
		os.Remove(path)
		return ImportTaskResp{}, errCode, err
	}
	// 导入在后台跑，必须用与请求生命周期无关的上下文
	go s.runImport(context.Background(), taskID, path)
	return ImportTaskResp{TaskID: taskID}, 0, nil
}

// GetImportProgress 查询导入任务进度，任务已被淘汰时返回 CodeImportTaskNotFound
func (s *Service) GetImportProgress(_ context.Context, taskID string) (ImportProgressResp, int, error) {
	task := s.importTasks.get(taskID)
	if task == nil {
		return ImportProgressResp{}, CodeImportTaskNotFound, errors.New("导入任务不存在或已过期")
	}
	resp := ImportProgressResp{
		TaskID:           task.id,
		Status:           task.status,
		CurrentTable:     task.currentTable,
		DanmuCount:       task.danmuCount,
		GiftCount:        task.giftCount,
		UserCount:        task.userCount,
		CreditLogCount:   task.creditLogCount,
		SkippedUserCount: task.skippedUserCount,
		BytesProcessed:   task.bytesProcessed,
		BytesTotal:       task.bytesTotal,
		ErrorCode:        task.errorCode,
		StartedAt:        task.startedAt,
		FinishedAt:       task.finishedAt,
	}
	if task.bytesTotal > 0 {
		resp.Percent = math.Round(float64(task.bytesProcessed)/float64(task.bytesTotal)*10000) / 100
	}
	return resp, 0, nil
}
