package feedback

import (
	"context"
	"errors"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/sqlutil"
	"gorm.io/gorm"
)

// sortColumns 允许排序的字段 → 实际 SQL 列，键与前端表格列的 key 对齐
var sortColumns = map[string]string{
	"id":         "feedbacks.id",
	"type":       "feedbacks.type",
	"contact":    "feedbacks.contact",
	"user_id":    "feedbacks.user_id",
	"created_at": "feedbacks.created_at",
	"updated_at": "feedbacks.updated_at",
	"uid":        "lu.uid",
	"uname":      "lu.uname",
}

// sortOrder 允许的排序方向 → SQL 方向。
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

type Repository interface {
	base.Repository[model.Feedback]
	// ListPage 分页查询投诉，联查 live_users 补充 uid/uname，
	// 支持 UID 精确、Uname 模糊筛选与白名单字段排序；列表不返回 content
	ListPage(ctx context.Context, tx *gorm.DB, query model.FeedbackListPageQuery) ([]model.FeedbackListItem, int64, error)
	// GetDetailByID 按主键查询投诉详情，联查 live_users 补充 uid/uname；不存在返回 (nil, nil)
	GetDetailByID(ctx context.Context, tx *gorm.DB, id int64) (*model.FeedbackListItem, error)
}

// ListPage 分页查询投诉，联查 live_users 补充 uid/uname。
// 列表中不展示投诉正文，因此不 select content（TEXT 列，属于大字段）。
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.FeedbackListPageQuery) ([]model.FeedbackListItem, int64, error) {
	var list []model.FeedbackListItem
	var total int64
	// 投诉是历史记录、用户可能被移除，用 LEFT JOIN 保证投诉不丢；
	// JOIN 打在 live_users 主键上不放大行数，count(*) 依然精确，无需 DISTINCT。
	// lu 侧不过滤 deleted_at 是有意为之（对齐 live_user_order 的处理）。
	db := r.ResolveDB(ctx, tx).Model(&model.Feedback{}).
		Select("feedbacks.id, feedbacks.user_id, feedbacks.type, feedbacks.contact, feedbacks.created_at, lu.uid, lu.uname").
		Joins("LEFT JOIN live_users lu ON lu.id = feedbacks.user_id")
	if v := query.UID; v != nil {
		db = db.Where("lu.uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("lu.uname LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 默认按最新投诉在前；排序参数非法/缺失时静默回退该默认
	orderClause := "feedbacks.created_at desc, feedbacks.id desc"
	if query.SortField != nil && query.SortOrder != nil {
		// 方向先小写归一化再查白名单，非法值直接回退默认
		if field, ok := sortColumns[*query.SortField]; ok {
			if dir, ok := sortOrder[strings.ToLower(*query.SortOrder)]; ok {
				// field/dir 均来自字面量白名单，杜绝注入；id asc 保证同键值时翻页稳定
				orderClause = field + " " + dir + ", feedbacks.id asc"
			}
		}
	}
	err := db.Order(orderClause).Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	return list, total, err
}

// GetDetailByID 按主键查询投诉详情，联查 live_users 补充 uid/uname；不存在返回 (nil, nil)
func (r *gormRepo) GetDetailByID(ctx context.Context, tx *gorm.DB, id int64) (*model.FeedbackListItem, error) {
	var item model.FeedbackListItem
	err := r.ResolveDB(ctx, tx).Model(&model.Feedback{}).
		Select("feedbacks.*, lu.uid, lu.uname").
		Joins("LEFT JOIN live_users lu ON lu.id = feedbacks.user_id").
		Where("feedbacks.id = ?", id).
		Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
