package model

type LiveUser struct {
	ID    int64  `gorm:"primaryKey"`
	Uid   int64  `gorm:"not null;comment:用户uid"`
	Uname string `gorm:"type:varchar(100);not null;comment:用户名称"`
	BaseModel
}

func (LiveUser) TableName() string {
	return "live_users"
}

// LiveUserListPageQuery 用户分页查询入参，不对应数据库表
type LiveUserListPageQuery struct {
	UID    *int64
	Uname  *string
	Offset int
	Limit  int
}
