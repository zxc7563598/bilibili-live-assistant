package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// 用户订单表
type LiveUserOrder struct {
	ID                    int64            `gorm:"primaryKey"`
	UserID                int64            `gorm:"not null;comment:用户ID"`
	OrderSn               string           `gorm:"type:varchar(50);not null;comment:订单号"`
	ProductID             int64            `gorm:"not null;comment:商品ID"`
	ProductSkuID          int64            `gorm:"not null;comment:商品SKU ID"`
	ProductName           string           `gorm:"type:varchar(255);not null;comment:商品名称"`
	ProductCover          string           `gorm:"type:varchar(255);not null;comment:商品封面图"`
	ProductSpecProperties string           `gorm:"type:varchar(1000);not null;comment:规格快照，[{key_name:value_name}]"`
	Quantity              int64            `gorm:"not null;comment:购买数量"`
	CreditType            enum.CreditType  `gorm:"type:smallint;not null;default:0;comment:支付类型"`
	Price                 int64            `gorm:"not null;default:0;comment:支付价格"`
	ReceiverName          string           `gorm:"type:varchar(100);comment:收货人姓名"`
	ReceiverPhone         string           `gorm:"type:varchar(100);comment:收货人手机号"`
	ReceiverRegionCode    string           `gorm:"type:varchar(100);comment:收货人地区code,['370000', '370100', '370116']"`
	ReceiverRegion        string           `gorm:"type:varchar(255);comment:收货人地区文字描述，空格隔开"`
	ReceiverDetail        string           `gorm:"type:varchar(255);comment:收货人详细地址"`
	ReceiverEmail         string           `gorm:"type:varchar(255);comment:收货人邮箱地址"`
	ReceiverType          enum.AddressType `gorm:"type:smallint;not null;default:0;comment:收货人地址类型"`
	OrderStatus           enum.OrderStatus `gorm:"type:smallint;not null;default:0;comment:订单状态"`
	PayStatus             enum.PayStatus   `gorm:"type:smallint;not null;default:0;comment:支付状态"`
	ShipStatus            enum.ShipStatus  `gorm:"type:smallint;not null;default:0;comment:发货状态"`
	ExpressCompany        string           `gorm:"type:varchar(100);comment:快递公司"`
	ExpressNo             string           `gorm:"type:varchar(100);comment:快递单号"`
	PayAt                 int64            `gorm:"comment:支付时间"`
	ProcessedAt           int64            `gorm:"comment:发货时间"`
	CancelAt              int64            `gorm:"comment:取消时间"`
	Remark                string           `gorm:"type:varchar(255);comment:用户备注"`
	BaseModel
}

func (LiveUserOrder) TableName() string {
	return "live_user_orders"
}

// LiveUserOrderListPageQuery 用户订单分页查询入参，不对应数据库表
type LiveUserOrderListPageQuery struct {
	UserID      int64
	OrderStatus *int
	Offset      int
	Limit       int
}
