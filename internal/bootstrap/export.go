package bootstrap

import (
	"log"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livegift"
)

// initExportService 装配导出服务并注册各业务模块的导出数据源。
//
// 新增一个可导出的模块 = 在 internal/service/<模块>/export.go 里实现 export.Source，
// 然后把它追加到下面的数据源列表；前端对应页面给 MeCrud 加一个 export-module 属性即可。
// 详见 internal/service/export/CLAUDE.md。
func initExportService(cfg *config.Config, svc *Services) *export.Service {
	exportSvc := export.New(
		export.Config{
			Secret:           cfg.JWT.Secret,
			MaxRows:          cfg.Export.MaxRows,
			MaxConcurrent:    cfg.Export.MaxConcurrent,
			BatchSize:        cfg.Export.BatchSize,
			TicketTTL:        time.Duration(cfg.Export.TicketTTL) * time.Second,
			MaxDuration:      time.Duration(cfg.Export.MaxDuration) * time.Second,
			WriteIdleTimeout: time.Duration(cfg.Export.WriteIdleTimeout) * time.Second,
		},
		svc.LiveDanmu,
		svc.LiveUser,
		livegift.NewGiftExporter(svc.LiveGift),
		livegift.NewBlindBoxExporter(svc.LiveGift),
		svc.LivePk,
	)
	// 启动时列出已注册的模块
	log.Printf("已注册导出模块: %v（单次上限 %d 行，并发上限 %d）", exportSvc.Modules(), cfg.Export.MaxRows, cfg.Export.MaxConcurrent)
	return exportSvc
}
