package liveuser

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// UserBalance 用户余额响应结构
type UserBalance struct {
	Points int64 // 积分
	Stars  int64 // 星光
}

// AddCreditLogParams 添加积分/星光
type AddCreditLogParams struct {
	UserID       int64             // 用户ID
	ChangeType   enum.ChangeType   // 变动类型（增加/减少）
	ChangeAmount int64             // 变动数值
	BizType      string            // 业务类型
	Remark       string            // 备注
	OperatorType enum.OperatorType // 操作方
	OperatorID   int64             // 操作人标识ID
}
