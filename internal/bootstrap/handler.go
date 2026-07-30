package bootstrap

import (
	"github.com/zxc7563598/GoAdminKit/internal/handler/admin"
	"github.com/zxc7563598/GoAdminKit/internal/handler/altcha"
	"github.com/zxc7563598/GoAdminKit/internal/handler/menu"
	"github.com/zxc7563598/GoAdminKit/internal/handler/role"
)

type Handlers struct {
	Admin  *admin.Handler
	Menu   *menu.Handler
	Role   *role.Handler
	Altcha *altcha.Handler
}

func InitHandlers(svc *Services) *Handlers {
	return &Handlers{
		Admin:  admin.New(&svc.Admin, &svc.Altcha),
		Menu:   menu.New(&svc.Menu),
		Role:   role.New(&svc.Role),
		Altcha: altcha.New(&svc.Altcha),
	}
}
