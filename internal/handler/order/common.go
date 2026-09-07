package order

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/order"
)

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
