package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

type LiveUser struct {
	ID              int64       `gorm:"primaryKey"`
	UID             int64       `gorm:"not null;uniqueIndex;comment:用户uid"`
	Uname           string      `gorm:"type:varchar(100);not null;comment:用户名称"`
	Password        string      `gorm:"type:varchar(255);not null;comment:密码"`
	Token           *string     `gorm:"type:varchar(255);comment:登录凭证"`
	Face            string      `gorm:"type:varchar(255);not null;comment:用户头像URL"`
	Points          int64       `gorm:"not null;default:0;comment:用户积分"`
	Stars           int64       `gorm:"not null;default:0;comment:用户星光"`
	TotalDanmuCount int64       `gorm:"not null;default:0;comment:累计发送弹幕数"`
	TotalGiftAmount int64       `gorm:"not null;default:0;comment:累计赠送礼物金额(分)"`
	Enable          enum.Enable `gorm:"type:smallint;not null;default:1;comment:是否启用"`
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
