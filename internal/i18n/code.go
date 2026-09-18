package i18n

// MM=00 通用模块的错误码，与 internal/i18n/locales/zh.yaml 的 error 节点一一对应。
//
// 这一段是唯一允许跨模块共用的：使用方是中间件、参数校验与 handler —
// 业务 service 不应引用，各模块有自己的 MM 段（见 internal/service/*/code.go）。
//
// 注意：dto/input 的 err:"..." struct tag 只能是字符串字面量，无法引用常量，
// 那里仍会写数字。
//
// 放在 i18n 包内是因为这批码就是 i18n 的键，旁边就是它镜像的 YAML。

const (
	CodeParamInvalid                  = 10001 // 请求参数不合法（参数校验失败（通用））
	CodeAccessTokenParseFailed        = 10002 // 登录状态异常，请重新登录（accessToken 解析失败）
	CodeAccessTokenInvalid            = 10003 // 登录状态异常，请重新登录（accessToken 数据异常）
	CodeAuthorizationMissing          = 10004 // 登录状态异常，请重新登录（未检测到 Authorization 请求头）
	CodeAuthorizationParseFailed      = 10005 // 登录状态异常，请重新登录（Authorization 解析失败）
	CodeCachedTokenParseFailed        = 10006 // 登录状态异常，请重新登录（redis accessToken 解析失败）
	CodeCachedTokenInvalid            = 10007 // 登录状态异常，请重新登录（redis accessToken 数据异常）
	CodeTokenRefreshed                = 10008 // 登录状态异常，请重新登录（redis accessToken 已刷新）
	CodeDecryptFailed                 = 10009 // 请求参数不合法（加密参数解密失败（通用））
	CodeTokenExpired                  = 20001 // 登录已过期，请重新登录（accessToken / session 失效）
	CodeRateLimited                   = 20002 // 请求过于频繁，请稍后再试（速率限制）
	CodeForbidden                     = 30001 // 无权限操作（当前账号没有访问权限）
	CodeSystemBusy                    = 60001 // 系统繁忙，请稍后重试（RSA 密钥初始化失败（系统-通用模块））
)
