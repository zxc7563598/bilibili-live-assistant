package liveuser

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeParamInvalid                  = 10801 // 请求参数不合法（参数校验失败）
	CodeTokenInvalid                  = 20801 // 登录状态异常，请重新登录
	CodeTokenTypeInvalid              = 20802 // 登录状态异常，请重新登录
	CodeTokenExpired                  = 20803 // 登录已过期，请重新登录
	CodeLoginFailed                   = 40801 // 账号或密码错误（密码错误（避免提示过于具体））
	CodeAccountDisabled               = 40802 // 当前账号不可用（用户账号已停用）
	CodeInsufficientBalance           = 40803 // 余额不足（用户余额不足以完成本次兑换）
	CodeTokenizerInitFailed           = 50801 // 分词器初始化失败（分词器初始化失败）
	CodeUserNotFound                  = 50802 // 未知用户（未能查询到用户信息）
	CodeQueryFailed                   = 60801 // 系统繁忙，请稍后重试（数据库查询异常）
	CodeAccessTokenGenFailed          = 60802 // 系统繁忙，请稍后重试（accessToken 生成失败）
	CodeRefreshTokenGenFailed         = 60803 // 系统繁忙，请稍后重试（refreshToken 生成失败）
	CodeTokenPersistFailed            = 60804 // 系统繁忙，请稍后重试（更新管理员 token 失败）
	CodeAccessTokenCacheFailed        = 60805 // 系统繁忙，请稍后重试（redis 存储 accessToken 失败）
	CodeRefreshTokenCacheFailed       = 60806 // 系统繁忙，请稍后重试（redis 存储 refreshToken 失败）
	CodeTokenClearFailed              = 60807 // 系统繁忙，请稍后重试（redis 清空 token 失败）
	CodeBilibiliUserFailed            = 60808 // 获取B站用户信息失败，请稍后重试
)
