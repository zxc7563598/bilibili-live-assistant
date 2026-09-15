package resp

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

// OrderGetConfirmResp 获取用户下单数据请求返回
type OrderGetConfirmResp struct {
	// id
	ID int64 `json:"id" example:"1"`
	// 到期时间(毫秒级时间戳)
	ExpireAt int64 `json:"expire_at" example:"1788417485000"`
	// 产品信息
	Product ProductItem `json:"product"`
}

// ProductItem 单个产品信息
type ProductItem struct {
	// 产品ID
	ID int64 `json:"id" example:"2"`
	// 产品名称
	Name string `json:"name" example:"蓝牙耳机"`
	// 产品封面
	Cover string `json:"cover" example:"https://cdn.hejunjie.life/avatars/oneadmin.jpeg"`
	// 产品价格
	Price int64 `json:"price" example:"30"`
	// 产品价格类型
	CreditType int `json:"credit_type" example:"1" enums:"0,1"`
	// 产品类型
	ProductType int `json:"product_type" example:"1" enums:"0,1"`
	// 产品SKU
	Sku string `json:"sku" example:"[{'aa':'bb'},{'aa':'bb'}]"`
	// 购买数量
	Count int64 `json:"count" example:"0"`
}

// OrderListPageByUserResp 请求返回
type OrderListPageByUserResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []OrderListPageByUserItem `json:"pageData"`
}

type OrderListPageByUserItem struct {
	// id
	ID int64 `json:"id" example:"1"`
	// 订单号
	OrderSn string `json:"order_sn" example:"P00029182821723123"`
	// 商品ID
	ProductID int64 `json:"product_id" example:"1"`
	// 商品名称
	ProductName string `json:"product_name" example:"商品名称"`
	// 商品封面图
	ProductCover string `json:"product_cover" example:"https://cdn.hejunjie.life/avatars/oneadmin.jpeg"`
	// 规格快照
	ProductSpecProperties string `json:"product_spec_properties" example:"[{'aa':'bb'},{'aa':'bb'}]"`
	// 购买数量
	Quantity int64 `json:"quantity" example:"1"`
	// 支付类型
	CreditType enum.CreditType `json:"credit_type" example:"1" enums:"0,1"`
	// 支付价格
	Price int64 `json:"price" example:"100"`
	// 订单状态
	OrderStatus enum.OrderStatus `json:"order_status" example:"1" enums:"0,1,2,3,4,5"`
	// 支付状态
	PayStatus enum.PayStatus `json:"pay_status" example:"1" enums:"0,1,2"`
	// 发货状态
	ShipStatus enum.ShipStatus `json:"ship_status" example:"1" enums:"0,1,2"`
	// 快递公司
	ExpressCompany string `json:"express_company" example:"xxxxxxx"`
	// 快递单号
	ExpressNo string `json:"express_no" example:"P000002919238123"`
	// 支付时间
	PayAt string `json:"pay_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 发货时间
	ProcessedAt string `json:"processed_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 取消时间
	CancelAt string `json:"cancel_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 用户备注
	Remark string `json:"remark" example:"xxxxxxxxxxxxx"`
}

// OrderListPageResp 后台分页查询订单返回
type OrderListPageResp struct {
	// 总计条数
	Total int64 `json:"total" example:"100"`
	// 当前页码数据
	PageData []OrderListPageItem `json:"pageData"`
}

type OrderListPageItem struct {
	// id
	ID int64 `json:"id" example:"1"`
	// 用户内部ID
	UserID int64 `json:"user_id" example:"1"`
	// 用户UID
	UID int64 `json:"uid" example:"54272611"`
	// 用户昵称
	Uname string `json:"uname" example:"哎呀又胖啦"`
	// 订单号
	OrderSn string `json:"order_sn" example:"P00029182821723123"`
	// 商品ID
	ProductID int64 `json:"product_id" example:"1"`
	// 商品名称
	ProductName string `json:"product_name" example:"商品名称"`
	// 商品封面图
	ProductCover string `json:"product_cover" example:"https://cdn.hejunjie.life/avatars/oneadmin.jpeg"`
	// 规格快照
	ProductSpecProperties string `json:"product_spec_properties" example:"[{'aa':'bb'},{'aa':'bb'}]"`
	// 购买数量
	Quantity int64 `json:"quantity" example:"1"`
	// 支付类型
	CreditType enum.CreditType `json:"credit_type" example:"1" enums:"0,1"`
	// 支付价格
	Price int64 `json:"price" example:"100"`
	// 收货人地址类型，0 虚拟 / 1 实体
	ReceiverType enum.AddressType `json:"receiver_type" example:"1" enums:"0,1"`
	// 收货人姓名（实体订单）
	ReceiverName string `json:"receiver_name" example:"张三"`
	// 收货人手机号（实体订单）
	ReceiverPhone string `json:"receiver_phone" example:"18888888888"`
	// 收货人地区文字描述（实体订单）
	ReceiverRegion string `json:"receiver_region" example:"山东省 济南市 历下区"`
	// 收货人详细地址（实体订单）
	ReceiverDetail string `json:"receiver_detail" example:"xx路xx号"`
	// 收货人邮箱地址（虚拟订单）
	ReceiverEmail string `json:"receiver_email" example:"x@x.com"`
	// 订单状态
	OrderStatus enum.OrderStatus `json:"order_status" example:"1" enums:"0,1,2,3,4,5"`
	// 支付状态
	PayStatus enum.PayStatus `json:"pay_status" example:"1" enums:"0,1,2"`
	// 发货状态
	ShipStatus enum.ShipStatus `json:"ship_status" example:"1" enums:"0,1,2"`
	// 快递公司
	ExpressCompany string `json:"express_company" example:"xxxxxxx"`
	// 快递单号
	ExpressNo string `json:"express_no" example:"P000002919238123"`
	// 支付时间
	PayAt string `json:"pay_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 发货时间
	ProcessedAt string `json:"processed_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 取消时间
	CancelAt string `json:"cancel_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 用户备注
	Remark string `json:"remark" example:"xxxxxxxxxxxxx"`
}

// OrderDetailsResp 后台获取订单详情返回，包含订单全部字段与用户 uid/uname
type OrderDetailsResp struct {
	// id
	ID int64 `json:"id" example:"1"`
	// 用户内部ID
	UserID int64 `json:"user_id" example:"1"`
	// 用户UID
	UID int64 `json:"uid" example:"54272611"`
	// 用户昵称
	Uname string `json:"uname" example:"哎呀又胖啦"`
	// 订单号
	OrderSn string `json:"order_sn" example:"P00029182821723123"`
	// 商品ID
	ProductID int64 `json:"product_id" example:"1"`
	// 商品SKU ID
	ProductSkuID int64 `json:"product_sku_id" example:"1"`
	// 商品名称
	ProductName string `json:"product_name" example:"商品名称"`
	// 商品封面图
	ProductCover string `json:"product_cover" example:"https://cdn.hejunjie.life/avatars/oneadmin.jpeg"`
	// 规格快照
	ProductSpecProperties string `json:"product_spec_properties" example:"[{'aa':'bb'},{'aa':'bb'}]"`
	// 购买数量
	Quantity int64 `json:"quantity" example:"1"`
	// 支付类型
	CreditType enum.CreditType `json:"credit_type" example:"1" enums:"0,1"`
	// 支付价格
	Price int64 `json:"price" example:"100"`
	// 收货人姓名
	ReceiverName string `json:"receiver_name" example:"张三"`
	// 收货人手机号（实体订单）
	ReceiverPhone string `json:"receiver_phone" example:"18888888888"`
	// 收货人地区code（实体订单），JSON 数组字符串
	ReceiverRegionCode string `json:"receiver_region_code" example:"['370000', '370100', '370116']"`
	// 收货人地区文字描述（实体订单）
	ReceiverRegion string `json:"receiver_region" example:"山东省 济南市 历下区"`
	// 收货人详细地址（实体订单）
	ReceiverDetail string `json:"receiver_detail" example:"xx路xx号"`
	// 收货人邮箱地址（虚拟订单）
	ReceiverEmail string `json:"receiver_email" example:"x@x.com"`
	// 收货人地址类型，0 虚拟 / 1 实体；决定发货与收货信息可变更的字段
	ReceiverType enum.AddressType `json:"receiver_type" example:"1" enums:"0,1"`
	// 订单状态
	OrderStatus enum.OrderStatus `json:"order_status" example:"1" enums:"0,1,2,3,4,5"`
	// 支付状态
	PayStatus enum.PayStatus `json:"pay_status" example:"1" enums:"0,1,2"`
	// 发货状态
	ShipStatus enum.ShipStatus `json:"ship_status" example:"1" enums:"0,1,2"`
	// 快递公司
	ExpressCompany string `json:"express_company" example:"xxxxxxx"`
	// 快递单号
	ExpressNo string `json:"express_no" example:"P000002919238123"`
	// 支付时间
	PayAt string `json:"pay_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 发货时间
	ProcessedAt string `json:"processed_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 取消时间
	CancelAt string `json:"cancel_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"xxxx-xx-xx xx:xx:xx"`
	// 用户备注
	Remark string `json:"remark" example:"xxxxxxxxxxxxx"`
}
