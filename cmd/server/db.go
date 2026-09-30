package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dump"
	"gorm.io/gorm"
)

// dbUsage 子命令帮助
const dbUsage = `用法:
  BiliLiveAssistant db export [-config config.yaml] [-out dump.sql[.gz]] [-force]
  BiliLiveAssistant db import -config config.yaml -in dump.sql[.gz]

db export  把当前配置指向的数据库导出成 SQL 文件（可用 sqlite3 / mysql / psql 直接导入）
db import  把 SQL 文件导入到当前配置指向的空库，支持跨数据库（如 sqlite 导出后导入 mysql）
`

// runDBCommand 处理 db 子命令，返回 true 表示命令已被子命令接管
func runDBCommand(args []string) bool {
	if len(args) == 0 || args[0] != "db" {
		return false
	}
	if len(args) < 2 {
		fmt.Fprint(os.Stderr, dbUsage)
		os.Exit(2)
	}
	switch args[1] {
	case "export":
		runDBExport(args[2:])
	case "import":
		runDBImport(args[2:])
	case "-h", "--help", "help":
		fmt.Print(dbUsage)
	default:
		fmt.Fprintf(os.Stderr, "未知的 db 子命令: %s\n\n%s", args[1], dbUsage)
		os.Exit(2)
	}
	return true
}

// runDBExport 导出当前数据库
func runDBExport(args []string) {
	flags := flag.NewFlagSet("db export", flag.ExitOnError)
	configPath := flags.String("config", "config.yaml", "配置文件路径")
	out := flags.String("out", "", "导出文件路径，以 .gz 结尾时写出 gzip 压缩；默认为配置文件同目录下 dump-<数据库类型>-<时间>.sql")
	force := flags.Bool("force", false, "输出文件已存在时覆盖")
	flags.Parse(args)

	db, cfg := openDB(*configPath)
	defer closeDB(db)

	path := *out
	if path == "" {
		name := fmt.Sprintf("dump-%s-%s.sql", db.Dialector.Name(), time.Now().Format("20060102-150405"))
		path = filepath.Join(cfg.ConfigDir, name)
	}
	result, err := dump.Export(db, dump.ExportOptions{Path: path, Force: *force, Logf: log.Printf})
	if err != nil {
		log.Fatalf("导出失败: %v", err)
	}
	log.Printf("导出完成：%s（%d 张表 %d 行，%s，%.1f MB，耗时 %s）",
		result.Path, result.Tables, result.Rows, db.Dialector.Name(),
		float64(result.Size)/(1<<20), result.Duration.Round(time.Millisecond))
}

// runDBImport 导入到当前配置指向的空库
func runDBImport(args []string) {
	flags := flag.NewFlagSet("db import", flag.ExitOnError)
	configPath := flags.String("config", "config.yaml", "配置文件路径")
	in := flags.String("in", "", "导入文件路径（必填）")
	flags.Parse(args)
	if *in == "" {
		fmt.Fprintf(os.Stderr, "-in 参数必填\n\n%s", dbUsage)
		os.Exit(2)
	}

	db, _ := openDB(*configPath)
	defer closeDB(db)

	result, err := dump.Import(db, dump.ImportOptions{Path: *in, Logf: log.Printf})
	if err != nil {
		log.Fatalf("导入失败: %v", err)
	}
	log.Printf("导入完成：%s → %s（%d 张表 %d 行，耗时 %s）",
		result.Path, db.Dialector.Name(), result.Tables, result.Rows,
		result.Duration.Round(time.Millisecond))
}

// openDB 加载配置并连接数据库。与正常启动一致：配置文件缺失时自动生成，库不存在时自动建。
func openDB(configPath string) (*gorm.DB, *config.Config) {
	if _, err := config.EnsureConfigFile(configPath); err != nil {
		log.Fatalf("初始化配置文件失败: %v", err)
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("无法初始化数据库: %v", err)
	}
	return db, cfg
}

// closeDB 关闭底层连接
func closeDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	}
}
