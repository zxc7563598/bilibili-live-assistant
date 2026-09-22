package order

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
)

// exportColumns 允许导出的列，是订单列表与发货页两个页面列的并集
var exportColumns = []export.ColumnSpec[model.LiveUserOrderListItem]{
	{Key: "order_sn", TitleKey: "export.column.order_sn", Value: func(r model.LiveUserOrderListItem) any { return r.OrderSn }},
	{Key: "uname", TitleKey: "export.column.uname", Value: func(r model.LiveUserOrderListItem) any { return r.Uname }},
	{Key: "product_info", TitleKey: "export.column.product_info", Value: func(r model.LiveUserOrderListItem) any { return productInfo(r) }},
	{Key: "pay_status", TitleKey: "export.column.pay_status", Value: func(r model.LiveUserOrderListItem) any { return r.PayStatus }},
	{Key: "ship_status", TitleKey: "export.column.ship_status", Value: func(r model.LiveUserOrderListItem) any { return r.ShipStatus }},
	{Key: "order_status", TitleKey: "export.column.order_status", Value: func(r model.LiveUserOrderListItem) any { return r.OrderStatus }},
	{Key: "receiver_type", TitleKey: "export.column.receiver_type", Value: func(r model.LiveUserOrderListItem) any { return r.ReceiverType }},
	{Key: "receiver_info", TitleKey: "export.column.receiver_info", Value: func(r model.LiveUserOrderListItem) any { return receiverInfo(r) }},
	{Key: "created_at", TitleKey: "export.column.created_at", Value: func(r model.LiveUserOrderListItem) any { return export.UnixTime(r.CreatedAt) }},
}

var columnValue = export.ValueMapOf(exportColumns)

// exportFilterInput 前端传来的筛选条件，字段与列表接口同口径
type exportFilterInput struct {
	UID         *int64  `json:"uid"`
	Uname       *string `json:"uname"`
	OrderSn     *string `json:"order_sn"`
	OrderStatus *int    `json:"order_status"`
	PayStatus   *int    `json:"pay_status"`
	ShipStatus  *int    `json:"ship_status"`
}

// Module 导出模块标识
func (s *Service) Module() string { return "order" }

// Columns 本模块允许导出的列
func (s *Service) Columns() []export.Column { return export.ColumnsOf(exportColumns) }

// Normalize 解析并校验前端筛选条件
func (s *Service) Normalize(filters json.RawMessage) (json.RawMessage, int, error) {
	var f exportFilterInput
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &f); err != nil {
			return nil, CodeParamInvalid, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	// 仓储的构建器对非法枚举是**静默忽略**的，不在这里拦下来就会导出一份筛选条件没生效的全量数据
	if f.OrderStatus != nil && !enum.OrderStatus(*f.OrderStatus).IsValid() {
		return nil, CodeParamInvalid, fmt.Errorf("order_status 内容非法: %d", *f.OrderStatus)
	}
	if f.PayStatus != nil && !enum.PayStatus(*f.PayStatus).IsValid() {
		return nil, CodeParamInvalid, fmt.Errorf("pay_status 内容非法: %d", *f.PayStatus)
	}
	if f.ShipStatus != nil && !enum.ShipStatus(*f.ShipStatus).IsValid() {
		return nil, CodeParamInvalid, fmt.Errorf("ship_status 内容非法: %d", *f.ShipStatus)
	}
	out, err := json.Marshal(f)
	if err != nil {
		return nil, CodeParamInvalid, fmt.Errorf("序列化导出筛选条件失败: %w", err)
	}
	return out, 0, nil
}

// Count 统计命中行数，limit > 0 时提前停止
func (s *Service) Count(ctx context.Context, filters json.RawMessage, limit int) (int64, error) {
	q, err := parseExportQuery(filters)
	if err != nil {
		return 0, err
	}
	return s.liveUserOrderRepo.CountFiltered(ctx, nil, q, limit)
}

// FetchChunk 取一块数据，按主键倒序
func (s *Service) FetchChunk(ctx context.Context, filters json.RawMessage, afterID int64, keys []string, limit int) ([][]any, int64, error) {
	q, err := parseExportQuery(filters)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.liveUserOrderRepo.ExportChunk(ctx, nil, q, afterID, limit)
	if err != nil {
		return nil, 0, err
	}
	return export.BuildRecords(rows, keys, columnValue, func(r model.LiveUserOrderListItem) int64 { return r.ID })
}

// parseExportQuery 把规范化的筛选条件还原为仓储查询结构
func parseExportQuery(filters json.RawMessage) (model.LiveUserOrderListPageQuery, error) {
	var f exportFilterInput
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &f); err != nil {
			return model.LiveUserOrderListPageQuery{}, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	return model.LiveUserOrderListPageQuery{
		UID:         f.UID,
		Uname:       f.Uname,
		OrderSn:     f.OrderSn,
		OrderStatus: f.OrderStatus,
		PayStatus:   f.PayStatus,
		ShipStatus:  f.ShipStatus,
	}, nil
}

// emptyCell 前端的 dash() 把空值显示成破折号，导出保留同样观感
const emptyCell = "—"

// productInfo 拼出「商品信息」列的文案：`商品名 ×2 规格：颜色:红 / 尺码:L`
func productInfo(row model.LiveUserOrderListItem) string {
	specs := formatSpecProperties(row.ProductSpecProperties)
	if specs == "" {
		// 与前端 `规格：${specs || '无'}` 一致
		specs = "无"
	}
	return fmt.Sprintf("%s ×%d 规格：%s", row.ProductName, row.Quantity, specs)
}

// receiverInfo 拼出「收货信息」列的文案：实体单要姓名、电话、完整地址，虚拟单只要邮箱
func receiverInfo(row model.LiveUserOrderListItem) string {
	if row.ReceiverType == enum.AddressTypeActual {
		parts := make([]string, 0, 4)
		for _, v := range []string{row.ReceiverName, row.ReceiverPhone, row.ReceiverRegion, row.ReceiverDetail} {
			if v != "" {
				parts = append(parts, v)
			}
		}
		if len(parts) == 0 {
			return emptyCell
		}
		return strings.Join(parts, " ")
	}
	if row.ReceiverEmail == "" {
		return emptyCell
	}
	return row.ReceiverEmail
}

// formatSpecProperties 把规格快照渲染成 `颜色:红 尺码:L`，多条规格用 ` / ` 分隔
func formatSpecProperties(raw string) string {
	if raw == "" {
		return ""
	}
	var pairs []map[string]string
	if err := json.Unmarshal([]byte(raw), &pairs); err != nil {
		return ""
	}
	items := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		keys := make([]string, 0, len(pair))
		for k := range pair {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			// 与前端一致：空值不参与渲染
			if pair[k] == "" {
				continue
			}
			parts = append(parts, k+":"+pair[k])
		}
		if len(parts) > 0 {
			items = append(items, strings.Join(parts, " "))
		}
	}
	return strings.Join(items, " / ")
}
