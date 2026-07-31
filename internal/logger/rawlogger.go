package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// rawLogDir 原始消息日志根目录
	rawLogDir = "logs/直播间监听原始信息"
	// rawLogMaxLines 每个文件最大行数
	rawLogMaxLines = 1000
)

// RawMessageLogger 直播间原始消息日志记录器。
//
// 特性：
//   - 日志写入独立目录 logs/直播间监听原始信息/
//   - 按日期分子目录（YYYY-MM-DD）
//   - 按 msg.Cmd 分文件（每种消息类型独立文件）
//   - 每 1000 行自动分割新文件（编号递增）
//   - 线程安全
//
// 文件命名：{CMD}_{编号}.log  例如 DANMU_MSG_1.log, DANMU_MSG_2.log
// 每行格式：时间  原始 JSON
type RawMessageLogger struct {
	mu      sync.Mutex
	writers map[string]*cmdWriter // key: msg.Cmd (string 类型)
}

// cmdWriter 管理单个 cmd 类型的日志文件写入
type cmdWriter struct {
	cmd       string
	fileNum   int
	lineCount int
	file      *os.File
	date      string // 当前日期 "2006-01-02"
}

// NewRawMessageLogger 创建原始消息日志记录器。
func NewRawMessageLogger() *RawMessageLogger {
	return &RawMessageLogger{
		writers: make(map[string]*cmdWriter),
	}
}

// Log 记录一条原始消息。
//
// cmd 为消息类型（如 "DANMU_MSG", "SEND_GIFT"），raw 为原始 JSON 字节。
func (l *RawMessageLogger) Log(cmd string, raw []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()

	today := time.Now().Format("2006-01-02")

	w, ok := l.writers[cmd]
	if !ok {
		w = &cmdWriter{cmd: cmd, date: today}
		l.writers[cmd] = w
	}

	// 日期变了 → 关闭旧文件，重置编号从 1 开始
	if w.date != today {
		w.close()
		w.date = today
		w.fileNum = 0
		w.lineCount = 0
	}

	// 需要打开/切换文件
	if w.file == nil || w.lineCount >= rawLogMaxLines {
		if w.file != nil {
			w.close()
		}
		w.fileNum++
		w.lineCount = 0

		dir := filepath.Join(rawLogDir, today)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return // 目录创建失败，静默丢弃本条日志
		}

		filename := fmt.Sprintf("%s_%d.log", cmd, w.fileNum)
		path := filepath.Join(dir, filename)

		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return // 文件打开失败，静默丢弃
		}
		w.file = f
	}

	// 写入：时间  原始内容
	line := fmt.Sprintf("%s   %s\n", time.Now().Format("2006-01-02 15:04:05.000"), string(raw))
	w.file.WriteString(line)
	w.lineCount++
}

// Close 关闭所有打开的文件。
func (l *RawMessageLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, w := range l.writers {
		w.close()
	}
	l.writers = make(map[string]*cmdWriter)
}

// close 关闭单个 cmdWriter 的文件（不加锁，由调用方负责）。
func (w *cmdWriter) close() {
	if w.file != nil {
		w.file.Close()
		w.file = nil
	}
}
