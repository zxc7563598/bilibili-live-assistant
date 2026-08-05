package model

type LiveUser struct {
	ID    int64 `gorm:"primaryKey"`
	Uid   int64 `gorm:"not null;comment:用户uid"`
	Uname string `gorm:"type:varchar(100);not null;comment:用户名称"`
	BaseModel
}

func (LiveUser) TableName() string {
	return "live_users"
}
