package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// newRotatingWriter 按模块日志的既有约定构建轮转写入器
//
// 落盘路径与文件名规则和模块 logger 保持一致：<日志根目录>/<subDir>/<YYYY-MM-DD>_<name>.log，
// 轮转参数沿用既有取值（单文件 100MB / 保留 30 个 / 7 天 / 压缩）。
// 目录创建失败时返回 error，由调用方决定降级方式。
func newRotatingWriter(subDir, name string) (io.Writer, error) {
	dir := filepath.Join(logDir, subDir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("无法创建日志目录: %w", err)
	}
	// 按天分割日志文件
	filename := filepath.Join(dir, fmt.Sprintf("%s_%s.log", time.Now().Format(time.DateOnly), name))
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100, // MB
		MaxBackups: 30,  // 保留最近30个日志文件
		MaxAge:     7,   // 天
		Compress:   true,
	}, nil
}

// lockWriter 给并发写入的 writer 加锁
//
// lumberjack 自身会串行化对单个文件的写入，但 io.MultiWriter 在「文件 + 终端」两路之间没有锁。
// Gin 的日志中间件会从多个请求 goroutine 并发写同一个 writer，不加锁会出现跨路错序
// （A 的文件行、B 的文件行、B 的终端行、A 的终端行）。
type lockWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (l *lockWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}
