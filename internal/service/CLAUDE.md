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

每个 Service 模块默认包含三个文件：

```
internal/service/<模块名>/
├── dto.go       # 本模块的入参/出参类型
├── common.go    # 子步骤方法、转换函数
└── service.go   # 流程编排
```

**例外**：这是「请求型 CRUD 模块」的默认布局，不是硬性约束。`altcha` 只有
`service.go`（没有自己的 DTO，直接用库的 `*altcha.Challenge`，也没有拆出的助手），
`live` 有 16 个文件（长驻的有状态服务，另有 receiver / dispatcher / hub 等）。
为凑格式而建空文件没有意义。

### `dto.go` —— 入参与出参

```go
// 分页请求统一嵌入 pkg/pagination 的 PageResp
type ListPageReq struct {
    pagination.PageResp
    Name   *string
    Enable *int
}

type ListPageResp struct {
    Total    int64
    PageData []ListPageItem
}

type ListPageItem struct {
    ID     int64
    Code   string
    Name   string
    Enable bool
}
```

入参与出参在 Service 内部定义，与 Handler 层的 `dto/input` / `dto/resp` 解耦。

**分页结构体不要各包自己写一份**：`PageResp` + `OffsetLimit()`（页码最小 1、
页大小默认 10、上限 100）已在 `pkg/pagination`，嵌入即可。

### `common.go` —— 子步骤

把复杂流程拆成小方法，例如：

```go
func (s *Service) add(ctx context.Context, tx *gorm.DB, req SaveReq) (int64, int, error)
func (s *Service) update(ctx context.Context, tx *gorm.DB, req SaveReq) (int64, int, error)
func (s *Service) toListPageItems(list []model.Admin) []ListPageItem
```

关键约定：**凡是写库的子步骤都要收 `tx *gorm.DB`**，由调用方（`service.go` 里的
事务）传进来。自己开事务的子步骤要在名字或注释里说清楚。

### `service.go` —— 流程编排

只做编排，不写具体步骤：

```go
func (s *Service) Save(ctx context.Context, req SaveReq) (int, error) {
    var errCode int
    err := s.db.Transaction(func(tx *gorm.DB) error {
        var roleID int64
        var err error
        if req.ID == nil || *req.ID == 0 {
            roleID, errCode, err = s.add(ctx, tx, req)
        } else {
            roleID, errCode, err = s.update(ctx, tx, req)
        }
        if err != nil {
            return err
        }
        // ...
        return nil
    })
    if err != nil {
        if errCode != 0 {
            return errCode, err
        }
        return CodeSaveFailed, err
    }
    return 0, nil
}
```

## 返回值约定

三种形态并存，按用途选：

| 形态 | 用在 | 例 |
|------|------|-----|
| `(payload, int, error)` | 读方法与多数写方法 | `ListPage`、`Details` |
| `(int, error)` | 不需要返回数据的写方法 | `Delete`、`UpdateEnable` |
| `error` | 定时任务等非 HTTP 入口 | `ExpireDrafts`、`UnmuteDueUsers` |

**两条容易踩的规则**：

1. **以错误码为准，不以 error 为准**。`errorCode == 0` 才表示成功；校验失败常写成
   `(zero, 错误码, nil)`——**有错误码但 error 为 nil**。调用方只判断 `err != nil`
   会把一次被拒绝的请求当成成功。
2. **事务闭包里必须同时判 error 与错误码**。只写 `if errCode > 0 { return err }`
   的话，`err` 为 nil 时闭包返回 nil、事务照常提交、错误码被丢掉。正确写法见
   `admin.Save` 与 `role.Save`。

## 方法命名约定

承接 [repository/CLAUDE.md](../repository/CLAUDE.md) 的词汇，加上服务层特有的：

- 分页查询一律叫 `ListPage`，返回 `(列表, 总数, error)`
- 读方法有动词：`Get*` 返回单条/整体，`List*` 返回集合。**不要用名词当方法名**
  （`Manifest()` → `GetManifest()`，`UserInfo()` → `GetUserInfo()`）
- 树形结果用 `Get*` 而非 `List*`：`GetMenuTree`
- 名字要能看出来有副作用：会发网络请求的用 `Fetch*`，会翻转状态的用 `Toggle*`，
  只按主键写单列的用 `Update<列名>ByID`
- 子步骤（`common.go` 里的小写方法）用动词开头：`add` / `update` / `syncXxx` /
  `resetXxx` / `toXxx`
- 谓词用 `XxxExists`（`MenuExists`、`ExistsAccount`）

## 错误码约定

**常量声明在各模块的 `code.go`**，与 `internal/i18n/locales/*.yaml` 的 `error`
节点一一对应，常量旁的注释取自 YAML 里的原始说明。

```go
const (
    CodeCodeRequired   = 10202 // 请输入角色标识
    CodeNameRequired   = 10203 // 请输入角色名称
    CodeQueryFailed    = 60201 // 系统繁忙，请稍后重试
)
```

MM=00 通用码（`10001`–`10009`、`20001`、`20002`、`30001`、`60001`）放在
`internal/i18n/code.go`——它们的使用方是中间件、参数校验与 handler，
**业务 service 不应引用**。

**核心规则：每个模块只使用本模块 MM 段的错误码，不引用其它模块的。**
错误码的编号规则见 [i18n/CLAUDE.md](../i18n/CLAUDE.md)。

由此产生的一条推论，别去「优化」它：

> **不同模块出现相同文案是可以接受的。** 例如多个模块都有「系统繁忙，请稍后重试」，
> 但它们是各自的码。不要为了消除文案重复而让两个模块共用一个错误码——
> 那样做了，错误码就不再能标识是哪个模块出的问题。

**已知限制：`dto/input` 里的 struct tag 无法使用常量。**

```go
Enable *bool `json:"enable" binding:"required" err:"required=11001"`
```

Go 的 struct tag 只能是字符串字面量，没有机制引用常量，所以这里仍写数字。
改这些 tag 时请到对应模块的 `code.go` 里核对含义。

新增错误码：在 YAML 的 `error` 节点下添加（中英两份都要），
再在对应 `code.go` 里加常量。**漏加 YAML 不会编译失败**，
只会在界面上渲染成 `unknown error`。

## 依赖注入

```go
type Service struct {
    adminRepo     admin.Repository
    adminRoleRepo admin_role.Repository
    roleRepo      role.Repository
    db            *gorm.DB
    rdb           *redis.Client
}

func New(adminRepo admin.Repository, adminRoleRepo admin_role.Repository, roleRepo role.Repository, db *gorm.DB, rdb *redis.Client) *Service {
    return &Service{ /* ... */ }
}
```

- `New` 一律返回 `*Service`（`bootstrap.Services` 里存的也是指针，
  因为 `live.Service` 内含 `sync.Mutex`，值类型会被复制）
- **只注入当前模块真正需要的依赖**，避免「顺手全注入」引入不必要依赖
- 跨服务调用目前只有 `live` → `liveuser` / `robotconfig`、
  `order` → `liveuser`。新增跨服务依赖前先想清楚能不能改成共享逻辑

## 事务与 Redis

- 事务用 `s.db.Transaction(func(tx *gorm.DB) error { ... })`，
  `tx` 一路透传给 Repository 的第二个参数
- **Redis 的操作放在事务之外**：事务回滚时 Redis 无法回滚，
  放在里面会造成两边不一致。见 `admin.Save` 结尾的 `Logout`、`role.RemoveRoleUsers`

## 模块注册

新增 Service 后，在 `internal/bootstrap/service.go` 中注册：

```go
type Services struct {
    Role *role.Service
    // ...
}

func InitServices(repo *Repositories, db *gorm.DB, rdb *redis.Client, /* ... */) *Services {
    return &Services{
        Role: role.New(repo.Role, repo.Admin, repo.RoleMenu, repo.AdminRole, repo.Menu, db, rdb),
        // ...
    }
}
```
