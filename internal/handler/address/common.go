package address

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/address"
)

// toAddressItemResp 将 service 层返回的收货地址项转换为对外响应结构
func toAddressItemResp(item address.AddressItem) resp.AddressItem {
	return resp.AddressItem{
		ID:         item.ID,
		Name:       item.Name,
		Phone:      item.Phone,
		RegionCode: item.RegionCode,
		Region:     item.Region,
		Detail:     item.Detail,
		Email:      item.Email,
		Type:       item.Type,
		IsDefault:  item.IsDefault,
	}
}

// intVal 安全解引用 int 指针（用于日志字段，nil 记 0）
func intVal(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
