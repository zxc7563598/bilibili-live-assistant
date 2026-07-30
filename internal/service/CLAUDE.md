# CLAUDE.md — Service 层

## 职责边界

Service 层是**业务逻辑核心**，负责流程编排，不直接操作数据库。

核心职责：
1. 组合 Repository 提供的数据访问接口完成业务流程
2. 封装操作流程（事务、校验、日志记录等）
3. 对外只暴露业务方法，不泄露数据库字段

**严禁**：
- 直接操作数据库（必须通过 Repository）
- 将 Repository 返回的 Model 结构体透传给 Handler（应通过 DTO 转换）

## 文件组织

每个 Service 模块包含三个文件：

```
internal/service/<模块名>/
├── dto.go       # Service 入参与出参定义（与 Handler DTO 解耦）
├── common.go    # 通用方法拆分（复杂逻辑的子步骤）
└── service.go   # 核心业务逻辑编排
```

### dto.go

```go
type SaveForm struct {
    ID       *uint64
    Username string
    Status   int
}

type ListPageResp struct {
    Total    int64
    PageData []ListItem
}
```

入参与出参在 Service 内部定义，与 Handler 层的 `dto/input` / `dto/resp` 解耦，保持层间独立性。

### common.go

将复杂操作拆分为小方法，例如：

```go
func (s *Service) insert(form SaveForm) error { /* ... */ }
func (s *Service) update(form SaveForm) error { /* ... */ }
func (s *Service) validateUnique(username string) (bool, error) { /* ... */ }
```

### service.go

只负责流程编排：

```go
func (s *Service) Save(ctx context.Context, adminID uint64, form SaveForm) (int, error) {
    isCreate := form.ID == nil || *form.ID == 0
    if isCreate {
        if err := s.insert(ctx, form); err != nil {
            return 40001, err
        }
    } else {
        if err := s.update(ctx, form); err != nil {
            return 40002, err
        }
    }
    // 记录操作日志
    errCode, err := s.saveLogRepo.CreateByAdminID(ctx, adminID, isCreate)
    if errCode > 0 {
        return errCode, err
    }
    return 0, nil
}
```

返回值约定：`(int, error)` — 第一个返回值是错误码（0 表示成功），第二个是原始 error（用于日志）。

## 依赖注入

### Service 结构体

```go
type Service struct {
    saveLogRepo save_log.Repository
}

func New(saveLogRepo save_log.Repository) *Service {
    return &Service{saveLogRepo: saveLogRepo}
}
```

关键原则：**只注入当前模块真正需要的 Repository，避免"顺手全注入"引入不必要依赖。**

### 模块注册

新增 Service 后，在 `internal/bootstrap/service.go` 中注册：

```go
type Services struct {
    Admin *admin.Service
}

func InitServices(repo *Repositories, db *gorm.DB, rdb *redis.Client) *Services {
    return &Services{
        Admin: admin.New(repo.SaveLog),
    }
}
```

## 核心约定

1. **不直接操作数据库**：所有数据操作通过 Repository
2. **入参与出参用 DTO**：避免 Repository 层的 Model 结构体被传递到 Handler
3. **Service 只做流程编排**：复杂子步骤拆分到 `common.go`
4. **按需注入依赖**：不注入与当前 Service 无关的 Repository
5. **返回值格式**：`(errorCode int, err error)`，errorCode=0 表示成功
