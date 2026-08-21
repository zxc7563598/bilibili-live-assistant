package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
	serverDB, err := gorm.Open(mysql.Open(serverDSN), &gorm.Config{})
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
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	return db, nil
}

// 初始化 SQLite
func initSQLite(cfg *Config) (*gorm.DB, error) {
	s := cfg.Database.Sqlite
	filePath := s.FilePath
	// 如果是相对路径，则相对于二进制文件所在目录解析
	if !filepath.IsAbs(filePath) {
		execPath, err := os.Executable()
		if err == nil {
			filePath = filepath.Join(filepath.Dir(execPath), filePath)
		}
	}
	// 确保数据库文件目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建 SQLite 数据库目录失败: %w", err)
	}
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{TranslateError: true})
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
	serverDB, err := gorm.Open(postgres.Open(serverDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	// 创建数据库（尝试创建，忽略"已存在"的错误）
	createSQL := fmt.Sprintf(
		"CREATE DATABASE %s",
		p.DBName,
	)
	if err := serverDB.Exec(createSQL).Error; err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("创建数据库失败: %w", err)
		}
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	return db, nil
}
