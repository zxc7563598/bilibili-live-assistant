package logger

import (
	"io"
	"log"
	"os"
	"sync"
)

// 输出分流用到的子目录名，同时也是文件名中的模块名
const (
	ginLogName = "gin"    // Gin 访问日志与 panic 堆栈
	stdLogName = "stdlog" // 标准库 log（log.Printf / log.Println / log.Fatalf 等）
)

var (
	redirectOnce sync.Once
	ginSink      io.Writer
)

// InitRedirect 把标准库 log 与 Gin 的输出分别重定向到日志根目录下的独立子目录
//
// 必须在 InitAll 之后调用：两者都依赖 logDir，且 InitAll 自身用 log.Fatalf 报错，
// 那条报错必须落在 stderr（宝塔面板日志）——日志目录建不出来时，也没有文件可写。
// mirrorConsole 为 true（开发模式）时文件与 stderr 双写，本地终端照旧可见；
// 为 false（生产）时只写文件，不再往面板日志里灌内容。
func InitRedirect(mirrorConsole bool) {
	redirectOnce.Do(func() {
		// 先建 sink、最后再 SetOutput：构建过程中的告警仍走 stderr
		stdSink := newSink(stdLogName, mirrorConsole)
		ginSink = newSink(ginLogName, mirrorConsole)
		log.SetOutput(stdSink)
	})
}

// GinWriter 返回 Gin 访问日志与 panic 堆栈的写入目标
//
// 返回 io.Writer 而非 gin 相关类型，internal/logger 因此不需要依赖 gin 包。
// 未调用 InitRedirect，或日志目录不可用时返回 os.Stderr，Gin 输出照旧能被面板收集。
func GinWriter() io.Writer {
	if ginSink != nil {
		return ginSink
	}
	return os.Stderr
}

// newSink 构建写入目标；日志目录不可用时降级为 stderr，不让进程挂掉
func newSink(name string, mirrorConsole bool) io.Writer {
	w, err := newRotatingWriter(name, name)
	if err != nil {
		log.Printf("[日志] %s 日志目录不可用，输出回退到 stderr: %v", name, err)
		return os.Stderr
	}
	if mirrorConsole {
		// 文件在前：先写持久化目标再回显终端，避免终端阻塞拖住落盘
		return &lockWriter{w: io.MultiWriter(w, os.Stderr)}
	}
	return &lockWriter{w: w}
}
