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
├── common.go    # 数据转换方法（Service 返回结构 → resp 响应结构）
└── handler.go   # HTTP 接口处理逻辑
```

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
- 错误码设计参考 [多语言错误码设计](internal/i18n/CLAUDE.md)，遵循其**错误码设计**，优先复用符合情况的错误码，无符合情况的错误码时，根据设计规定新增错误码并同步所有语言文件

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
    adminSvc admin.Service
}

func New(adminSvc admin.Service) *Handler {
    return &Handler{adminSvc: adminSvc}
}
```

新增接口后，在 `internal/bootstrap/router.go` 注册路由。

## Swagger 注释

每个接口方法必须编写标准 Swagger 注释（描述、参数、响应结构），用于自动生成 API 文档。
