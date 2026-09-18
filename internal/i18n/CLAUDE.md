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

| 编号 | 模块 | 常量文件 |
|------|------|---------|
| 00 | 通用模块（Common） | `internal/i18n/code.go` |
| 01 | 管理员模块（Admin） | `internal/service/admin/code.go` |
| 02 | 角色模块（Role） | `internal/service/role/code.go` |
| 03 | 菜单模块（Menu） | `internal/service/menu/code.go` |
| 04 | 直播模块（Live） | `internal/service/live/code.go` |
| 05 | 机器人配置模块（RobotConfig） | `internal/service/robotconfig/code.go` |
| 06 | 弹幕模块（LiveDanmu） | `internal/service/livedanmu/code.go` |
| 07 | 礼物模块（LiveGift） | `internal/service/livegift/code.go` |
| 08 | 用户模块（LiveUser） | `internal/service/liveuser/code.go` |
| 09 | App 配置模块（AppConfig） | `internal/service/appconfig/code.go` |
| 10 | 商品模块（Product） | `internal/service/product/code.go` |
| 11 | 订单模块（Order） | `internal/service/order/code.go` |
| 12 | 反馈模块（Feedback） | `internal/service/feedback/code.go` |
| 13 | 收货地址模块（Address） | `internal/service/address/code.go` |
| 14 | 上传模块（Upload） | `internal/service/upload/code.go` |
| 15 | PK 对战记录模块（LivePk） | `internal/service/livepk/code.go` |
| 16 | 验证码模块（Altcha） | `internal/service/altcha/code.go` |
| 17–99 | 未分配 | — |

新增业务模块时，分配新的 MM 编号并在此文档更新。

## 错误码的归属规则

**每个模块只使用本模块 MM 段的错误码，不引用其它模块的。**

MM=00 是唯一例外：它是「通用模块」，中间件、参数校验与 handler 共用这一段
（例如 20001 登录已过期由 12 个 handler 包共同返回）。**业务 service 不应引用 MM=00**，
自己的登录态、参数错误等都要用本模块的码。

由此产生一条推论，别去「优化」它：

> **不同模块出现相同文案是可以接受的。** 多个模块都有「系统繁忙，请稍后重试」、
> 「请求参数不合法」，但它们是各自的码。不要为了消除文案重复而让两个模块共用一个
> 错误码——那样错误码就不再能标识是哪个模块出的问题，按模块分组的监控告警也会失效。

代码里的引用一律用常量，不写五位数裸字面量：常量声明在各模块的 `code.go`，
MM=00 的在 `internal/i18n/code.go`。

**已知限制**：`internal/dto/input/*.go` 的 `err:"required=11001"` 这类 struct tag
只能是字符串字面量，Go 没有机制引用常量，所以那里仍写数字。

## 核心约定

1. 所有用户可见文本通过 i18n 获取，不在代码中写死
2. 错误信息统一通过错误码 + `i18n.E()` 返回
3. 新增错误码在 YAML 的 `error` 节点下添加，**中英两份都要加**。漏加不会编译失败，
   只会在界面上渲染成 `unknown error`（`i18n.go` 的兜底文案），且没有任何自动化校验
4. 同一错误类型+模块下，XX 从 01 递增且不复用
5. 优先使用业务错误（4xx），避免滥用参数错误（1xx）。
   注意现有的 `10002`–`10008` 是登录态失效（认证错误）却编在 T=1 下，属历史遗留；
   新模块的登录态错误请用 T=2（如 admin 的 `20101`–`20103`）
6. 系统错误（6xx）不暴露内部细节，统一返回通用提示
