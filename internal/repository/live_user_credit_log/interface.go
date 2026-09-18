package live_user_credit_log

import (
	"context"
	"strings"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/base"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/sqlutil"
	"gorm.io/gorm"
)

// sortColumns 允许参与 ListPage 排序的 DB 列
var sortColumns = map[string]string{
	"id":            "id",
	"user_id":       "user_id",
	"credit_type":   "credit_type",
	"change_type":   "change_type",
	"change_amount": "change_amount",
	"before_value":  "before_value",
	"after_value":   "after_value",
	"biz_type":      "biz_type",
	"remark":        "remark",
	"operator_type": "operator_type",
	"operator_id":   "operator_id",
	"created_at":    "created_at",
	"updated_at":    "updated_at",
}

// sortOrder 允许的排序方向 → SQL 方向。
var sortOrder = map[string]string{
	"ascend":  "asc",
	"descend": "desc",
}

// Repository 接口定义
type Repository interface {
	base.Repository[model.LiveUserCreditLog]
	// ListPage 分页查询积分/星光变动记录，联查 live_users 补充 uid/uname/face
	ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserCreditLogListPageQuery) ([]model.LiveUserCreditLogListItem, int64, error)
}

// ListPage 分页查询积分/星光变动记录，联查 live_users 补充 uid/uname/face
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.LiveUserCreditLogListPageQuery) ([]model.LiveUserCreditLogListItem, int64, error) {
	var list []model.LiveUserCreditLogListItem
	var total int64
	db := r.getDB(ctx, tx)
	// 日志流水不可变、用户可能被移除，用 LEFT JOIN 保证历史流水不丢；
	// 用户侧只取列表需要的 uid/uname/face
	db = db.Model(&model.LiveUserCreditLog{}).
		Select("live_user_credit_logs.*, lu.uid, lu.uname, lu.face").
		Joins("LEFT JOIN live_users lu ON lu.id = live_user_credit_logs.user_id")
	if v := query.UID; v != nil {
		db = db.Where("lu.uid = ?", *v)
	}
	if v := query.Uname; v != nil && *v != "" {
		db = db.Where("lu.uname LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
	}
	if v := query.UserID; v != nil {
		db = db.Where("live_user_credit_logs.user_id = ?", *v)
	}
	if v := query.CreditType; v != nil {
		c := enum.CreditType(*v)
		if c.IsValid() {
			db = db.Where("live_user_credit_logs.credit_type = ?", c)
		}
	}
	if v := query.ChangeType; v != nil {
		c := enum.ChangeType(*v)
		if c.IsValid() {
			db = db.Where("live_user_credit_logs.change_type = ?", c)
		}
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 默认与现状一致；排序参数非法/缺失时静默回退该默认
	orderClause := "live_user_credit_logs.created_at desc, live_user_credit_logs.id desc"
	if query.SortField != nil && query.SortOrder != nil {
		// 方向先小写归一化再查白名单，非法值直接回退默认
		if field, ok := sortColumns[*query.SortField]; ok {
			if dir, ok := sortOrder[strings.ToLower(*query.SortOrder)]; ok {
				// field/dir 均来自字面量白名单，杜绝注入；id desc 兜底，同键值时翻页稳定
				orderClause = "live_user_credit_logs." + field + " " + dir + ", live_user_credit_logs.id desc"
			}
		}
	}
	err := db.Order(orderClause).Offset(query.Offset).Limit(query.Limit).Find(&list).Error
	return list, total, err
}
