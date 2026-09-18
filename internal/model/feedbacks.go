package model

type Feedback struct {
	ID      int64  `gorm:"primaryKey"`
	Type    string `gorm:"type:varchar(100);not null;comment:问题类型"`
	UserID  int64  `gorm:"not null;index;comment:用户id"`
	Content string `gorm:"type:text;not null;comment:投诉内容"`
	Contact string `gorm:"type:varchar(100);not null;comment:联系方式"`
	BaseModel
}

func (Feedback) TableName() string {
	return "feedbacks"
}

// FeedbackListItem 投诉列表/详情查询结果：feedbacks 联查 live_users 补充 uid/uname，不对应数据库表
type FeedbackListItem struct {
	UID   int64
	Uname string
	Feedback
}

// FeedbackListPageQuery 投诉分页查询入参，不对应数据库表
type FeedbackListPageQuery struct {
	// UID 联查 live_users.uid，按 B 站 UID 精确筛选
	UID *int64
	// Uname 联查 live_users.uname，按昵称模糊筛选
	Uname     *string
	SortField *string
	SortOrder *string
	Offset    int
	Limit     int
}
