# CLAUDE.md — Handler 层

## 职责边界

Handler 层是 HTTP 请求的入口，负责：

1. 接收请求、解析参数
2. 调用 Service 层
3. 将结果转换为统一响应格式返回
4. 统一处理错误（记录日志 + 返回错误码）

Handler **只依赖 Service**，不直接依赖 Repository 或数据库。

## 文件组织

```
internal/handler/<模块名>/
├── common.go    # Handler 结构体、New、包级常量/变量、数据转换方法
├── shop.go      # /api/shop 接口方法（该端有接口时才存在）
└── admin.go     # /api/admin 接口方法（该端有接口时才存在）
```

**按端分文件**，一眼能看出某个方法是给商城端还是管理端用的。约定：

- `Handler` 结构体与 `New()` 一律放 `common.go`，`shop.go` / `admin.go` 只放接口方法
- 一个包若两端都有接口，两个文件都在；只有一端就只有一个
- 路由既不属于 `/api/shop` 也不属于 `/api/admin` 的包（如 `altcha` 的公开验证码接口），
  接口方法也放 `common.go`
- **不再有 `handler.go`**

## 方法命名

承接 [service/CLAUDE.md](../service/CLAUDE.md) 的词汇：接口方法名尽量与它调用的
Service 方法一致，两边含义相同时直接沿用同一个词。

**与 Service 层相同的规则**：

- 分页查询叫 `ListPage`；同一模块有多个分页接口时加限定词区分
  （`order.ListPageByUser`、`livegift.BlindBoxListPage`）
- 读方法有动词：`Get*` 返回单条/整体，`List*` 返回集合。**不要用名词当方法名**
  （`List()` → `GetMenuTree()`、`Buttons()` → `ListMenuButtons()`、
  `Permissions()` → `GetPermissions()`）
- 树形结果用 `Get*` 而非 `List*`：`GetMenuTree`
- 名字要能看出来有副作用：会发网络请求的用 `Fetch*`（`FetchRoomGroups`）、
  会翻转状态的用 `Toggle*`（`Toggle`）、持续推送的用 `Stream*`（`StreamMessages`）
- 谓词用 `XxxExists`（`ExistsAccount`）；要把结果当数据下发给前端时用动词开头
  （`CheckMenuExists` 返回 `{has: bool}`）

**唯一的例外是 `Details`。** 读单条时 `admin` / `order` / `liveuser` / `feedback`
都用 `Details`（`product` 因为两端同名而写作 `ShopDetails` / `AdminDetails`），
按上面的规则该叫 `GetDetail`；但 Service 层用的是同一个名字，两层保持一致比
单层内部自洽更重要，所以保留。这是全项目唯一的「名词当方法名」。

**写操作**用动词：`Save` / `Delete` / `Update<对象>` / `Apply<配置>` / `Submit` /
`Login` / `Logout`。`Save` 留给「有 id 则更新、无 id 则新增」的接口
（`Save`、`AdminSave`、`SaveAddress`），只改一个对象的用 `Update*`
（`UpdateRoom`、`UpdateShipStatus`、`AdminUpdateEnable`）。

**双端前缀**：只有同一包内商城端与管理端方法**同名冲突**时才加 `Shop` / `Admin`
前缀。全项目只有 `product` 一个包需要——它两端都有 list 与 details，其余双端包靠
`ListPageByUser` ↔ `ListPage` 这类词本身就能区分，与 Service 层的做法一致。

**包内小工具**：转换函数用 `toXxx` 开头（`toAdminListItems`），构造函数统一叫
`New`（不是 `NewHandler`）。

## 标准接口处理流程

每个接口方法遵循统一的四步流程：

### 1. 获取上下文信息

```go
ctx := c.Request.Context()
lang := i18n.GetLang(ctx)
adminInfo, ok := handler.GetAdminInfo(c)
if !ok {
    response.Error(c, lang, 20001)
    return
}
```

### 2. 解析并校验请求参数

```go
var req input.AdminListPageReq
if code, ok, err := handler.BindAndValidate(c, &req); !ok {
    handler.ErrorLog(logger.AdminLogger, "参数异常", code, err)
    response.Error(c, lang, code)
    return
}
```

参数结构定义在 `internal/dto/input/`，使用 `binding` + `err` 标签完成校验。

### 3. 调用 Service 层

```go
svcResp, errCode, err := h.adminSvc.Login(ctx, req.Username, req.Password, req.Captcha)
if errCode != 0 {
    handler.ErrorLog(logger.AdminLogger, "adminSvc.Login 调用失败", errCode, err,
        zap.String("uname", req.Username),
    )
    response.Error(c, lang, errCode)
    return
}
```

错误处理要点：
- `errCode != 0` 判定业务失败
- 用 `handler.ErrorLog()` 记录原始错误（含上下文信息）
- 用 `response.Error(c, lang, errCode)` 返回给用户（只暴露错误码，不暴露原始错误）

### 4. 转换并返回响应

```go
response.Success(c, lang, resp.AdminListPageResp{
    Total:    svcResp.Total,
    PageData: toAdminListItems(svcResp.PageData),
})
```

数据转换方法（`toXxx`）放在 `common.go` 中。

## 错误处理机制

### 错误流转

```
Repository 抛 error → Service 透传 → Handler 统一处理
```

- Handler 用 `handler.ErrorLog()` 记录完整错误到日志
- 用户只看到错误码映射后的多语言信息
- **绝不直接将原始 error 返回给用户**（防止泄露数据库结构等内部信息）
- 错误码设计参考 [多语言错误码设计](../i18n/CLAUDE.md)，遵循其**错误码设计**，优先复用符合情况的错误码，无符合情况的错误码时，根据设计规定新增错误码并同步所有语言文件

## 统一日志

每个 Handler 模块维护独立的 `*zap.Logger`：

```go
handler.ErrorLog(
    logger.AdminLogger,        // 模块日志实例
    "adminSvc.Login 调用失败",  // 日志描述
    errCode,                   // 错误码
    err,                       // 原始错误
    zap.String("uname", req.Username),  // 附加字段
)
```

## 依赖注入与注册

```go
// internal/bootstrap/handler.go
type Handlers struct {
    Admin *admin.Handler
}

func InitHandlers(svc *Services) *Handlers {
    return &Handlers{
        Admin: admin.New(&svc.Admin),
    }
}
```

```go
// handler 结构体
type Handler struct {
    adminSvc *admin.Service
}

func New(adminSvc *admin.Service) *Handler {
    return &Handler{adminSvc: adminSvc}
}
```

新增接口后，在 `internal/bootstrap/router.go` 注册路由。

## Swagger 注释

每个接口方法必须编写标准 Swagger 注释（描述、参数、响应结构），用于自动生成 API 文档。
