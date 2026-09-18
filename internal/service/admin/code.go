package admin

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeParamInvalid            = 10101 // 请求参数不合法（参数校验失败）
	CodePasswordRequired        = 10103 // 请输入密码（密码不能为空）
	CodePasswordTooShort        = 10104 // 密码长度不能少于6位（密码长度 < 6）
	CodePasswordTooLong         = 10105 // 密码长度不能超过32位（密码长度 > 32）
	CodeRoleRequired            = 10108 // 请至少为该账号分配一个角色（管理员最少需要绑定1个角色）
	CodeTokenInvalid            = 20101 // 登录状态异常，请重新登录
	CodeTokenTypeInvalid        = 20102 // 登录状态异常，请重新登录
	CodeTokenExpired            = 20103 // 登录已过期，请重新登录
	CodeRoleSwitchForbidden     = 30101 // 无法切换到该角色（当前账号没有切换到该角色的权限）
	CodeLoginFailed             = 40101 // 账号或密码错误（密码错误（避免提示过于具体））
	CodeAccountDisabled         = 40102 // 当前账号不可用（管理员账号已停用）
	CodeRoleDisabled            = 40103 // 当前角色不可用（角色已停用）
	CodePasswordSameAsOld       = 40104 // 新密码不能与旧密码相同（新旧密码一致）
	CodeAdminNotFound           = 50101 // 管理员不存在或已被删除（未能查询到管理员信息）
	CodeRoleNotFound            = 50102 // 角色不存在或已被删除（未能查询到角色信息）
	CodeQueryFailed             = 60101 // 系统繁忙，请稍后重试（数据库查询异常）
	CodeAccessTokenGenFailed    = 60102 // 系统繁忙，请稍后重试（accessToken 生成失败）
	CodeRefreshTokenGenFailed   = 60103 // 系统繁忙，请稍后重试（refreshToken 生成失败）
	CodeTokenPersistFailed      = 60104 // 系统繁忙，请稍后重试（更新管理员 token 失败）
	CodeAccessTokenCacheFailed  = 60105 // 系统繁忙，请稍后重试（redis 存储 accessToken 失败）
	CodeRefreshTokenCacheFailed = 60106 // 系统繁忙，请稍后重试（redis 存储 refreshToken 失败）
	CodeTokenClearFailed        = 60107 // 系统繁忙，请稍后重试（redis 清空 token 失败）
	CodeRoleChangeFailed        = 60108 // 操作失败，请稍后重试（变更管理员角色失败）
	CodePasswordHashFailed      = 60109 // 系统繁忙，请稍后重试（密码哈希生成失败）
	CodePasswordUpdateFailed    = 60110 // 操作失败，请稍后重试（修改管理员密码失败）
	CodeAdminSaveFailed         = 60111 // 操作失败，请稍后重试（管理员基本信息变更失败）
	CodeAdminDeleteFailed       = 60112 // 操作失败，请稍后重试（管理员删除失败）
	CodeProfileUpdateFailed     = 60113 // 操作失败，请稍后重试（管理员个人信息变更失败）
	CodeRoleUnbindFailed        = 60114 // 操作失败，请稍后重试（管理员绑定角色删除失败）
	CodeRoleBindFailed          = 60115 // 操作失败，请稍后重试（管理员绑定角色添加失败）
)
