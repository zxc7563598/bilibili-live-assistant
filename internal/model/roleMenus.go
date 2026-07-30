package model

type RoleMenu struct {
	ID     uint64 `gorm:"primaryKey"`
	RoleID uint64 `gorm:"not null;default:0;comment:角色ID"`
	MenuID uint64 `gorm:"not null;default:0;comment:菜单ID"`
	BaseModel
}

func (RoleMenu) TableName() string {
	return "role_menus"
}
