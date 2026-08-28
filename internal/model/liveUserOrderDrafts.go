package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// 用户订单草稿/预下单表
type LiveUserOrderDraft struct {
	ID           int64            `gorm:"primaryKey"`
	UserID       int64            `gorm:"not null;comment:用户ID"`
	ProductID    int64            `gorm:"not null;comment:商品ID"`
	ProductSkuID int64            `gorm:"not null;comment:商品SKU ID"`
	Quantity     int64            `gorm:"not null;comment:购买数量"`
	Status       enum.DraftStatus `gorm:"type:smallint;not null;default:0;comment:状态"`
	ExpireAt     int64            `gorm:"comment:到期时间"`
	Remark       string           `gorm:"type:varchar(255);comment:用户备注"`
	BaseModel
}

func (LiveUserOrderDraft) TableName() string {
	return "live_user_order_drafts"
}
