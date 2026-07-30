# CLAUDE.md — DTO 层

## 目录结构

```
internal/dto/
├── input/    # 请求参数结构体
└── resp/     # 响应数据结构体
```

每个 Handler 模块对应一组 DTO 文件（如 `admin.go`、`role.go`），便于快速定位。

## Input DTO（请求参数）

定义在 `internal/dto/input/`，用于 Handler 层接收和校验请求参数。

### 标准写法

```go
type AdminListPageReq struct {
    PageNo   int     `json:"pageNo" binding:"required" err:"required=10101" example:"1"`
    PageSize int     `json:"pageSize" binding:"required" err:"required=10101" example:"20"`
    Username *string `json:"username" example:"admin"`
    Gender   *int    `json:"gender" example:"1" enums:"0,1,2"`
    Enable   *int    `json:"enable" example:"1" enums:"0,1"`
}

type AdminChangePasswordReq struct {
    OldPassword string `json:"oldPassword" binding:"required,min=6,max=32" err:"required=10103,min=10104,max=10105" example:"123456"`
    NewPassword string `json:"newPassword" binding:"required,min=6,max=32" err:"required=10103,min=10104,max=10105" example:"654321"`
}
```

### Tag 说明

| Tag | 说明 |
|-----|------|
| `json` | JSON 字段名 |
| `binding` | 校验规则（required, min, max, email 等），多个规则用逗号分隔 |
| `err` | 自定义错误码映射，格式 `rule=code`，多个用逗号分隔 |
| `example` | Swagger 示例值 |
| `enums` | 可选值列表（Swagger 文档用） |
| `comment` | 字段注释（Swagger 描述） |

### 可选字段处理

可选字段使用**指针类型**（`*string`、`*int`），这样可以区分"未传"和"传了零值"。

### 校验错误码示例

```go
binding:"required,min=6,max=32" err:"required=10103,min=10104,max=10105"
```

- `required` 校验失败 → 错误码 `10103`
- `min` 校验失败 → 错误码 `10104`
- `max` 校验失败 → 错误码 `10105`

当需要使用错误码时，可以阅读[多语言错误码设计](internal/i18n/CLAUDE.md)，遵循其**错误码设计**，优先复用符合情况的错误码，无符合情况的错误码时，根据设计规定新增错误码并同步所有语言文件

## Resp DTO（响应结构）

定义在 `internal/dto/resp/`，用于 Swagger 文档生成和 Handler 响应封装。

```go
type AdminListPageResp struct {
    Total    int64            `json:"total"`
    PageData []AdminListItem  `json:"pageData"`
}

type AdminListItem struct {
    ID       uint64 `json:"id"`
    Username string `json:"username"`
    Status   int    `json:"status"`
}
```

## 与 Service DTO 的关系

Service 层有自己的 `dto.go`，定义内部入参与出参。Handler 的 `common.go` 负责在 Service DTO 与 Handler DTO 之间做转换：

```
Request (input) → Handler 接收
    ↓
Handler 将 input 转为 Service DTO 或直接传值
    ↓
Service 处理，返回 Service DTO
    ↓
Handler common.go 将 Service DTO 转为 resp 结构
    ↓
response.Success(c, lang, resp)
```

这样保持了层与层之间的解耦，避免 Service 层的内部结构暴露到接口层。
