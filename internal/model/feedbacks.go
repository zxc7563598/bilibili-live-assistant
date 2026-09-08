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
