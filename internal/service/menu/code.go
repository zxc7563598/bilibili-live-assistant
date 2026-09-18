package menu

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeParamInvalid                  = 10301 // 请求参数不合法（参数校验失败）
	CodeNotFound                      = 50301 // 菜单不存在或已被删除（未能查询到菜单信息）
	CodeQueryFailed                   = 60301 // 系统繁忙，请稍后重试（数据库查询异常）
	CodeCreateFailed                  = 60302 // 操作失败，请稍后重试（菜单信息添加失败）
	CodeUpdateFailed                  = 60303 // 操作失败，请稍后重试（菜单信息变更失败）
	CodeToggleEnableFailed            = 60304 // 操作失败，请稍后重试（菜单状态切换失败）
	CodeDeleteFailed                  = 60305 // 操作失败，请稍后重试（菜单删除失败）
)
