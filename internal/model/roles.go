package model

import "github.com/zxc7563598/bilibili-live-assistant/internal/enum"

type Role struct {
	ID     int64       `gorm:"primaryKey"`
	Code   string      `gorm:"type:varchar(100);not null;uniqueIndex;comment:编码"`
	Name   string      `gorm:"type:varchar(100);not null;comment:名称"`
	Enable enum.Enable `gorm:"type:smallint;not null;default:0;comment:是否启用"`
	BaseModel
}

func (Role) TableName() string {
	return "roles"
}

// RoleListPageQuery 用于后台列表查询入参，不对应数据库表
type RoleListPageQuery struct {
	Offset int
	Limit  int
	Name   *string
	Enable *int
}

// RoleListItem 用于后台列表展示，不对应数据库表
type RoleListItem struct {
	ID     int64
	Code   string
	Name   string
	Enable enum.Enable
}

// RoleUpdateByIdForm 用于更新角色资料，不对应数据库表
type RoleUpdateByIdForm struct {
	Code   *string
	Name   *string
	Enable *enum.Enable
}
