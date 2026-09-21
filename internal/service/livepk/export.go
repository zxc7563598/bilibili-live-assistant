package livepk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// PK 对战记录的导出数据源。

// exportColumns 允许导出的列，顺序与页面列一致。
//
// pk_status / battle_type / match_type 在库里是裸 int64（值来自 B 站协议，后端没有对应枚举），
// 页面上的中文标签来自前端本地的选项数组，所以这里直接出原始数值 —— 不在后端凭空造一份
// 与前端并存的文案，那会多出一个需要两边同步的维护点。
var exportColumns = []export.ColumnSpec[model.LivePkLog]{
	{Key: "pk_id", TitleKey: "export.column.pk_id", Value: func(r model.LivePkLog) any { return r.PkID }},
	{Key: "pk_status", TitleKey: "export.column.pk_status", Value: func(r model.LivePkLog) any { return r.PkStatus }},
	{Key: "battle_type", TitleKey: "export.column.battle_type", Value: func(r model.LivePkLog) any { return r.BattleType }},
	{Key: "rival_uid", TitleKey: "export.column.rival_uid", Value: func(r model.LivePkLog) any { return r.RivalUID }},
	{Key: "rival_uname", TitleKey: "export.column.rival_uname", Value: func(r model.LivePkLog) any { return r.RivalUname }},
	{Key: "self_votes", TitleKey: "export.column.self_votes", Value: func(r model.LivePkLog) any { return r.SelfVotes }},
	{Key: "rival_votes", TitleKey: "export.column.rival_votes", Value: func(r model.LivePkLog) any { return r.RivalVotes }},
	{Key: "start_at", TitleKey: "export.column.start_at", Value: func(r model.LivePkLog) any { return export.UnixTime(r.StartAt) }},
	{Key: "settle_at", TitleKey: "export.column.settle_at", Value: func(r model.LivePkLog) any { return export.UnixTime(r.SettleAt) }},
}

// columnValue 列 key → 取值函数，与上面的声明同源
var columnValue = export.ValueMapOf(exportColumns)

// exportFilterInput 前端传来的筛选条件，字段与列表接口同口径。
//
// 注意时间字段在前端叫 start_at（不是 send_at），值是毫秒级区间。
type exportFilterInput struct {
	RoomID     *int64   `json:"room_id"`
	RivalUID   *int64   `json:"rival_uid"`
	RivalUname *string  `json:"rival_uname"`
	Result     *int     `json:"result"`
	StartAt    *[]int64 `json:"start_at"`
}

// exportFilterQuery 规范化后的筛选条件：时间区间已换算成秒级闭区间
type exportFilterQuery struct {
	RoomID       *int64  `json:"room_id"`
	RivalUID     *int64  `json:"rival_uid"`
	RivalUname   *string `json:"rival_uname"`
	Result       *int    `json:"result"`
	StartAtStart *int64  `json:"start_at_start"`
	StartAtEnd   *int64  `json:"start_at_end"`
}

// Module 导出模块标识
func (s *Service) Module() string { return "livepk" }

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
	// 仓储的构建器对 result 不做校验就直接拼进 SQL，非法值在这里拦下来
	if in.Result != nil {
		switch *in.Result {
		case resultLose, resultWin:
		default:
			return nil, CodeParamInvalid, fmt.Errorf("result 内容非法: %d", *in.Result)
		}
	}
	q := exportFilterQuery{
		RoomID:     in.RoomID,
		RivalUID:   in.RivalUID,
		RivalUname: in.RivalUname,
		Result:     in.Result,
	}
	if in.StartAt != nil {
		q.StartAtStart, q.StartAtEnd = timeutil.SecondRange(*in.StartAt)
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
	return s.livePkLogRepo.CountFiltered(ctx, nil, q, limit)
}

// FetchChunk 取一块数据，按主键倒序
func (s *Service) FetchChunk(ctx context.Context, filters json.RawMessage, afterID int64, keys []string, limit int) ([][]any, int64, error) {
	q, err := parseExportQuery(filters)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.livePkLogRepo.ExportChunk(ctx, nil, q, afterID, limit)
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
func parseExportQuery(filters json.RawMessage) (model.LivePkLogListPageQuery, error) {
	var q exportFilterQuery
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &q); err != nil {
			return model.LivePkLogListPageQuery{}, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	return model.LivePkLogListPageQuery{
		RoomID:       q.RoomID,
		RivalUID:     q.RivalUID,
		RivalUname:   q.RivalUname,
		SelfResult:   q.Result,
		StartAtStart: q.StartAtStart,
		StartAtEnd:   q.StartAtEnd,
	}, nil
}
