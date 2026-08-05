package model

type AdminRole struct {
	ID      int64 `gorm:"primaryKey"`
	AdminID int64 `gorm:"not null;default:0;uniqueIndex:uk_admin_role;comment:管理员ID"`
	RoleID  int64 `gorm:"not null;default:0;uniqueIndex:uk_admin_role;comment:角色ID"`
	BaseModel
}

func (AdminRole) TableName() string {
	return "admin_roles"
}

// AdminRoleListItem 用于后台列表展示，不对应数据库表
type AdminRoleListItem struct {
	ID      int64
	AdminID int64
	RoleID  int64
}
