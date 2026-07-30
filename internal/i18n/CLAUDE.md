# CLAUDE.md — 多语言与错误码

## 概述

多语言模块（`internal/i18n/`）提供统一的国际化支持，用于：
- 错误信息的用户展示
- 枚举值的多语言展示
- 其他用户可见文本

## 语言文件

### 位置

```
internal/i18n/locales/
├── zh.yaml
└── en.yaml
```

### 文件结构

```yaml
error:
  0: 操作成功
  10001: 请求参数不合法

gender:
  female: 女
  male: 男

no: 否
yes: 是
```

解析规则：
- `error` 节点 → 错误码映射（`errorsMap`，key 为 int）
- 其他节点 → 普通文本映射（`keysMap`，key 为用 `.` 拼接的路径）

### 加载机制

- 开发环境（debug/test）：从本地文件系统加载，支持热更新（修改 YAML 无需重启）
- 生产环境（release）：从 embed 内嵌文件加载（编译时打包进二进制，无需本地文件）

### 初始化

已在 `bootstrap/app.go` 中自动调用 `i18n.InitLocales()`，无需额外配置。

## 核心 API

### 获取当前请求语言

```go
lang := i18n.GetLang(ctx)
```

语言信息由中间件（`internal/middleware/language.go`）写入 `context.Context`。

### 获取普通文本

```go
i18n.T(lang string, key string, args ...string) string
```

示例：
```go
i18n.T("en", "no")                    // → "No"
i18n.T("en", "gender.female")         // → "Female"
i18n.T("en", "key1", "key2", "key3")  // 等价于 key="key1.key2.key3"
```

### 获取错误信息

```go
i18n.E(lang string, code int) string
```

从 `errorsMap` 中根据错误码返回对应语言的错误文本。

## 在 Handler 中使用

```go
func (h *Handler) ListPage(c *gin.Context) {
    ctx := c.Request.Context()
    lang := i18n.GetLang(ctx)

    // 参数校验失败
    if code, ok, err := handler.BindAndValidate(c, &req); !ok {
        handler.ErrorLog(logger.AdminLogger, "参数异常", code, err)
        response.Error(c, lang, code)  // response 内部调用 i18n.E(lang, code)
        return
    }

    // 业务调用失败
    svcResp, errCode, err := h.adminSvc.Login(ctx, ...)
    if errCode != 0 {
        handler.ErrorLog(logger.AdminLogger, "调用失败", errCode, err)
        response.Error(c, lang, errCode)
        return
    }

    // 成功
    response.Success(c, lang, respData)
}
```

`response.Error()` 和 `response.Success()` 内部会自动处理多语言翻译。

## 错误码设计

```
T MM XX
```

| 位 | 范围 | 含义 |
|----|------|------|
| T | 1-6 | 错误类型 |
| MM | 00-99 | 模块编号 |
| XX | 01-99 | 具体错误 |

### 错误类型（T）

| T | 类型 | 说明 |
|---|------|------|
| 1 | 参数错误 | 参数缺失/类型错误/格式错误 |
| 2 | 认证错误 | 未登录/Token无效/已过期 |
| 3 | 权限错误 | 非管理员/越权访问 |
| 4 | 业务错误 | 密码错误/余额不足/状态不允许 |
| 5 | 资源错误 | 仅表示"资源不存在" |
| 6 | 系统错误 | 数据库异常/缓存异常/panic |

### 模块编号（MM）

| 编号 | 模块 |
|------|------|
| 00 | 通用模块（Common） |
| 01 | 管理员模块（Admin） |
| 02 | 角色模块（Role） |
| 03 | 菜单模块（Menu） |

新增业务模块时，分配新的 MM 编号并在此文档更新。

## 核心约定

1. 所有用户可见文本通过 i18n 获取，不在代码中写死
2. 错误信息统一通过错误码 + `i18n.E()` 返回
3. 新增错误码在 YAML 的 `error` 节点下添加
4. 同一错误类型+模块下，XX 从 01 递增且不复用
5. 优先使用业务错误（4xx），避免滥用参数错误（1xx）
6. 系统错误（6xx）不暴露内部细节，统一返回通用提示
