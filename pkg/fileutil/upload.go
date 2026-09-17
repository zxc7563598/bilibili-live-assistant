// Package fileutil 提供上传文件的本地落盘工具，供各业务模块复用。
//
// 文件统一落到 <UploadRoot>/（由 bootstrap 按配置注入，默认工作目录下的 uploads/）下，
// 文件名带纳秒时间戳 + 随机后缀，通过 internal/bootstrap 注册的 /uploads 静态路由对外
// 提供访问。落盘目录与对外 URL 前缀相互独立：目录可配置到任意位置（含绝对路径），
// 对外 URL 始终是 /uploads/...。本包无任何业务依赖。
package fileutil

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// URLPrefix 上传文件的对外访问前缀，与 internal/bootstrap 注册的静态路由保持一致
const URLPrefix = "/uploads"

// defaultUploadRoot 上传文件落盘根目录的默认值
const defaultUploadRoot = "uploads"

// uploadRoot 上传文件落盘根目录，由 SetUploadRoot 在启动时按配置注入
var uploadRoot = defaultUploadRoot

// SetUploadRoot 设置上传文件落盘根目录，传空值时沿用默认值，便于未初始化场景（如单元测试）直接使用
func SetUploadRoot(dir string) {
	if dir != "" {
		uploadRoot = dir
	}
}

// UploadRoot 返回当前上传文件落盘根目录
func UploadRoot() string {
	return uploadRoot
}

// MaxUploadSize 单文件正文大小上限（20MB），超出直接拒绝
const MaxUploadSize int64 = 20 << 20

var (
	// ErrInvalidDir 子目录不合法（为空、绝对路径、目录穿越等），多为调用方传参错误。
	ErrInvalidDir = errors.New("fileutil: 非法上传子目录")
	// ErrEmpty 空文件。
	ErrEmpty = errors.New("fileutil: 空文件")
	// ErrTooLarge 文件超过大小上限。
	ErrTooLarge = errors.New("fileutil: 文件超过大小上限")
	// ErrInvalidUploadPath 不是合法的本地上传访问路径。
	ErrInvalidUploadPath = errors.New("fileutil: 非法上传路径")
)

// SaveUploadedFile 把 multipart 上传文件保存到 <UploadRoot>/<subDir>/ 下，返回可直接访问的相对 URL（/uploads/<subDir>/<文件名>），目录不存在时自动创建
func SaveUploadedFile(file *multipart.FileHeader, subDir string) (string, error) {
	if file == nil {
		return "", ErrEmpty
	}
	sub, ok := safeSubDir(subDir)
	if !ok {
		return "", ErrInvalidDir
	}
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("fileutil: 打开上传文件失败: %w", err)
	}
	defer src.Close()
	targetDir := filepath.Join(uploadRoot, filepath.FromSlash(sub))
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("fileutil: 创建上传目录失败: %w", err)
	}
	name, err := uploadFileName(file.Filename)
	if err != nil {
		return "", err
	}
	absPath := filepath.Join(targetDir, name)
	// O_EXCL：目标已存在则直接失败，绝不覆盖已有文件
	dst, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", fmt.Errorf("fileutil: 创建上传文件失败: %w", err)
	}
	// 边写边限长：multipart 头里的 size 可能与实际不符，不能仅靠它兜底
	n, copyErr := io.Copy(dst, io.LimitReader(src, MaxUploadSize+1))
	if closeErr := dst.Close(); copyErr == nil {
		copyErr = closeErr
	}
	if copyErr != nil {
		_ = os.Remove(absPath)
		return "", fmt.Errorf("fileutil: 写入上传文件失败: %w", copyErr)
	}
	if n == 0 {
		_ = os.Remove(absPath)
		return "", ErrEmpty
	}
	if n > MaxUploadSize {
		_ = os.Remove(absPath)
		return "", ErrTooLarge
	}
	// URL 固定用正斜杠拼接，与落盘目录的平台相关分隔符无关
	return URLPrefix + "/" + path.Join(sub, name), nil
}

// ResolveLocalPath 将上传文件的可访问 URL 路径（如 /uploads/login_bg/<文件名>）
// 换算为落盘根目录下的本地路径（如 uploads/login_bg/<文件名>），供回读 / 同步等场景复用。
// 仅做结构与目录穿越校验，不检查文件是否存在；非 /uploads/ 前缀或无法限定在
// 落盘根目录内时返回 ErrInvalidUploadPath。
func ResolveLocalPath(accessPath string) (string, error) {
	rel, err := accessRel(accessPath)
	if err != nil {
		return "", err
	}
	local := filepath.Join(uploadRoot, filepath.FromSlash(rel))
	// path.Clean 已把 ../ 折叠回根目录，正常到不了这里；保留兜底防御
	root := filepath.Clean(uploadRoot)
	if local == root || !strings.HasPrefix(local, root+string(filepath.Separator)) {
		return "", ErrInvalidUploadPath
	}
	return local, nil
}

// ObjectKey 将上传文件的可访问 URL 路径换算为 OSS objectKey（如 uploads/login_bg/<文件名>）
//
// 只由对外 URL 前缀推导，与落盘根目录的实际位置无关。落盘目录可以配置成任意绝对路径，
// 但入库的 objectKey 必须稳定：否则换个部署目录后同一张图会被当成新对象重复上传，
// 而已入库的老 key 也拼不回公开链接（绝对路径还会把服务器目录结构泄露到 URL 里）。
func ObjectKey(accessPath string) (string, error) {
	rel, err := accessRel(accessPath)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(URLPrefix, "/") + "/" + rel, nil
}

// accessRel 校验可访问 URL 路径并归一化为 <子目录>/<文件名>
func accessRel(accessPath string) (string, error) {
	p := strings.ReplaceAll(strings.TrimSpace(accessPath), "\\", "/")
	// 归一化 ../ 等片段；TrimLeft 兼容调用方省略前导斜杠的写法
	p = path.Clean("/" + strings.TrimLeft(p, "/"))
	rel, ok := strings.CutPrefix(p, URLPrefix+"/")
	if !ok || rel == "" {
		return "", ErrInvalidUploadPath
	}
	return rel, nil
}

// safeSubDir 将调用方传入的子目录净化为相对路径，用于拼装 UploadRoot 下的落盘目录
func safeSubDir(dir string) (string, bool) {
	clean := path.Clean(strings.ReplaceAll(strings.TrimSpace(dir), "\\", "/"))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return "", false
	}
	for _, c := range clean {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case c == '-', c == '_', c == '/':
		default:
			return "", false
		}
	}
	return clean, true
}

// uploadFileName 生成唯一的落盘文件名：<纳秒时间戳>_<8位随机hex>[.安全扩展名]
func uploadFileName(original string) (string, error) {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("fileutil: 生成随机文件名失败: %w", err)
	}
	ext := strings.ToLower(path.Ext(strings.TrimSpace(original)))
	for i := 1; i < len(ext); i++ {
		c := ext[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
			ext = ""
			break
		}
	}
	if len(ext) < 2 || len(ext) > 6 {
		ext = ""
	}
	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), hex.EncodeToString(buf), ext), nil
}
