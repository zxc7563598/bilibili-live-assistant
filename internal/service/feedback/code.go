package feedback

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeTypeRequired                  = 11201 // 请选择问题类型（问题类型为空）
	CodeTypeTooLong                   = 11202 // 问题类型过长（问题类型超过最大长度限制）
	CodeContentRequired               = 11203 // 请填写投诉内容（投诉内容为空）
	CodeContentTooLong                = 11204 // 投诉内容过长（投诉内容超过最大长度限制）
	CodeContactRequired               = 11205 // 请填写联系方式（联系方式为空）
	CodeContactTooLong                = 11206 // 联系方式过长（联系方式超过最大长度限制）
	CodeNotFound                      = 51201 // 投诉记录不存在（未能查询到投诉记录（反馈模块-后台管理））
	CodeQueryFailed                   = 61201 // 系统繁忙，请稍后重试（数据库异常（反馈模块））
)
