package model

import "github.com/zxc7563598/GoAdminKit/internal/enum"

type Menu struct {
	ID        uint64        `gorm:"primaryKey"`
	Code      string        `gorm:"type:varchar(100);not null;uniqueIndex;comment:编码"`
	Enable    enum.Enable   `gorm:"type:smallint;not null;default:0;comment:是否启用"`
	Show      enum.YesNo    `gorm:"type:smallint;not null;default:0;comment:显示状态"`
	KeepAlive enum.YesNo    `gorm:"type:smallint;not null;default:0;comment:是否KeepAlive"`
	Layout    string        `gorm:"type:varchar(100);not null;default:'';comment:layout"`
	Type      enum.MenuType `gorm:"type:varchar(50);not null;default:'';comment:类型"`
	ParentID  uint64        `gorm:"not null;default:0;index;comment:父级ID"`
	Name      string        `gorm:"type:varchar(100);not null;comment:名称"`
	Icon      string        `gorm:"type:varchar(100);not null;comment:菜单图标"`
	Path      string        `gorm:"type:varchar(255);not null;comment:路由地址"`
	Component string        `gorm:"type:varchar(255);not null;comment:组件路径"`
	Order     int           `gorm:"type:int;not null;comment:排序"`
	BaseModel
}

func (Menu) TableName() string {
	return "menus"
}

// MenuUpdateByIdForm 用于更新菜单基本信息，不对应数据库表
type MenuUpdateByIdForm struct {
	Code      string        `json:"code"`
	Enable    enum.Enable   `json:"enable"`
	Show      enum.YesNo    `json:"show"`
	KeepAlive enum.YesNo    `json:"keep_alive"`
	Layout    string        `json:"layout"`
	Type      enum.MenuType `json:"type"`
	ParentID  uint64        `json:"parent_id"`
	Name      string        `json:"name"`
	Icon      string        `json:"icon"`
	Path      string        `json:"path"`
	Component string        `json:"component"`
	Order     int           `json:"order"`
}
