package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// danmuSendLogDir 弹幕发送记录根目录
	danmuSendLogDir = "logs/弹幕发送记录"
	// danmuSendLogMaxLines 每个文件最大行数
	danmuSendLogMaxLines = 1000
)

// DanmuSendLogger 弹幕发送记录器。
//
// 特性：
//   - 日志写入独立目录 logs/弹幕发送记录/
//   - 按日期分子目录（YYYY-MM-DD）
//   - 每 1000 行自动分割新文件（编号递增）
//   - 线程安全
//
// 文件命名：danmu_send_{编号}.log
// 每行格式：时间  房间号  弹幕内容
type DanmuSendLogger struct {
	mu        sync.Mutex
	fileNum   int
	lineCount int
	file      *os.File
	date      string // 当前日期 "2006-01-02"
}

// NewDanmuSendLogger 创建弹幕发送记录器。
func NewDanmuSendLogger() *DanmuSendLogger {
	return &DanmuSendLogger{}
}

// Log 记录一条弹幕发送。
func (l *DanmuSendLogger) Log(roomID int64, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	today := time.Now().Format("2006-01-02")

	// 日期变了 → 关闭旧文件，重置编号从 1 开始
	if l.date != today {
		if l.file != nil {
			l.file.Close()
			l.file = nil
		}
		l.date = today
		l.fileNum = 0
		l.lineCount = 0
	}

	// 需要打开/切换文件
	if l.file == nil || l.lineCount >= danmuSendLogMaxLines {
		if l.file != nil {
			l.file.Close()
		}
		l.fileNum++
		l.lineCount = 0

		dir := filepath.Join(danmuSendLogDir, today)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return // 目录创建失败，静默丢弃本条日志
		}

		filename := fmt.Sprintf("danmu_send_%d.log", l.fileNum)
		path := filepath.Join(dir, filename)

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return // 文件打开失败，静默丢弃
		}
		l.file = f
	}

	// 写入：时间  房间号  弹幕内容
	line := fmt.Sprintf("%s   %d   %s\n", time.Now().Format("2006-01-02 15:04:05.000"), roomID, message)
	l.file.WriteString(line)
	l.lineCount++
}

// Close 关闭文件。
func (l *DanmuSendLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}
