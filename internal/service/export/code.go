package export

// 本模块错误码，与 internal/i18n/locales/zh/en.yaml 的 error 节点一一对应。
//
// 使用约定（见 internal/service/CLAUDE.md）：
//   - 每个模块只使用本模块 MM 段的错误码，不引用其它模块的
//   - 因此不同模块出现相同文案是可以接受的，不要为了消除重复而共用错误码
//
// 各业务模块自己的筛选条件校验错误用**本模块**的码（如弹幕模块的 10601），
// 本段只覆盖导出机制本身的失败。

const (
	CodeColumnInvalid  = 11701 // 导出列不合法（含模块未声明的列）
	CodeColumnEmpty    = 11702 // 请至少选择一列导出
	CodeFilterInvalid  = 11703 // 导出参数不合法
	CodeTicketInvalid  = 21701 // 导出链接已失效，请重新导出（下载凭证无效或已过期）
	CodeExportBusy     = 41701 // 同时导出的任务过多，请稍后再试（超过 export.max_concurrent）
	CodeRowLimit       = 41702 // 导出数据量超过上限，请缩小筛选范围（超过 export.max_rows）
	CodeRowEmpty       = 41703 // 没有可导出的数据（命中 0 行）
	CodeModuleNotFound = 51701 // 未能查询到导出模块（模块名不在导出注册表中）
	CodeExportFailed   = 61701 // 系统繁忙，请稍后重试（导出流程内部异常：计数/读取/签发失败）
)
