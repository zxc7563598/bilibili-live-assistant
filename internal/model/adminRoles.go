package model

type AdminRole struct {
	ID      uint64 `gorm:"primaryKey"`
	AdminID uint64 `gorm:"not null;default:0;uniqueIndex:uk_admin_role;comment:管理员ID"`
	RoleID  uint64 `gorm:"not null;default:0;uniqueIndex:uk_admin_role;comment:角色ID"`
	BaseModel
}

func (AdminRole) TableName() string {
	return "admin_roles"
}

// AdminRoleListItem 用于后台列表展示，不对应数据库表
type AdminRoleListItem struct {
	ID      uint64
	AdminID uint64
	RoleID  uint64
}
