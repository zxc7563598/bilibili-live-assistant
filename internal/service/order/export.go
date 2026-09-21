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

// 订单列表的导出数据源。
//
// 管理端的「订单列表」与「发货」两个页面打的是同一个 /order/list（发货页只是把
// order_status / ship_status 预设成了待发货），因此共用一个模块名，白名单取两个页面列的并集，
// 各页只会发自己显示的那几列。

// productInfo 拼出「商品信息」列的文案。
//
// 对齐 web/src/views/order/list/index.vue 与 order/delivery/index.vue 的 renderProduct：
// 该单元格里摊了商品名、数量、规格（还有封面图，导出带不上）。改前端展示时这里要一起改。
func productInfo(row model.LiveUserOrderListItem) string {
	specs := formatSpecProperties(row.ProductSpecProperties)
	if specs == "" {
		// 与前端 `规格：${specs || '无'}` 一致
		specs = "无"
	}
	return fmt.Sprintf("%s ×%d 规格：%s", row.ProductName, row.Quantity, specs)
}

// receiverInfo 拼出「收货信息」列的文案。
//
// 对齐 web/src/views/order/delivery/index.vue 的 renderReceiver：
// 实体订单要姓名、电话、完整地址，虚拟订单只要邮箱。前端用两行排版，导出拼成一行。
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

// emptyCell 前端 renderReceiver 用 dash() 把空值显示成破折号，导出保留同样观感：
// 整格都没内容时给一个破折号，而不是留一格空白让人分不清「没填」和「没这一列」。
const emptyCell = "—"

// formatSpecProperties 把规格快照渲染成 `颜色:红 尺码:L`，多条规格用 ` / ` 分隔。
//
// 复刻 web/src/utils/common.js 的 formatProductSpecs。落库形状是**按规格定义顺序**序列化的
// [{"规格名":"规格值"},…]，且每个元素恰好一对（见 internal/service/product/common.go 的
// marshalSpecProperties），所以按元素顺序遍历即可与页面同序；元素内排序只是为了
// 万一出现多对时结果稳定。解析失败按前端的做法返回空串（外层会显示成「无」）。
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

// exportColumns 允许导出的列，是两个页面列的并集。
//
// 状态列都是真枚举（实现了 Text(lang)），由 export 包渲染成当前语言的文案。
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

// columnValue 列 key → 取值函数，与上面的声明同源
var columnValue = export.ValueMapOf(exportColumns)

// exportFilterInput 前端传来的筛选条件，字段与列表接口同口径。
//
// 状态字段用指针：裸 int 分不清「没传」与「传 0」，而 0 是合法的筛选值
// （未支付、未发货）—— 发货页的默认筛选就是 order_status=1 & ship_status=0。
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

// Normalize 解析并校验前端筛选条件。
//
// 三个状态字段在这里显式校验：仓储的构建器对非法枚举是静默忽略的，不拦下来就会
// 导出一份筛选条件没生效的全量订单。
func (s *Service) Normalize(filters json.RawMessage) (json.RawMessage, int, error) {
	var f exportFilterInput
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &f); err != nil {
			return nil, CodeParamInvalid, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
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
	out := make([][]any, 0, len(rows))
	for _, row := range rows {
		rec := make([]any, 0, len(keys))
		for _, key := range keys {
			// keys 已由 export 包按 Columns() 校验过，这里的兜底只为防两处声明分叉
			value, ok := columnValue[key]
			if !ok {
				return nil, 0, fmt.Errorf("未支持的导出列: %q", key)
			}
			rec = append(rec, value(row))
		}
		out = append(out, rec)
	}
	// 游标取本块最后一行的主键；空块时由调用方按「不足一块」结束循环，不会再用到
	var nextID int64
	if len(rows) > 0 {
		nextID = rows[len(rows)-1].ID
	}
	return out, nextID, nil
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
