package bootstrap

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/admin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/admin_role"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/menu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/role"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/role_menu"
	"gorm.io/gorm"
)

type Repositories struct {
	Admin     admin.Repository
	Role      role.Repository
	Menu      menu.Repository
	AdminRole admin_role.Repository
	RoleMenu  role_menu.Repository
}

func InitRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Admin:     admin.New(db),
		Role:      role.New(db),
		Menu:      menu.New(db),
		AdminRole: admin_role.New(db),
		RoleMenu:  role_menu.New(db),
	}
}
