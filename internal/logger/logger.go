package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 可以定义多个 *zap.Logger 类型的 logger
var (
	AdminLogger       *zap.Logger
	RoleLogger        *zap.Logger
	MenuLogger        *zap.Logger
	AltchaLogger      *zap.Logger
	LiveLogger        *zap.Logger
	RobotConfigLogger *zap.Logger
	LiveDanmuLogger   *zap.Logger
	LiveGiftLogger    *zap.Logger
	LiveUserLogger    *zap.Logger
)

// InitAll 初始化所有模块 logger
func InitAll() {
	AdminLogger = InitLogger("admin", zapcore.InfoLevel)
	RoleLogger = InitLogger("role", zapcore.InfoLevel)
	MenuLogger = InitLogger("menu", zapcore.InfoLevel)
	AltchaLogger = InitLogger("altcha", zapcore.InfoLevel)
	LiveLogger = InitLogger("live", zapcore.InfoLevel)
	RobotConfigLogger = InitLogger("robotconfig", zapcore.InfoLevel)
	LiveDanmuLogger = InitLogger("livedanmu", zapcore.InfoLevel)
	LiveGiftLogger = InitLogger("livegift", zapcore.InfoLevel)
	LiveUserLogger = InitLogger("liveuser", zapcore.InfoLevel)
}

// InitLogger 初始化指定模块的 logger
func InitLogger(module string, level zapcore.Level) *zap.Logger {
	// 确保日志目录存在
	logDir := filepath.Join("logs", module)
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("无法创建日志目录: %v", err)
	}
	// 按天分割日志文件
	filename := filepath.Join(logDir, fmt.Sprintf("%s_%s.log", time.Now().Format(time.DateOnly), module))
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100, // MB
		MaxBackups: 30,  // 保留最近30个日志文件
		MaxAge:     7,   // 天
		Compress:   true,
	}
	// zap encoder 配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.LevelKey = "level"
	encoderConfig.CallerKey = "caller"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // JSON格式
		zapcore.AddSync(lumberjackLogger),
		level,
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger
}
