package livepk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// exportColumns 允许导出的列
var exportColumns = []export.ColumnSpec[model.LivePkLog]{
	{Key: "pk_id", TitleKey: "export.column.pk_id", Value: func(r model.LivePkLog) any { return r.PkID }},
	{Key: "pk_status", TitleKey: "export.column.pk_status", Value: func(r model.LivePkLog) any { return enum.PkStatus(r.PkStatus) }},
	{Key: "battle_type", TitleKey: "export.column.battle_type", Value: func(r model.LivePkLog) any { return enum.PkBattleType(r.BattleType) }},
	{Key: "rival_uid", TitleKey: "export.column.rival_uid", Value: func(r model.LivePkLog) any { return r.RivalUID }},
	{Key: "rival_uname", TitleKey: "export.column.rival_uname", Value: func(r model.LivePkLog) any { return r.RivalUname }},
	{Key: "self_votes", TitleKey: "export.column.self_votes", Value: func(r model.LivePkLog) any { return r.SelfVotes }},
	{Key: "rival_votes", TitleKey: "export.column.rival_votes", Value: func(r model.LivePkLog) any { return r.RivalVotes }},
	{Key: "start_at", TitleKey: "export.column.start_at", Value: func(r model.LivePkLog) any { return export.UnixTime(r.StartAt) }},
	{Key: "settle_at", TitleKey: "export.column.settle_at", Value: func(r model.LivePkLog) any { return export.UnixTime(r.SettleAt) }},
}

var columnValue = export.ValueMapOf(exportColumns)

// exportFilterInput 前端传来的筛选条件，字段与列表接口同口径
type exportFilterInput struct {
	RoomID     *int64   `json:"room_id"`
	RivalUID   *int64   `json:"rival_uid"`
	RivalUname *string  `json:"rival_uname"`
	Result     *int     `json:"result"`
	StartAt    *[]int64 `json:"start_at"`
}

// exportFilterQuery 规范化后的筛选条件：时间区间已换算成秒级的闭区间
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
	// 仓储的构建器不校验 result 就直接拼进 SQL，非法值会静默变成「命中 0 行」而不是报错
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
	return export.BuildRecords(rows, keys, columnValue, func(r model.LivePkLog) int64 { return r.ID })
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
