package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DatabaseMysqlConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type DatabasePostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type DatabaseSqliteConfig struct {
	FilePath string `yaml:"filepath"`
}

type DatabaseConfig struct {
	Driver   string                 `yaml:"driver"`
	Mysql    DatabaseMysqlConfig    `yaml:"mysql"`
	Postgres DatabasePostgresConfig `yaml:"postgres"`
	Sqlite   DatabaseSqliteConfig   `yaml:"sqlite"`
}

type RedisConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}
type JWTConfig struct {
	Secret     string `yaml:"secret"`
	AccessTTL  int    `yaml:"access_ttl"`
	RefreshTTL int    `yaml:"refresh_ttl"`
}

type AltchaConfig struct {
	HMACKey string `yaml:"hmac_key"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type ServerConfig struct {
	Port         int `yaml:"port"`
	ReadTimeout  int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`
	IdleTimeout  int `yaml:"idle_timeout"`
}

type DatabasePoolConfig struct {
	MaxOpenConns    int `yaml:"max_open_conns"`
	MaxIdleConns    int `yaml:"max_idle_conns"`
	ConnMaxLifetime int `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime int `yaml:"conn_max_idle_time"`
}

type LiveConfig struct {
	StateFile string  `yaml:"state_file"`
	TestUIDs  []int64 `yaml:"test_uids"` // 测试机器人 UID 白名单，命中则仅记录日志不真正发送弹幕（可为空）
}

type Config struct {
	Server   ServerConfig       `yaml:"server"`
	Database DatabaseConfig     `yaml:"database"`
	Pool     DatabasePoolConfig `yaml:"pool"`
	Redis    RedisConfig        `yaml:"redis"`
	JWT      JWTConfig          `yaml:"jwt"`
	CORS     CORSConfig         `yaml:"cors"`
	Altcha   AltchaConfig       `yaml:"altcha"`
	Live     LiveConfig         `yaml:"live"`
}

// LoadConfig 解析 YAML
func LoadConfig(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	if err := yaml.Unmarshal(raw, c); err != nil {
		return nil, err
	}
	// 验证配置
	if err := ValidateConfig(c); err != nil {
		return nil, err
	}
	return c, nil
}

// ValidateConfig 验证配置的有效性，并为未设置的字段提供默认值
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("配置不能为空")
	}
	// 服务端口默认值
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 9000
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 15
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 15
	}
	if cfg.Server.IdleTimeout == 0 {
		cfg.Server.IdleTimeout = 60
	}
	// 数据库连接池默认值
	if cfg.Pool.MaxOpenConns == 0 {
		cfg.Pool.MaxOpenConns = 25
	}
	if cfg.Pool.MaxIdleConns == 0 {
		cfg.Pool.MaxIdleConns = 10
	}
	if cfg.Pool.ConnMaxLifetime == 0 {
		cfg.Pool.ConnMaxLifetime = 3600
	}
	if cfg.Pool.ConnMaxIdleTime == 0 {
		cfg.Pool.ConnMaxIdleTime = 1800
	}
	switch cfg.Database.Driver {
	case "mysql":
		if cfg.Database.Mysql.Host == "" {
			return fmt.Errorf("mysql 配置错误：host 不能为空")
		}
		if cfg.Database.Mysql.Port == 0 {
			return fmt.Errorf("mysql 配置错误：port 不能为 0")
		}
		if cfg.Database.Mysql.User == "" {
			return fmt.Errorf("mysql 配置错误：user 不能为空")
		}
		if cfg.Database.Mysql.Password == "" {
			return fmt.Errorf("mysql 配置错误：password 不能为空")
		}
		if cfg.Database.Mysql.DBName == "" {
			return fmt.Errorf("mysql 配置错误：dbname 不能为空")
		}
	case "postgres":
		if cfg.Database.Postgres.Host == "" {
			return fmt.Errorf("postgres 配置错误：host 不能为空")
		}
		if cfg.Database.Postgres.Port == 0 {
			return fmt.Errorf("postgres 配置错误：port 不能为 0")
		}
		if cfg.Database.Postgres.User == "" {
			return fmt.Errorf("postgres 配置错误：user 不能为空")
		}
		if cfg.Database.Postgres.Password == "" {
			return fmt.Errorf("postgres 配置错误：password 不能为空")
		}
		if cfg.Database.Postgres.DBName == "" {
			return fmt.Errorf("postgres 配置错误：dbname 不能为空")
		}
	case "sqlite":
		if cfg.Database.Sqlite.FilePath == "" {
			cfg.Database.Sqlite.FilePath = "data.db"
		}
	default:
		return fmt.Errorf("不支持的数据库驱动程序: %s", cfg.Database.Driver)
	}
	if cfg.Live.StateFile == "" {
		cfg.Live.StateFile = "bilibili_state.json"
	}
	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("JWT 密钥长度不能低于 32 位")
	}
	if cfg.JWT.AccessTTL <= 0 {
		return fmt.Errorf("access ttl 必须大于 0")
	}
	if cfg.JWT.RefreshTTL <= 0 {
		return fmt.Errorf("refresh ttl 必须大于 0")
	}
	return nil
}
