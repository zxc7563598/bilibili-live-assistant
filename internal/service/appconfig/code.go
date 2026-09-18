package appconfig

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeImageInvalid                  = 10901 // 图片信息不合法（图片 MIME 类型获取失败）
	CodeSaveFailed                    = 60902 // 保存配置失败，请稍后重试（数据库更新异常（App 配置模块））
	CodeCacheReloadFailed             = 60903 // 配置缓存刷新失败，请稍后重试（缓存更新异常（App 配置模块））
	CodeOSSInitFailed                 = 60905 // 阿里云OSS初始化失败（阿里云OSS初始化失败）
	CodeOSSConfigFailed               = 60906 // 阿里云OSS配置或连接异常（阿里云OSS配置或连接异常）
)
