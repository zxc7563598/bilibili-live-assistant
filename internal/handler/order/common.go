package order

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/order"
)

// toOrderAdminListPageItem 后台订单列表转换：在商城端字段基础上额外下发 user_id（内部ID）与 uid/uname
func toOrderAdminListPageItem(list []order.ListPageItem) []resp.OrderListPageItem {
	res := make([]resp.OrderListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.OrderListPageItem{
			ID:                    v.ID,
			UserID:                v.UserID,
			UID:                   v.UID,
			Uname:                 v.Uname,
			OrderSn:               v.OrderSn,
			ProductID:             v.ProductID,
			ProductName:           v.ProductName,
			ProductCover:          v.ProductCover,
			ProductSpecProperties: v.ProductSpecProperties,
			Quantity:              v.Quantity,
			CreditType:            v.CreditType,
			Price:                 v.Price,
			OrderStatus:           v.OrderStatus,
			PayStatus:             v.PayStatus,
			ShipStatus:            v.ShipStatus,
			ExpressCompany:        v.ExpressCompany,
			ExpressNo:             v.ExpressNo,
			PayAt:                 v.PayAt,
			ProcessedAt:           v.ProcessedAt,
			CancelAt:              v.CancelAt,
			CreatedAt:             v.CreatedAt,
			Remark:                v.Remark,
		})
	}
	return res
}

// toOrderDetailsResp 订单详情转换：订单全字段 + 用户 uid/uname
func toOrderDetailsResp(v order.DetailsItem) resp.OrderDetailsResp {
	return resp.OrderDetailsResp{
		ID:                    v.ID,
		UserID:                v.UserID,
		UID:                   v.UID,
		Uname:                 v.Uname,
		OrderSn:               v.OrderSn,
		ProductID:             v.ProductID,
		ProductSkuID:          v.ProductSkuID,
		ProductName:           v.ProductName,
		ProductCover:          v.ProductCover,
		ProductSpecProperties: v.ProductSpecProperties,
		Quantity:              v.Quantity,
		CreditType:            v.CreditType,
		Price:                 v.Price,
		ReceiverName:          v.ReceiverName,
		ReceiverPhone:         v.ReceiverPhone,
		ReceiverRegionCode:    v.ReceiverRegionCode,
		ReceiverRegion:        v.ReceiverRegion,
		ReceiverDetail:        v.ReceiverDetail,
		ReceiverEmail:         v.ReceiverEmail,
		ReceiverType:          v.ReceiverType,
		OrderStatus:           v.OrderStatus,
		PayStatus:             v.PayStatus,
		ShipStatus:            v.ShipStatus,
		ExpressCompany:        v.ExpressCompany,
		ExpressNo:             v.ExpressNo,
		PayAt:                 v.PayAt,
		ProcessedAt:           v.ProcessedAt,
		CancelAt:              v.CancelAt,
		CreatedAt:             v.CreatedAt,
		UpdatedAt:             v.UpdatedAt,
		Remark:                v.Remark,
	}
}

func toOrderListPageByUserItem(list []order.ListPageItem) []resp.OrderListPageByUserItem {
	res := make([]resp.OrderListPageByUserItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.OrderListPageByUserItem{
			ID:                    v.ID,
			OrderSn:               v.OrderSn,
			ProductID:             v.ProductID,
			ProductName:           v.ProductName,
			ProductCover:          v.ProductCover,
			ProductSpecProperties: v.ProductSpecProperties,
			Quantity:              v.Quantity,
			CreditType:            v.CreditType,
			Price:                 v.Price,
			OrderStatus:           v.OrderStatus,
			PayStatus:             v.PayStatus,
			ShipStatus:            v.ShipStatus,
			ExpressCompany:        v.ExpressCompany,
			ExpressNo:             v.ExpressNo,
			PayAt:                 v.PayAt,
			ProcessedAt:           v.ProcessedAt,
			CancelAt:              v.CancelAt,
			CreatedAt:             v.CreatedAt,
			Remark:                v.Remark,
		})
	}
	return res
}
