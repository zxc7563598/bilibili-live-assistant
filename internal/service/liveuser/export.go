package liveuser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/export"
)

// 用户列表的导出数据源。
//
// 管理端的「用户列表」与商城端的「用户管理」打的是同一个 /liveuser/list，列也完全一致，
// 所以两处页面共用一个模块名。

// exportColumns 允许导出的列。
//
// 不含裸 id；也不含 password / token —— model 里有这两列，它们是凭证，绝不能出现在导出文件里。
var exportColumns = []export.ColumnSpec[model.LiveUser]{
	{Key: "uid", TitleKey: "export.column.uid", Value: func(r model.LiveUser) any { return r.UID }},
	{Key: "uname", TitleKey: "export.column.uname", Value: func(r model.LiveUser) any { return r.Uname }},
	{Key: "points", TitleKey: "export.column.points", Value: func(r model.LiveUser) any { return r.Points }},
	{Key: "stars", TitleKey: "export.column.stars", Value: func(r model.LiveUser) any { return r.Stars }},
	{Key: "total_danmu_count", TitleKey: "export.column.total_danmu_count", Value: func(r model.LiveUser) any { return r.TotalDanmuCount }},
	// 页面按 row.total_gift_amount / 100 展示，落库单位是分
	{Key: "total_gift_amount", TitleKey: "export.column.total_gift_amount", Value: func(r model.LiveUser) any { return export.Money(r.TotalGiftAmount) }},
}

// columnValue 列 key → 取值函数，与上面的声明同源
var columnValue = export.ValueMapOf(exportColumns)

// exportFilter 前端传来的筛选条件，字段与列表接口同口径
type exportFilter struct {
	UID   *int64  `json:"uid"`
	Uname *string `json:"uname"`
}

// Module 导出模块标识
func (s *Service) Module() string { return "liveuser" }

// Columns 本模块允许导出的列
func (s *Service) Columns() []export.Column { return export.ColumnsOf(exportColumns) }

// Normalize 解析并校验前端筛选条件。
//
// 本模块没有时间区间一类的需要换算的字段，这里只做一次解析确认前端传的是合法 JSON，
// 再回写去掉未知字段的规范形式。
func (s *Service) Normalize(filters json.RawMessage) (json.RawMessage, int, error) {
	f, err := parseExportFilter(filters)
	if err != nil {
		return nil, CodeParamInvalid, err
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

// parseExportFilter 解析筛选条件
func parseExportFilter(filters json.RawMessage) (exportFilter, error) {
	var f exportFilter
	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &f); err != nil {
			return exportFilter{}, fmt.Errorf("解析导出筛选条件失败: %w", err)
		}
	}
	return f, nil
}

// parseExportQuery 把规范化的筛选条件还原为仓储查询结构
func parseExportQuery(filters json.RawMessage) (model.LiveUserListPageQuery, error) {
	f, err := parseExportFilter(filters)
	if err != nil {
		return model.LiveUserListPageQuery{}, err
	}
	return model.LiveUserListPageQuery{UID: f.UID, Uname: f.Uname}, nil
}
