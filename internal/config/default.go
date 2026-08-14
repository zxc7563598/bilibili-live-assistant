package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

// defaultConfigYAML 首次运行且配置文件缺失时写入的默认配置
const defaultConfigYAML = `# BiliLive Assistant 默认配置
# 此文件在首次运行且未找到 config.yaml 时自动生成，可参考 config.example.yaml 修改。

server:
  port: 25443
  read_timeout: 15
  write_timeout: 15
  idle_timeout: 60

database:
  driver: sqlite # 数据库驱动：mysql | postgres | sqlite
  # mysql: # mysql 数据库配置（不使用可不进行配置）
  #   host: 127.0.0.1
  #   port: 3306
  #   user: user
  #   password: password
  #   dbname: databasename
  # postgres: # postgres 数据库配置（不使用可不进行配置）
  #   host: 127.0.0.1
  #   port: 5432
  #   user: user
  #   password: password
  #   dbname: databasename
  sqlite: # sqlite 数据库配置（不使用可不进行配置）
    filepath: data.db # 数据库文件路径，相对路径相对于二进制所在目录

pool:
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 3600
  conn_max_idle_time: 1800

# redis: # redis 配置（不使用可不进行配置）
#   host: 127.0.0.1
#   port: 6379
#   password:
#   db: 0
#   pool_size: 20
#   min_idle_conns: 5

jwt: # jwt 配置
  secret: "{{JWT_SECRET}}" # 自动生成的随机密钥
  access_ttl: 7200
  refresh_ttl: 604800

cors: # CORS 跨域配置
  allowed_origins: [] # 允许的前端来源列表，为空时允许所有来源

altcha: # altcha 验证码配置（hmac_key 留空则关闭验证码）
  hmac_key: ""

live: # B站 直播监听配置
  state_file: "bilibili_state.json" # B站 Cookie 持久化文件路径
  test_uids: [] # 测试机器人 UID 白名单，命中的机器人只记录日志不真正发送弹幕（可为空）
`

// defaultConfigContent 返回填充了随机 JWT 密钥的默认配置内容
func defaultConfigContent() (string, error) {
	secret, err := randomJWTSecret()
	if err != nil {
		return "", err
	}
	return strings.Replace(defaultConfigYAML, "{{JWT_SECRET}}", secret, 1), nil
}

// randomJWTSecret 使用 crypto/rand 生成 32 字节随机密钥并十六进制编码（64 字符）
func randomJWTSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// EnsureConfigFile 若 path 指向的配置文件不存在则写入默认配置
func EnsureConfigFile(path string) (created bool, err error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil // 已存在，不覆盖
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return false, err
		}
	}
	content, err := defaultConfigContent()
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return false, err
	}
	return true, nil
}
