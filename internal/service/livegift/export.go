package livegift

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// 礼物列表与盲盒列表的导出数据源。
//
// 两者是两个不同的接口（/livegift/list 与 /livegift/blindbox），列与筛选条件都不一样，
// 因此注册成两个模块名；Service 只有一个，用两个薄包装类型各自挂上 Source 方法。

// giftColumns 礼物列表允许导出的列，顺序与页面列一致
var giftColumns = []export.ColumnSpec[model.LiveGift]{
	{Key: "uid", TitleKey: "export.column.uid", Value: func(r model.LiveGift) any { return r.UID }},
	// 枚举实现了 Text(lang)，由 export 包渲染成当前语言的文案
	{Key: "badge_name", TitleKey: "export.column.badge_name", Value: func(r model.LiveGift) any { return r.BadgeName }},
	{Key: "uname", TitleKey: "export.column.uname", Value: func(r model.LiveGift) any { return r.Uname }},
	{Key: "gift_name", TitleKey: "export.column.gift_name", Value: func(r model.LiveGift) any { return r.GiftName }},
	{Key: "price", TitleKey: "export.column.price", Value: func(r model.LiveGift) any { return export.Money(r.Price) }},
	{Key: "num", TitleKey: "export.column.num", Value: func(r model.LiveGift) any { return r.Num }},
	// 页面上没有这个字段，是按 (price/100)*num 算出来的展示值，导出侧同样算一遍
	{Key: "total", TitleKey: "export.column.total", Value: func(r model.LiveGift) any { return export.Money(r.Price * r.Num) }},
	{Key: "badge_level", TitleKey: "export.column.badge_level", Value: func(r model.LiveGift) any { return r.BadgeLevel }},
	{Key: "badge_type", TitleKey: "export.column.badge_type", Value: func(r model.LiveGift) any { return r.BadgeType }},
	{Key: "send_at", TitleKey: "export.column.send_at", Value: func(r model.LiveGift) any { return export.UnixTime(r.SendAt) }},
}

// blindBoxColumns 盲盒列表允许导出的列，顺序与页面列一致
var blindBoxColumns = []export.ColumnSpec[model.LiveGift]{
	{Key: "uid", TitleKey: "export.column.uid", Value: func(r model.LiveGift) any { return r.UID }},
	{Key: "badge_name", TitleKey: "export.column.badge_name", Value: func(r model.LiveGift) any { return r.BadgeName }},
	{Key: "uname", TitleKey: "export.column.uname", Value: func(r model.LiveGift) any { return r.Uname }},
	{Key: "gift_name", TitleKey: "export.column.gift_name", Value: func(r model.LiveGift) any { return r.GiftName }},
	{Key: "price", TitleKey: "export.column.price", Value: func(r model.LiveGift) any { return export.Money(r.Price) }},
	{Key: "num", TitleKey: "export.column.num", Value: func(r model.LiveGift) any { return r.Num }},
	{Key: "total", TitleKey: "export.column.total", Value: func(r model.LiveGift) any { return export.Money(r.Price * r.Num) }},
	{Key: "original_gift_name", TitleKey: "export.column.original_gift_name", Value: func(r model.LiveGift) any { return r.OriginalGiftName }},
	{Key: "original_gift_price", TitleKey: "export.column.original_gift_price", Value: func(r model.LiveGift) any { return export.Money(r.OriginalGiftPrice) }},
	// 页面按 ((price - original_gift_price) * num) / 100 展示，可为负
	{Key: "profit", TitleKey: "export.column.profit", Value: func(r model.LiveGift) any { return export.Money((r.Price - r.OriginalGiftPrice) * r.Num) }},
	{Key: "send_at", TitleKey: "export.column.send_at", Value: func(r model.LiveGift) any { return export.UnixTime(r.SendAt) }},
}

var (
	giftValue     = export.ValueMapOf(giftColumns)
	blindBoxValue = export.ValueMapOf(blindBoxColumns)
)

// giftFilterInput 礼物列表的前端筛选条件，字段与列表接口同口径
type giftFilterInput struct {
	RoomID   *int64   `json:"room_id"`
	UID      *int64   `json:"uid"`
	Uname    *string  `json:"uname"`
	GiftName *string  `json:"gift_name"`
	GiftType *int     `json:"gift_type"`
	Original *int     `json:"original"`
	SendAt   *[]int64 `json:"send_at"`
}

// giftFilterQuery 规范化后的礼物筛选条件：时间区间已换算成秒级闭区间
type giftFilterQuery struct {
	RoomID      *int64  `json:"room_id"`
	UID         *int64  `json:"uid"`
	Uname       *string `json:"uname"`
	GiftName    *string `json:"gift_name"`
	GiftType    *int    `json:"gift_type"`
	Original    *int    `json:"original"`
	SendAtStart *int64  `json:"send_at_start"`
	SendAtEnd   *int64  `json:"send_at_end"`
}

// blindBoxFilterInput 盲盒列表的前端筛选条件
type blindBoxFilterInput struct {
	RoomID           *int64   `json:"room_id"`
	UID              *int64   `json:"uid"`
	Uname            *string  `json:"uname"`
	GiftName         *string  `json:"gift_name"`
	OriginalGiftName *string  `json:"original_gift_name"`
	SendAt           *[]int64 `json:"send_at"`
}

// blindBoxFilterQuery 规范化后的盲盒筛选条件
type blindBoxFilterQuery struct {
	RoomID           *int64  `json:"room_id"`
	UID              *int64  `json:"uid"`
	Uname            *string `json:"uname"`
	GiftName         *string `json:"gift_name"`
	OriginalGiftName *string `json:"original_gift_name"`
	SendAtStart      *int64  `json:"send_at_start"`
	SendAtEnd        *int64  `json:"send_at_end"`
}

// giftExporter 礼物列表的导出数据源
type giftExporter struct{ svc *Service }

// blindBoxExporter 盲盒列表的导出数据源
type blindBoxExporter struct{ svc *Service }

// NewGiftExporter 构造礼物列表的导出数据源
func NewGiftExporter(svc *Service) export.Source { return giftExporter{svc: svc} }

// NewBlindBoxExporter 构造盲盒列表的导出数据源
func NewBlindBoxExporter(svc *Service) export.Source { return blindBoxExporter{svc: svc} }

// ---------- 礼物列表 ----------

// Module 导出模块标识
func (e giftExporter) Module() string { return "livegift" }

// Columns 本模块允许导出的列
func (e giftExporter) Columns() []export.Column { return export.ColumnsOf(giftColumns) }

// Normalize 解析并校验前端筛选条件
func (e giftExporter) Normalize(filters json.RawMessage) (json.RawMessage, int, error) {
	var in giftFilterInput
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &in); err != nil {
			return nil, CodeParamInvalid, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	// 仓储的构建器对非法枚举是**静默忽略**的，不在这里拦下来就会导出一份筛选条件没生效的全量数据
	if in.GiftType != nil && !enum.GiftType(*in.GiftType).IsValid() {
		return nil, CodeParamInvalid, fmt.Errorf("gift_type 内容非法: %d", *in.GiftType)
	}
	if in.Original != nil && !enum.YesNo(*in.Original).IsValid() {
		return nil, CodeParamInvalid, fmt.Errorf("original 内容非法: %d", *in.Original)
	}
	q := giftFilterQuery{
		RoomID:   in.RoomID,
		UID:      in.UID,
		Uname:    in.Uname,
		GiftName: in.GiftName,
		GiftType: in.GiftType,
		Original: in.Original,
	}
	if in.SendAt != nil {
		q.SendAtStart, q.SendAtEnd = timeutil.SecondRange(*in.SendAt)
	}
	out, err := json.Marshal(q)
	if err != nil {
		return nil, CodeParamInvalid, fmt.Errorf("序列化导出筛选条件失败: %w", err)
	}
	return out, 0, nil
}

// Count 统计命中行数，limit > 0 时提前停止
func (e giftExporter) Count(ctx context.Context, filters json.RawMessage, limit int) (int64, error) {
	q, err := parseGiftQuery(filters)
	if err != nil {
		return 0, err
	}
	return e.svc.liveGiftRepo.CountFiltered(ctx, nil, q, limit)
}

// FetchChunk 取一块数据，按主键倒序
func (e giftExporter) FetchChunk(ctx context.Context, filters json.RawMessage, afterID int64, keys []string, limit int) ([][]any, int64, error) {
	q, err := parseGiftQuery(filters)
	if err != nil {
		return nil, 0, err
	}
	rows, err := e.svc.liveGiftRepo.ExportChunk(ctx, nil, q, afterID, limit)
	if err != nil {
		return nil, 0, err
	}
	return buildRecords(rows, keys, giftValue)
}

// ---------- 盲盒列表 ----------

// Module 导出模块标识
func (e blindBoxExporter) Module() string { return "livegiftblindbox" }

// Columns 本模块允许导出的列
func (e blindBoxExporter) Columns() []export.Column { return export.ColumnsOf(blindBoxColumns) }

// Normalize 解析并校验前端筛选条件
func (e blindBoxExporter) Normalize(filters json.RawMessage) (json.RawMessage, int, error) {
	var in blindBoxFilterInput
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &in); err != nil {
			return nil, CodeParamInvalid, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	q := blindBoxFilterQuery{
		RoomID:           in.RoomID,
		UID:              in.UID,
		Uname:            in.Uname,
		GiftName:         in.GiftName,
		OriginalGiftName: in.OriginalGiftName,
	}
	if in.SendAt != nil {
		q.SendAtStart, q.SendAtEnd = timeutil.SecondRange(*in.SendAt)
	}
	out, err := json.Marshal(q)
	if err != nil {
		return nil, CodeParamInvalid, fmt.Errorf("序列化导出筛选条件失败: %w", err)
	}
	return out, 0, nil
}

// Count 统计命中行数，limit > 0 时提前停止
func (e blindBoxExporter) Count(ctx context.Context, filters json.RawMessage, limit int) (int64, error) {
	q, err := parseBlindBoxQuery(filters)
	if err != nil {
		return 0, err
	}
	return e.svc.liveGiftRepo.BlindBoxCountFiltered(ctx, nil, q, limit)
}

// FetchChunk 取一块数据，按主键倒序
func (e blindBoxExporter) FetchChunk(ctx context.Context, filters json.RawMessage, afterID int64, keys []string, limit int) ([][]any, int64, error) {
	q, err := parseBlindBoxQuery(filters)
	if err != nil {
		return nil, 0, err
	}
	rows, err := e.svc.liveGiftRepo.BlindBoxExportChunk(ctx, nil, q, afterID, limit)
	if err != nil {
		return nil, 0, err
	}
	return buildRecords(rows, keys, blindBoxValue)
}

// ---------- 共用 ----------

// buildRecords 按调用方给的列序取值并组装成一块导出数据，游标取本块最后一行的主键
func buildRecords(rows []model.LiveGift, keys []string, values map[string]func(model.LiveGift) any) ([][]any, int64, error) {
	out := make([][]any, 0, len(rows))
	for _, row := range rows {
		rec := make([]any, 0, len(keys))
		for _, key := range keys {
			// keys 已由 export 包按 Columns() 校验过，这里的兜底只为防两处声明分叉
			value, ok := values[key]
			if !ok {
				return nil, 0, fmt.Errorf("未支持的导出列: %q", key)
			}
			rec = append(rec, value(row))
		}
		out = append(out, rec)
	}
	var nextID int64
	if len(rows) > 0 {
		nextID = rows[len(rows)-1].ID
	}
	return out, nextID, nil
}

// parseGiftQuery 把规范化的筛选条件还原为礼物列表的仓储查询结构
func parseGiftQuery(filters json.RawMessage) (model.LiveGiftListPageQuery, error) {
	var q giftFilterQuery
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &q); err != nil {
			return model.LiveGiftListPageQuery{}, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	return model.LiveGiftListPageQuery{
		RoomID:      q.RoomID,
		UID:         q.UID,
		Uname:       q.Uname,
		GiftName:    q.GiftName,
		GiftType:    q.GiftType,
		Original:    q.Original,
		SendAtStart: q.SendAtStart,
		SendAtEnd:   q.SendAtEnd,
	}, nil
}

// parseBlindBoxQuery 把规范化的筛选条件还原为盲盒列表的仓储查询结构
func parseBlindBoxQuery(filters json.RawMessage) (model.LiveGiftBlindBoxListPageQuery, error) {
	var q blindBoxFilterQuery
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &q); err != nil {
			return model.LiveGiftBlindBoxListPageQuery{}, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	return model.LiveGiftBlindBoxListPageQuery{
		RoomID:           q.RoomID,
		UID:              q.UID,
		Uname:            q.Uname,
		GiftName:         q.GiftName,
		OriginalGiftName: q.OriginalGiftName,
		SendAtStart:      q.SendAtStart,
		SendAtEnd:        q.SendAtEnd,
	}, nil
}
