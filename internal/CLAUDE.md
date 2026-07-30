# CLAUDE.md — 后端架构总览

## 架构概述

后端采用简化的 DDD 分层架构，核心链路为：

```
DTO → Handler → Service → Repository → Model
```

每一层职责明确，依赖单向流动。所有依赖注入集中在 `internal/bootstrap/` 统一管理。

## 目录结构与分层职责

```
internal/
├── bootstrap/       # 依赖注入、初始化、路由注册
├── config/          # 配置结构定义与加载
├── dto/             # 数据传输对象
│   ├── input/       # 请求参数结构体
│   └── resp/        # 响应数据结构体
├── enum/            # 枚举定义（状态、类型等）
├── handler/         # HTTP 接口层（请求处理、参数解析、响应封装）
├── i18n/            # 国际化（多语言错误信息）
├── logger/          # 日志模块（zap + lumberjack）
├── middleware/       # 中间件（认证、语言、CORS 等）
├── migrate/         # 数据库迁移与初始数据
├── model/           # 数据模型（GORM 映射，仅结构定义）
├── repository/      # 数据访问层（封装数据库操作）
├── response/        # 统一响应格式
├── service/         # 业务逻辑层（核心编排）
├── validation/      # 参数校验
└── webui/dist/      # 前端构建产物（生产环境嵌入）
```

## 依赖注入流程

在 `bootstrap/app.go` 中自下而上装配：

```go
// repository
repos := InitRepositories(db)

// service
services := InitServices(repos, db, rdb)

// handler
handlers := InitHandlers(services)

// 注册路由
r := gin.New()
r = RouteRegister(r, rdb, handlers)
```

关键原则：**每个模块只注入真正需要的依赖，避免"顺手全注入"。**

## 标准开发流程

开发一个新接口的推荐步骤：

### 1. 整理需求
明确输入参数、输出结构、是否涉及状态变更。

### 2. 评估数据层（Model + Repository）
- 数据库是否已有对应表？→ 没有则新增 Model（`internal/model/`）
- 现有 Repository 方法是否满足查询？→ 不够则在对应 Repository 新增方法
- 新增 Model 需要在 `internal/migrate/migrate.go` 注册自动迁移
- 详细规范见：[internal/model/CLAUDE.md](model/CLAUDE.md)、[internal/repository/CLAUDE.md](repository/CLAUDE.md)

### 3. 定义 DTO（`internal/dto/`）
- 在 `dto/input/` 中定义请求结构（含 binding 校验标签）
- 在 `dto/resp/` 中定义响应结构（含 Swagger 注释）
- 详细规范见：[internal/dto/CLAUDE.md](dto/CLAUDE.md)

### 4. 实现业务逻辑（Service）
- 创建/使用对应 Service 模块
- 在 `dto.go` 中定义 Service 入参与出参（与 Handler DTO 解耦）
- 复杂逻辑拆分为多个方法（`common.go`），Service 只做流程编排
- 在 `bootstrap/service.go` 中注册 Service
- 详细规范见：[internal/service/CLAUDE.md](service/CLAUDE.md)

### 5. 实现接口层（Handler）
- 在对应 Handler 模块创建方法
- 参数绑定 → 调用 Service → 转换响应 → 统一返回
- 在 `bootstrap/router.go` 中注册路由
- 编写 Swagger 注释
- 详细规范见：[internal/handler/CLAUDE.md](handler/CLAUDE.md)

### 6. 注册依赖注入
在 `bootstrap/` 对应文件中注册：
- `repository.go` → 注册 Repository
- `service.go` → 注册 Service
- `handler.go` → 注册 Handler

### 开发口诀

> 先想清楚数据，再设计接口，最后实现业务。

## 分层 CLAUDE.md 索引

| 层级 | 文件 | 核心关注点 |
|------|------|-----------|
| 数据模型 | [model/CLAUDE.md](model/CLAUDE.md) | GORM 映射、BaseModel、查询结构体 |
| 数据访问 | [repository/CLAUDE.md](repository/CLAUDE.md) | 接口定义、Base Repository、事务支持 |
| 业务逻辑 | [service/CLAUDE.md](service/CLAUDE.md) | 流程编排、DTO 定义、依赖注入 |
| 接口层 | [handler/CLAUDE.md](handler/CLAUDE.md) | 参数绑定、响应封装、Swagger、错误处理 |
| 数据传输 | [dto/CLAUDE.md](dto/CLAUDE.md) | input/resp 结构规范、校验标签 |
| 枚举 | [enum/CLAUDE.md](enum/CLAUDE.md) | Key/Text/IsValid 方法、避免硬编码 |
| 多语言 | [i18n/CLAUDE.md](i18n/CLAUDE.md) | 错误码映射、多语言文本获取 |
| 依赖注入 | [bootstrap/CLAUDE.md](bootstrap/CLAUDE.md) | 初始化流程、模块注册、路由配置 |
