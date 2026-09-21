package livedanmu

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// exportColumns 允许导出的列。key 与前端表格列的 key 对应，TitleKey 指向 i18n 的 export.column.* 文案。
//
// 不含裸 id：内部主键，各模块含义还不一致，导出没有意义。
var exportColumns = []export.ColumnSpec[model.LiveDanmu]{
	{Key: "room_id", TitleKey: "export.column.room_id", Value: func(r model.LiveDanmu) any { return r.RoomID }},
	{Key: "live_id", TitleKey: "export.column.live_id", Value: func(r model.LiveDanmu) any { return r.LiveID }},
	{Key: "uid", TitleKey: "export.column.uid", Value: func(r model.LiveDanmu) any { return r.UID }},
	{Key: "uname", TitleKey: "export.column.uname", Value: func(r model.LiveDanmu) any { return r.Uname }},
	{Key: "msg", TitleKey: "export.column.msg", Value: func(r model.LiveDanmu) any { return r.Msg }},
	{Key: "badge_room_id", TitleKey: "export.column.badge_room_id", Value: func(r model.LiveDanmu) any { return r.BadgeRoomID }},
	{Key: "badge_name", TitleKey: "export.column.badge_name", Value: func(r model.LiveDanmu) any { return r.BadgeName }},
	{Key: "badge_level", TitleKey: "export.column.badge_level", Value: func(r model.LiveDanmu) any { return r.BadgeLevel }},
	// 枚举实现了 Text(lang)，由 export 包渲染成当前语言的文案
	{Key: "badge_type", TitleKey: "export.column.badge_type", Value: func(r model.LiveDanmu) any { return r.BadgeType }},
	{Key: "send_at", TitleKey: "export.column.send_at", Value: func(r model.LiveDanmu) any { return export.UnixTime(r.SendAt) }},
}

// columnValue 列 key → 取值函数，与上面的声明同源，不会漂移
var columnValue = export.ValueMapOf(exportColumns)

// exportFilterInput 前端传来的筛选条件，字段与列表接口同口径
type exportFilterInput struct {
	RoomID *int64   `json:"room_id"`
	UID    *int64   `json:"uid"`
	Uname  *string  `json:"uname"`
	Msg    *string  `json:"msg"`
	SendAt *[]int64 `json:"send_at"`
}

// exportFilterQuery 规范化后的筛选条件：时间区间已换算成秒级的闭区间
type exportFilterQuery struct {
	RoomID      *int64  `json:"room_id"`
	UID         *int64  `json:"uid"`
	Uname       *string `json:"uname"`
	Msg         *string `json:"msg"`
	SendAtStart *int64  `json:"send_at_start"`
	SendAtEnd   *int64  `json:"send_at_end"`
}

// Module 导出模块标识
func (s *Service) Module() string { return "livedanmu" }

// Columns 本模块允许导出的列
func (s *Service) Columns() []export.Column { return export.ColumnsOf(exportColumns) }

// Normalize 解析并校验前端筛选条件，把时间区间换算成秒级闭区间后回写为规范 JSON
func (s *Service) Normalize(filters json.RawMessage) (json.RawMessage, int, error) {
	var in exportFilterInput
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &in); err != nil {
			return nil, CodeParamInvalid, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	q := exportFilterQuery{
		RoomID: in.RoomID,
		UID:    in.UID,
		Uname:  in.Uname,
		Msg:    in.Msg,
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
func (s *Service) Count(ctx context.Context, filters json.RawMessage, limit int) (int64, error) {
	q, err := parseExportQuery(filters)
	if err != nil {
		return 0, err
	}
	return s.liveDanmuRepo.CountFiltered(ctx, nil, q, limit)
}

// FetchChunk 取一块数据，按主键倒序
func (s *Service) FetchChunk(ctx context.Context, filters json.RawMessage, afterID int64, keys []string, limit int) ([][]any, int64, error) {
	q, err := parseExportQuery(filters)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.liveDanmuRepo.ExportChunk(ctx, nil, q, afterID, limit)
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

// parseExportQuery 把规范化后的筛选条件还原为仓储查询结构
func parseExportQuery(filters json.RawMessage) (model.LiveDanmuListPageQuery, error) {
	var q exportFilterQuery
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &q); err != nil {
			return model.LiveDanmuListPageQuery{}, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	return model.LiveDanmuListPageQuery{
		RoomID:      q.RoomID,
		UID:         q.UID,
		Uname:       q.Uname,
		Msg:         q.Msg,
		SendAtStart: q.SendAtStart,
		SendAtEnd:   q.SendAtEnd,
	}, nil
}

// 取值逻辑已并入上面的 exportColumns 单一字面量：声明与取数同源，不再单独维护一份 switch。
