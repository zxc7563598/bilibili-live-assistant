package model

type RoleMenu struct {
	ID     int64 `gorm:"primaryKey"`
	RoleID int64 `gorm:"not null;default:0;comment:角色ID"`
	MenuID int64 `gorm:"not null;default:0;comment:菜单ID"`
	BaseModel
}

func (RoleMenu) TableName() string {
	return "role_menus"
}
