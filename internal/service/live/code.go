package live

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeNotLoggedIn                   = 40401 // 未登录B站账号，请先扫码登录（未登录 B站）
	CodeListenerRunning               = 40402 // 监听器已在运行中（监听器已在运行）
	CodeRoomIDInvalid                 = 40404 // 无效的房间号（房间号无效）
	CodeStartListenerFailed           = 40406 // 启动监听器失败（启动监听失败）
	CodeStopListenerFailed            = 40407 // 停止监听器失败（停止监听失败）
	CodeBilibiliAPIFailed             = 60401 // B站API请求失败，请稍后重试（B站 API 请求失败）
	CodeStateFileFailed               = 60402 // 本地状态文件读写失败，请稍后重试
)
