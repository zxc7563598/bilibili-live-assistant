package liveuser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
)

// exportColumns 允许导出的列。
//
// 不含 password / token：model 里有这两列，它们是用户凭证，不能出现在导出文件里。
var exportColumns = []export.ColumnSpec[model.LiveUser]{
	{Key: "uid", TitleKey: "export.column.uid", Value: func(r model.LiveUser) any { return r.UID }},
	{Key: "uname", TitleKey: "export.column.uname", Value: func(r model.LiveUser) any { return r.Uname }},
	{Key: "points", TitleKey: "export.column.points", Value: func(r model.LiveUser) any { return r.Points }},
	{Key: "stars", TitleKey: "export.column.stars", Value: func(r model.LiveUser) any { return r.Stars }},
	{Key: "total_danmu_count", TitleKey: "export.column.total_danmu_count", Value: func(r model.LiveUser) any { return r.TotalDanmuCount }},
	// 落库单位是分，与页面 row.total_gift_amount / 100 的展示口径一致
	{Key: "total_gift_amount", TitleKey: "export.column.total_gift_amount", Value: func(r model.LiveUser) any { return export.Money(r.TotalGiftAmount) }},
}

var columnValue = export.ValueMapOf(exportColumns)

// exportFilterInput 前端传来的筛选条件，字段与列表接口同口径
type exportFilterInput struct {
	UID   *int64  `json:"uid"`
	Uname *string `json:"uname"`
}

// Module 导出模块标识
func (s *Service) Module() string { return "liveuser" }

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
	return s.liveUserRepo.CountFiltered(ctx, nil, q, limit)
}

// FetchChunk 取一块数据，按主键倒序
func (s *Service) FetchChunk(ctx context.Context, filters json.RawMessage, afterID int64, keys []string, limit int) ([][]any, int64, error) {
	q, err := parseExportQuery(filters)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.liveUserRepo.ExportChunk(ctx, nil, q, afterID, limit)
	if err != nil {
		return nil, 0, err
	}
	return export.BuildRecords(rows, keys, columnValue, func(r model.LiveUser) int64 { return r.ID })
}

// parseExportQuery 把规范化的筛选条件还原为仓储查询结构
func parseExportQuery(filters json.RawMessage) (model.LiveUserListPageQuery, error) {
	var f exportFilterInput
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &f); err != nil {
			return model.LiveUserListPageQuery{}, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	return model.LiveUserListPageQuery{UID: f.UID, Uname: f.Uname}, nil
}
