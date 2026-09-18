package order

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeParamInvalid                  = 11101 // 请求参数不合法（下单参数不合法）
	CodeAddressTypeInvalid            = 11102 // 地址类型不正确（地址类型非法）
	CodeReceiverNameRequired          = 11103 // 请填写收件人姓名（收件人姓名为空）
	CodeReceiverPhoneRequired         = 11104 // 请填写手机号（实体地址手机号为空）
	CodeRegionRequired                = 11105 // 请选择所在地区（实体地址地区为空或无效）
	CodeReceiverDetailRequired        = 11106 // 请填写详细地址（实体地址详细地址为空）
	CodeReceiverEmailRequired         = 11107 // 请填写电子邮箱（虚拟地址电子邮箱为空）
	CodeRegionInvalid                 = 11108 // 地区信息不正确（region_code 不是合法的省市区链）
	CodeInsufficientStock             = 41101 // 库存不足（库存不足，无法下单）
	CodeVirtualNoExpress              = 41102 // 虚拟订单不支持快递信息
	CodeVirtualEmailOnly              = 41103 // 虚拟订单只支持变更邮箱
	CodeActualAddressOnly             = 41104 // 实体订单只支持变更收货地址
	CodeInsufficientBalance           = 41105 // 余额不足
	CodeProductNotFound               = 51101 // 未能查询到商品（商品或SKU不存在）
	CodeNoPendingDraft                = 51102 // 无待支付订单（无待支付订单草稿）
	CodeDraftNotFound                 = 51103 // 待重新购买的订单不存在（历史草稿不存在或不属于当前用户）
	CodeDraftExpired                  = 51104 // 当前订单已过期（当前订单已过期）
	CodeOrderNotFound                 = 51105 // 未能查询到订单（订单不存在（订单模块-后台管理））
	CodeAddressNotFound               = 51106 // 收货地址不存在
	CodeQueryFailed                   = 61101 // 系统繁忙，请稍后重试（数据库异常（订单模块））
)
