package robotconfig

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeQueryFailed                   = 60501 // 获取配置失败，请稍后重试（数据库查询异常（机器人配置模块））
	CodeUpdateFailed                  = 60502 // 更新配置失败，请稍后重试（数据库更新异常（机器人配置模块））
	CodeCacheReloadFailed             = 60503 // 配置缓存刷新失败，请稍后重试（缓存更新异常（机器人配置模块））
)
