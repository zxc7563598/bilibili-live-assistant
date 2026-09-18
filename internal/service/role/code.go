package role

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeCodeRequired                  = 10202 // 请输入角色标识（角色 Code 不能为空）
	CodeNameRequired                  = 10203 // 请输入角色名称（角色名称不能为空）
	CodeEnableRequired                = 10204 // 请选择角色状态（角色状态不能为空）
	CodeMenuRequired                  = 10206 // 请至少为该角色分配一个菜单（角色最少需要绑定1个菜单）
	CodeRoleDisabled                  = 40202 // 绑定失败，当前角色不可用（当前角色已被停用）
	CodeOnlyRoleLeft                  = 40203 // 该管理员仅剩此角色，不可移除
	CodeNotFound                      = 50201 // 角色不存在或已被删除（未能查询到角色信息）
	CodeQueryFailed                   = 60201 // 系统繁忙，请稍后重试（数据库查询异常）
	CodeSaveFailed                    = 60202 // 操作失败，请稍后重试（角色信息修改失败）
	CodeDeleteFailed                  = 60203 // 操作失败，请稍后重试（角色删除失败）
	CodeMenuFetchFailed               = 60205 // 操作失败，请稍后重试（菜单信息获取失败）
	CodeBindAdminFailed               = 60206 // 操作失败，请稍后重试（角色绑定管理员失败）
	CodeTokenClearFailed              = 60207 // 系统繁忙，请稍后重试（角色变更后清理管理员 token 缓存失败）
	CodeTokenPersistFailed            = 60208 // 系统繁忙，请稍后重试（角色变更后更新管理员 token 记录失败）
)
