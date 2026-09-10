package config

import (
	"log"
	"os"
	"path/filepath"
)

// resolveRuntimeDirs 把配置里的运行目录统一解析为绝对路径
//
// 相对路径以**配置文件所在目录**为基准，而不是进程工作目录：以 systemd 启动且未配
// WorkingDirectory 时工作目录是 /，Windows 注册成服务时是 System32，从不同目录启动
// 同一个二进制也会得到不同的工作目录。这些情况下要么写文件直接失败（无权限），
// 要么写入了一个和读取时不一致的位置，表现为「换个目录启动后已有文件全部读不到」。
//
// 不用「相对可执行文件目录」是因为 go run 会把二进制放在系统临时目录，
// 开发时（make dev-go）文件会散落到临时目录里。
func resolveRuntimeDirs(cfg *Config, configPath string) {
	base := ""
	if abs, err := filepath.Abs(configPath); err == nil {
		base = filepath.Dir(abs)
	}
	cfg.File.UploadDir = resolveDir(cfg.File.UploadDir, base)
	cfg.Log.Dir = resolveDir(cfg.Log.Dir, base)
	// 登录态文件允许为空（表示不落盘），为空时原样返回
	cfg.Live.StateFile = resolveDir(cfg.Live.StateFile, base)
}

// resolveDir 解析单个目录或文件路径
//
// 绝对路径原样保留；相对路径展开为 baseDir 下的绝对路径。
// 兼容旧版本：旧版本把相对路径当作工作目录下的路径使用，若新位置还不存在、
// 而工作目录下的旧位置已有内容，则沿用旧位置，避免升级后已有文件读不到。
func resolveDir(target, baseDir string) string {
	if target == "" {
		return ""
	}
	if filepath.IsAbs(target) {
		return filepath.Clean(target)
	}
	abs, err := filepath.Abs(filepath.Join(baseDir, target))
	if err != nil {
		return target
	}
	legacy, legacyErr := filepath.Abs(target)
	if legacyErr == nil && legacy != abs && !pathExists(abs) && pathExists(legacy) {
		log.Printf("[配置] %s 按配置文件所在目录解析为 %s，该位置尚不存在；沿用工作目录下已有的 %s", target, abs, legacy)
		return legacy
	}
	return abs
}

// pathExists 判断路径是否存在
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
