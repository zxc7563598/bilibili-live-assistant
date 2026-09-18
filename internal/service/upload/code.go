package upload

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeNotImage          = 11402 // 图片信息不合法（上传内容不是有效图片（上传模块））
	CodeFormParseFailed   = 11401 // 请求参数不合法（上传表单解析失败）
	CodeFileRequired      = 11403 // 请上传图片文件（未接收到上传文件或文件为空（上传模块））
	CodeFileTooLarge      = 11404 // 图片大小超出限制（上传请求体或文件超过大小上限（上传模块））
	CodeSceneInvalid      = 11405 // 上传场景不合法（scene 不在上传白名单（上传模块））
	CodePathInvalid       = 11407 // 图片路径不合法（同步 OSS 的路径不是 /uploads 下的合法访问路径（上传模块））
	CodeOSSNotConfigured  = 41401 // OSS 尚未配置，请先保存阿里云 OSS 配置（同步前未检测到完整的 OSS 四项配置（上传模块））
	CodeLocalFileNotFound = 51401 // 本地图片不存在或已被清理，请重新上传（待同步到 OSS 的本地文件不存在（上传模块））
	CodeSaveFailed        = 61401 // 图片保存失败，请稍后重试（图片落盘异常（上传模块））
	CodeReadFailed        = 61402 // 图片读取失败，请稍后重试（读取本地图片状态异常（上传模块））
	CodeOSSInitFailed     = 61403 // 阿里云OSS初始化失败（阿里云OSS初始化失败（上传模块））
	CodeOSSUploadFailed   = 61404 // 图片同步到 OSS 失败，请稍后重试（图片上传 OSS 异常（上传模块））
)
