package i18n

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

var (
	errorsMap = map[string]map[int]string{}
	keysMap   = map[string]map[string]string{}
)

type ctxKey string

const LangKey ctxKey = "lang"

// InitLocales 初始化语言文件
func InitLocales() error {
	// 生产环境：使用 embed
	if gin.Mode() == gin.ReleaseMode {
		if err := loadLocalesFromEmbed(); err != nil {
			return fmt.Errorf("加载 embed 语言文件失败: %w", err)
		}
		return nil
	}
	// 开发环境：使用本地文件（方便热更新）
	if err := loadLocales("internal/i18n/locales"); err != nil {
		return fmt.Errorf("加载本地语言文件失败: %w", err)
	}
	return nil
}

// loadLocales 从指定目录扫描并加载所有语言文件
func loadLocales(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取语言目录失败: %w", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		lang := strings.TrimSuffix(name, filepath.Ext(name))
		path := filepath.Join(dir, name)
		if err := load(lang, path); err != nil {
			return err
		}
	}
	if len(errorsMap) == 0 && len(keysMap) == 0 {
		return fmt.Errorf("没有加载到任何语言文件")
	}
	return nil
}

// loadLocalesFromEmbed 从 embed.FS 中加载所有内嵌的语言文件
func loadLocalesFromEmbed() error {
	files, err := fs.ReadDir(localeFS, "locales")
	if err != nil {
		return fmt.Errorf("读取 embed 目录失败: %w", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		lang := strings.TrimSuffix(name, filepath.Ext(name))
		data, err := fs.ReadFile(localeFS, filepath.Join("locales", name))
		if err != nil {
			return err
		}
		if err := loadFromBytes(lang, data); err != nil {
			return err
		}
	}
	if len(errorsMap) == 0 && len(keysMap) == 0 {
		return fmt.Errorf("没有加载到任何语言文件")
	}
	return nil
}

// Load 从指定文件路径加载单个语言文件
func load(lang string, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取语言文件失败 %s: %w", lang, err)
	}
	return loadFromBytes(lang, data)
}

// loadFromBytes 从字节数据中解析并加载语言内容
func loadFromBytes(lang string, data []byte) error {
	raw := map[string]any{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("解析语言文件失败 %s: %w", lang, err)
	}
	errMap := map[int]string{}
	keyMap := map[string]string{}
	for k, v := range raw {
		if k == "error" {
			if m, ok := v.(map[any]any); ok {
				for codeKey, msg := range m {
					var code int
					switch c := codeKey.(type) {
					case int:
						code = c
					case float64:
						code = int(c)
					case string:
						if cInt, err := strconv.Atoi(c); err == nil {
							code = cInt
						} else {
							continue
						}
					default:
						continue
					}
					errMap[code] = fmt.Sprint(msg)
				}
			}
			continue
		}
		flatten(keyMap, k, v)
	}
	errorsMap[lang] = errMap
	keysMap[lang] = keyMap
	return nil
}

// E 获取指定 Code 对应语言内容
func E(lang string, code int) string {
	if langMap, ok := errorsMap[lang]; ok {
		if msg, ok := langMap[code]; ok {
			return msg
		}
	}
	// fallback zh
	if zhMap, ok := errorsMap["zh"]; ok {
		if msg, ok := zhMap[code]; ok {
			return msg
		}
	}
	return "unknown error"
}

// T 获取指定 Key 对应语言内容
func T(lang string, key string, args ...string) string {
	if len(args) > 0 {
		key = key + "." + strings.Join(args, ".")
	}
	if langMap, ok := keysMap[lang]; ok {
		if msg, ok := langMap[key]; ok {
			return msg
		}
	}
	// fallback zh
	if zhMap, ok := keysMap["zh"]; ok {
		if msg, ok := zhMap[key]; ok {
			return msg
		}
	}
	return key
}

// flatten 用于拼接 YAML 配置的 Key
func flatten(result map[string]string, prefix string, value any) {
	switch v := value.(type) {
	case string:
		result[prefix] = v
	case map[string]any:
		for k, val := range v {
			key := prefix + "." + k
			flatten(result, key, val)
		}
	}
}

// GetLang 获取当前上下文使用语言
func GetLang(ctx context.Context) string {
	lang, ok := ctx.Value(LangKey).(string)
	if !ok {
		return "zh"
	}
	return lang
}
