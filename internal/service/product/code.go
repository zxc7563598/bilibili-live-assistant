package product

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 注意：dto/input 里 err:"required=11001" 这类 struct tag 只能是字符串字面量，
// 无法引用常量，那里仍会写数字。本文件只覆盖 Go 表达式中的引用。

const (
	CodeSkuRequired           = 11012 // 请至少上架一个规格组合（SKU 数量为 0）
	CodeSkuTooMany            = 11013 // 规格组合数不能超过 200（SKU 数量超限）
	CodeSpecNameRequired      = 11015 // 规格名称不能为空（规格名称为空）
	CodeSpecNameTooLong       = 11016 // 规格名称不能超过 100 个字符（规格名称超长）
	CodeSpecNameDuplicate     = 11017 // 规格名称不能重复（规格名称重复）
	CodeSpecValueRequired     = 11018 // 规格值不能为空（规格下没有规格值 / 规格值为空）
	CodeSpecValueTooLong      = 11019 // 规格值不能超过 100 个字符（规格值超长）
	CodeSpecValueDuplicate    = 11020 // 同一规格下的规格值不能重复（规格值重复）
	CodeSkuSpecMismatch       = 11021 // 规格组合的规格信息与规格设置不匹配（SKU 规格快照与规格定义不一致）
	CodeSkuPriceInvalid       = 11022 // 规格组合的价格必须为不小于 0 的整数（SKU 价格非法）
	CodeSkuStockInvalid       = 11023 // 规格组合的库存必须为不小于 0 的整数（SKU 库存非法）
	CodeImagePathRequired     = 11024 // 商品图片地址不能为空（图片路径为空）
	CodeImagePathTooLong      = 11025 // 商品图片地址不能超过 1000 个字符（图片路径超长）
	CodeImageTypeInvalid      = 11026 // 商品图片类型不合法（图片类型非 0/1）
	CodeSkuCostInvalid        = 11028 // 规格组合的成本价必须为不小于 0 的整数（SKU 成本价非法）
	CodeSkuSpecTooLong        = 11029 // 规格组合的规格信息过长（SKU 规格快照序列化后超出列宽）
	CodeSkuDuplicate          = 11030 // 规格组合重复（同一商品下出现相同的规格组合）
	CodeMinMemberLevelInvalid = 11031 // 请选择限购类型（限购档位非 0/1/2/3）
	CodeNotFound              = 51001 // 商品不存在（商品信息不存在）
	CodeQueryFailed           = 61001 // 系统繁忙，请稍后重试（数据库查询异常）
	CodeSaveFailed            = 61002 // 操作失败，请稍后重试（商品保存失败（商品模块））
)
