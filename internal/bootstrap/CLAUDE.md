# CLAUDE.md — Bootstrap（依赖注入与初始化）

## 职责

`internal/bootstrap/` 是整个应用的装配中心，负责：

1. 配置加载
2. 数据库连接初始化
3. 各层依赖注入（Repository → Service → Handler）
4. 路由注册与中间件挂载

## 文件结构

```
internal/bootstrap/
├── app.go          # 应用启动入口（编排初始化流程）
├── repository.go   # Repository 依赖注入
├── service.go      # Service 依赖注入
├── handler.go      # Handler 依赖注入
└── router.go       # 路由注册
```

## 初始化流程

在 `app.go` 中自下而上装配：

```go
// 1. 加载配置
config.Load()

// 2. 初始化数据库
db := initDB()

// 3. Repository 层
repos := InitRepositories(db)

// 4. Service 层
services := InitServices(repos, db, rdb)

// 5. Handler 层
handlers := InitHandlers(services)

// 6. 注册路由
r := gin.New()
r = RouteRegister(r, rdb, handlers)
```

## repository.go

```go
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
```

## service.go

```go
type Services struct {
    Admin *admin.Service
    Role  *role.Service
    Menu  *menu.Service
}

func InitServices(repo *Repositories, db *gorm.DB, rdb *redis.Client) *Services {
    return &Services{
        Admin: admin.New(repo.Admin, repo.AdminRole, repo.Role),
        Role:  role.New(repo.Role, repo.RoleMenu, db),
        Menu:  menu.New(repo.Menu, repo.RoleMenu),
    }
}
```

## handler.go

```go
type Handlers struct {
    Admin *admin.Handler
    Role  *role.Handler
    Menu  *menu.Handler
}

func InitHandlers(svc *Services) *Handlers {
    return &Handlers{
        Admin: admin.New(svc.Admin),
        Role:  role.New(svc.Role),
        Menu:  menu.New(svc.Menu),
    }
}
```

## router.go

```go
func RouteRegister(r *gin.Engine, rdb *redis.Client, handlers *Handlers) *gin.Engine {
    // 中间件
    r.Use(middleware.Language())
    r.Use(middleware.CORS())

    // 公开路由
    public := r.Group("/api")
    {
        public.POST("/login", handlers.Admin.Login)
    }

    // 需认证路由
    auth := r.Group("/api")
    auth.Use(middleware.Auth(rdb))
    {
        auth.GET("/admin/list", handlers.Admin.ListPage)
        // ...
    }

    return r
}
```

## 新增模块注册清单

每次新增一个业务模块时，需要依次在以下文件注册：

1. [ ] `repository.go` — 注册 Repository 实例
2. [ ] `service.go` — 注册 Service 实例，按需注入 Repository
3. [ ] `handler.go` — 注册 Handler 实例，注入对应 Service
4. [ ] `router.go` — 注册路由，挂载中间件

## 核心原则

> **只注入当前模块真正需要的依赖，避免"顺手全注入"。**

这样可以：
- 保持模块边界清晰
- 降低耦合度
- 提升可测试性
- 避免误操作（如 Service 拿到了不该用的 Repository）
