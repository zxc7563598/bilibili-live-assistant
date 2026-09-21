package logger

import (
	"log"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 可以定义多个 *zap.Logger 类型的 logger
var (
	AdminLogger       *zap.Logger
	AppConfigLogger   *zap.Logger
	RoleLogger        *zap.Logger
	MenuLogger        *zap.Logger
	AltchaLogger      *zap.Logger
	LiveLogger        *zap.Logger
	RobotConfigLogger *zap.Logger
	LiveDanmuLogger   *zap.Logger
	LiveGiftLogger    *zap.Logger
	LivePkLogger      *zap.Logger
	LiveUserLogger    *zap.Logger
	ProductLogger     *zap.Logger
	OrderLogger       *zap.Logger
	AddressLogger     *zap.Logger
	FeedbackLogger    *zap.Logger
	UploadLogger      *zap.Logger
	ExportLogger      *zap.Logger
)

// logDir 日志根目录，由 InitAll 按配置注入；未注入时使用工作目录下的 logs
var logDir = "logs"

var (
	initedMu  sync.Mutex
	initedDir string // 已初始化过的日志根目录，空表示尚未初始化
)

// InitAll 初始化所有模块 logger，dir 为日志根目录（由 cmd/server 传入配置中已解析的路径）
func InitAll(dir string) {
	if dir != "" {
		logDir = dir
	}
	initedMu.Lock()
	defer initedMu.Unlock()
	// 同一目录重复初始化直接复用：否则会产生两组 lumberjack 实例轮转同一批文件，
	// 互相 rename 对方正在写的文件，日志会莫名丢失
	if initedDir == logDir {
		return
	}
	AdminLogger = InitLogger("admin", zapcore.InfoLevel)
	AppConfigLogger = InitLogger("appconfig", zapcore.InfoLevel)
	RoleLogger = InitLogger("role", zapcore.InfoLevel)
	MenuLogger = InitLogger("menu", zapcore.InfoLevel)
	AltchaLogger = InitLogger("altcha", zapcore.InfoLevel)
	LiveLogger = InitLogger("live", zapcore.InfoLevel)
	RobotConfigLogger = InitLogger("robotconfig", zapcore.InfoLevel)
	LiveDanmuLogger = InitLogger("livedanmu", zapcore.InfoLevel)
	LiveGiftLogger = InitLogger("livegift", zapcore.InfoLevel)
	LivePkLogger = InitLogger("livepk", zapcore.InfoLevel)
	LiveUserLogger = InitLogger("liveuser", zapcore.InfoLevel)
	ProductLogger = InitLogger("product", zapcore.InfoLevel)
	OrderLogger = InitLogger("order", zapcore.InfoLevel)
	AddressLogger = InitLogger("address", zapcore.InfoLevel)
	FeedbackLogger = InitLogger("feedback", zapcore.InfoLevel)
	UploadLogger = InitLogger("upload", zapcore.InfoLevel)
	ExportLogger = InitLogger("export", zapcore.InfoLevel)
	initedDir = logDir
}

// InitLogger 初始化指定模块的 logger
func InitLogger(module string, level zapcore.Level) *zap.Logger {
	// 模块日志目录不可用视为致命：业务日志缺失时继续运行没有意义
	w, err := newRotatingWriter(module, module)
	if err != nil {
		log.Fatalf("无法创建日志目录: %v", err)
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
		zapcore.AddSync(w),
		level,
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger
}
