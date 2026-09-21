package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// sqlStateDuplicateDatabase PostgreSQL「数据库已存在」的 SQLSTATE
const sqlStateDuplicateDatabase = "42P04"

// gormLogger 返回写入 <日志根目录>/gorm/ 的 GORM logger
//
// GORM 的默认 logger 自建 log.New(os.Stdout, "\r\n", log.LstdFlags) 输出，绕过 log.SetOutput，
// 宝塔部署时慢 SQL 会漏进面板日志，并且带着 Colorful 打开时的一串 ANSI 颜色码。
// 阈值与级别沿用 GORM 默认值：只记慢 SQL（≥200ms）与执行报错。
func gormLogger() gormlogger.Interface {
	return gormlogger.New(
		log.New(logger.GormWriter(), "", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      gormlogger.Warn,
			Colorful:      false, // 写文件不需要 ANSI 颜色码
		},
	)
}

// InitDB 根据配置初始化数据库，并返回 *gorm.DB
func InitDB(cfg *Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}
	var db *gorm.DB
	var err error
	switch cfg.Database.Driver {
	case "mysql":
		db, err = initMySQL(cfg)
	case "postgres":
		db, err = initPostgres(cfg)
	case "sqlite":
		db, err = initSQLite(cfg)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动程序: %s", cfg.Database.Driver)
	}
	if err != nil {
		return nil, err
	}
	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层 sql.DB 失败: %w", err)
	}
	if cfg.Database.Driver == "sqlite" {
		// SQLite 仅支持单写者，连接数固定为 1
		sqlDB.SetMaxOpenConns(1)
	} else {
		sqlDB.SetMaxOpenConns(cfg.Pool.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.Pool.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.Pool.ConnMaxLifetime) * time.Second)
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Pool.ConnMaxIdleTime) * time.Second)
	}
	return db, nil
}

// 初始化 MySQL
func initMySQL(cfg *Config) (*gorm.DB, error) {
	m := cfg.Database.Mysql
	if m.Port == 0 {
		m.Port = 3306
	}
	// 先连接 MySQL server（不指定数据库）
	serverDSN := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		m.User,
		m.Password,
		m.Host,
		m.Port,
	)
	serverDB, err := gorm.Open(mysql.Open(serverDSN), &gorm.Config{Logger: gormLogger()})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	// 创建数据库
	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
		m.DBName,
	)
	if err := serverDB.Exec(createSQL).Error; err != nil {
		return nil, fmt.Errorf("创建数据库失败: %w", err)
	}
	// 连接目标数据库
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.User,
		m.Password,
		m.Host,
		m.Port,
		m.DBName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true, Logger: gormLogger()})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	return db, nil
}

// 初始化 SQLite
func initSQLite(cfg *Config) (*gorm.DB, error) {
	// 路径已由 config.LoadConfig → resolveRuntimeDirs 按配置文件所在目录解析为绝对路径，
	// 此处不再二次推导（详见 internal/config/path.go）
	filePath := cfg.Database.Sqlite.FilePath
	// 确保数据库文件目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建 SQLite 数据库目录失败: %w", err)
	}
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{TranslateError: true, Logger: gormLogger()})
	if err != nil {
		return nil, fmt.Errorf("SQLite 数据库连接失败: %w", err)
	}
	return db, nil
}

// 初始化 PostgreSQL
func initPostgres(cfg *Config) (*gorm.DB, error) {
	p := cfg.Database.Postgres
	if p.Port == 0 {
		p.Port = 5432
	}
	// 连接 postgres 默认数据库
	serverDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%d sslmode=disable",
		p.Host,
		p.User,
		p.Password,
		p.Port,
	)
	serverDB, err := gorm.Open(postgres.Open(serverDSN), &gorm.Config{Logger: gormLogger()})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	// 创建数据库（尝试创建，忽略"已存在"的错误）
	createSQL := fmt.Sprintf(
		"CREATE DATABASE %s",
		p.DBName,
	)
	// 创建数据库（已存在则跳过，沿用现有库）
	if err := serverDB.Exec(createSQL).Error; err != nil && !isDatabaseExists(err) {
		return nil, fmt.Errorf("创建数据库失败: %w", err)
	}
	// 连接目标数据库
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		p.Host,
		p.User,
		p.Password,
		p.DBName,
		p.Port,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true, Logger: gormLogger()})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	return db, nil
}

// isDatabaseExists 判断错误是否为 PostgreSQL 的「数据库已存在」
//
// 按 SQLSTATE 42P04 判定而不是匹配错误文案：报错信息会随服务端 lc_messages
// 本地化，匹配 "already exists" 一旦失效，正常启动会被误判成「创建数据库失败」。
func isDatabaseExists(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == sqlStateDuplicateDatabase
}
