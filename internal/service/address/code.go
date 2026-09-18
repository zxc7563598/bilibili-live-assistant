package address

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeParamInvalid       = 11301 // 地址参数不合法（地址参数校验失败）
	CodeRegionInvalid      = 11302 // 地区信息不正确（region_code 不是合法的省市区链）
	CodeAddressTypeInvalid = 11303 // 地址类型不正确（地址类型非法）
	CodeNameRequired       = 11304 // 请填写收件人姓名（收件人姓名为空）
	CodePhoneRequired      = 11305 // 请填写手机号（实体地址手机号为空）
	CodeRegionRequired     = 11306 // 请选择所在地区（实体地址地区为空或无效）
	CodeDetailRequired     = 11307 // 请填写详细地址（实体地址详细地址为空）
	CodeEmailRequired      = 11308 // 请填写电子邮箱（虚拟地址电子邮箱为空）
	CodeNotFound           = 51301 // 收货地址不存在（地址不存在或不属于该用户）
	CodeQueryFailed        = 61301 // 系统繁忙，请稍后重试（数据库异常（地址模块））
)
