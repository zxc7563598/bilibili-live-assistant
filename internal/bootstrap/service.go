package bootstrap

import (
	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/GoAdminKit/internal/config"
	"github.com/zxc7563598/GoAdminKit/internal/service/admin"
	"github.com/zxc7563598/GoAdminKit/internal/service/altcha"
	"github.com/zxc7563598/GoAdminKit/internal/service/menu"
	"github.com/zxc7563598/GoAdminKit/internal/service/role"
	"gorm.io/gorm"
)

type Services struct {
	Admin  admin.Service
	Role   role.Service
	Menu   menu.Service
	Altcha altcha.Service
}

func InitServices(repo *Repositories, db *gorm.DB, rdb *redis.Client, cfg *config.Config) *Services {
	return &Services{
		Admin:  *admin.New(repo.Admin, repo.AdminRole, repo.Role, db, rdb),
		Role:   *role.New(repo.Role, repo.Admin, repo.RoleMenu, repo.AdminRole, repo.Menu, db, rdb),
		Menu:   *menu.New(repo.Menu),
		Altcha: *altcha.New(cfg.Altcha.HMACKey),
	}
}
