package liveguard

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_guard_record"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
)

// rosterExporter 大航海名单的导出数据源
type rosterExporter struct{ svc *Service }

// NewExporter 创建大航海名单导出数据源
func NewExporter(svc *Service) export.Source {
	return rosterExporter{svc: svc}
}

// rosterExportColumns 允许导出的列，顺序与页面列一致
//
// 月份两列的格式化与页面同一份实现（formatMonth / formatCoverTo），导出文件与页面
// 说的必须是同一回事。
var rosterExportColumns = []export.ColumnSpec[live_guard_record.RosterItem]{
	{Key: "uid", TitleKey: "export.column.uid", Value: func(r live_guard_record.RosterItem) any { return r.UID }},
	{Key: "uname", TitleKey: "export.column.uname", Value: func(r live_guard_record.RosterItem) any { return r.Uname }},
	{Key: "guard_level", TitleKey: "export.column.guard_level", Value: func(r live_guard_record.RosterItem) any { return r.GuardLevel }},
	{Key: "cover_from_month", TitleKey: "export.column.cover_from_month", Value: func(r live_guard_record.RosterItem) any { return formatMonth(r.CoverFrom) }},
	{Key: "cover_to_month", TitleKey: "export.column.cover_to_month", Value: func(r live_guard_record.RosterItem) any { return formatCoverTo(r.CoverTo) }},
	{Key: "open_num", TitleKey: "export.column.open_num", Value: func(r live_guard_record.RosterItem) any { return r.OpenNum }},
	{Key: "open_at", TitleKey: "export.column.open_at", Value: func(r live_guard_record.RosterItem) any { return export.UnixTime(r.OpenAt) }},
}

var rosterColumnValue = export.ValueMapOf(rosterExportColumns)

// rosterExportFilterInput 前端传入的筛选条件
type rosterExportFilterInput struct {
	GuardLevel *int    `json:"guard_level"`
	Month      string  `json:"month"`
	UID        *int64  `json:"uid"`
	Uname      *string `json:"uname"`
}

// rosterExportFilterQuery 规范化后的筛选条件：月份已换算成秒级区间
type rosterExportFilterQuery struct {
	GuardLevel *int    `json:"guard_level"`
	MonthFrom  int64   `json:"month_from"`
	MonthTo    int64   `json:"month_to"`
	UID        *int64  `json:"uid"`
	Uname      *string `json:"uname"`
}

// Module 导出模块标识
func (e rosterExporter) Module() string { return "liveguard" }

// Columns 本模块允许导出的列
func (e rosterExporter) Columns() []export.Column { return export.ColumnsOf(rosterExportColumns) }

// Normalize 解析并校验前端筛选条件，把月份换算成秒级区间后回写为规范 JSON
//
// guard_level 必传：名单按 uid 聚合，只有固定档位才能保证 uid 唯一，
// 而导出要靠 uid 做分块游标。
func (e rosterExporter) Normalize(filters json.RawMessage) (json.RawMessage, int, error) {
	var in rosterExportFilterInput
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &in); err != nil {
			return nil, CodeParamInvalid, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	// 仓储的构建器对非法枚举是**静默忽略**的，不在这里拦下来就会导出一份筛选条件没生效的全量数据
	if in.GuardLevel == nil {
		return nil, CodeGuardLevelInvalid, fmt.Errorf("guard_level 未传")
	}
	if _, ok := parseGuardLevel(*in.GuardLevel); !ok {
		return nil, CodeGuardLevelInvalid, fmt.Errorf("guard_level 内容非法: %d", *in.GuardLevel)
	}
	from, to, err := monthRange(in.Month)
	if err != nil {
		return nil, CodeParamInvalid, err
	}
	out, err := json.Marshal(rosterExportFilterQuery{
		GuardLevel: in.GuardLevel,
		MonthFrom:  from,
		MonthTo:    to,
		UID:        in.UID,
		Uname:      in.Uname,
	})
	if err != nil {
		return nil, CodeParamInvalid, fmt.Errorf("序列化导出筛选条件失败: %w", err)
	}
	return out, 0, nil
}

// Count 统计名单人数，limit > 0 时提前停止
func (e rosterExporter) Count(ctx context.Context, filters json.RawMessage, limit int) (int64, error) {
	query, err := parseRosterExportQuery(filters)
	if err != nil {
		return 0, err
	}
	return e.svc.liveGuardRepo.CountRoster(ctx, nil, query, limit)
}

// FetchChunk 取一块名单
//
// 名单按 uid 聚合、档位固定，所以 uid 就是唯一的分块游标，
// 不需要用主键——用主键会把同一用户的多段拆到不同块里。
func (e rosterExporter) FetchChunk(ctx context.Context, filters json.RawMessage, afterID int64, keys []string, limit int) ([][]any, int64, error) {
	query, err := parseRosterExportQuery(filters)
	if err != nil {
		return nil, 0, err
	}
	rows, err := e.svc.liveGuardRepo.RosterChunk(ctx, nil, query, afterID, limit)
	if err != nil {
		return nil, 0, err
	}
	return export.BuildRecords(rows, keys, rosterColumnValue, func(r live_guard_record.RosterItem) int64 { return r.UID })
}

// parseRosterExportQuery 把规范化后的 JSON 还原成仓储查询入参
func parseRosterExportQuery(filters json.RawMessage) (model.LiveGuardRecordListPageQuery, error) {
	var in rosterExportFilterQuery
	if err := json.Unmarshal(filters, &in); err != nil {
		return model.LiveGuardRecordListPageQuery{}, fmt.Errorf("解析导出筛选条件失败: %w", err)
	}
	return model.LiveGuardRecordListPageQuery{
		GuardLevel: in.GuardLevel,
		MonthFrom:  in.MonthFrom,
		MonthTo:    in.MonthTo,
		UID:        in.UID,
		Uname:      in.Uname,
	}, nil
}
