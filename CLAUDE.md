# CLAUDE.md — BiliLive Assistant 项目总览

## 项目简介

BiliLive Assistant 是一个前后端一体的后台管理系统开发框架：
- **后端**：Go + Gin + GORM，采用 DDD 分层架构
- **前端**：Vue 3 + Naive UI + Vite，基于 vue-naive-admin 二次开发
- **数据库**：MySQL / PostgreSQL（任选其一）
- **缓存**：Redis（可选，用于 JWT 会话控制）

## 环境要求

| 依赖 | 版本 | 用途 |
|------|------|------|
| Go | ≥ 1.25 | 后端编译与运行 |
| Node.js | ≥ 18 | 前端构建工具链（Vite），仅开发环境需要 |
| MySQL 或 PostgreSQL | 主流版本 | 数据库 |
| Redis | ≥ 6.0（可选） | JWT 认证令牌会话缓存 |

## 常用命令

```bash
# 开发环境（同时启动前后端，前端支持热更新）
make dev

# 仅启动 Go 后端
make dev-go

# 仅启动前端 dev server
make dev-web

# 构建项目（前端 + 后端，产出 ./bin/gak）
make build

# 仅构建前端资源 → internal/webui/dist
make build-web

# 跨平台构建（Linux/macOS/Windows）
make release

# 生成 Swagger 文档
make swagger

# 清理构建产物
make clean
```

## 项目结构概览

```
.
├── cmd/server/main.go     # 应用入口
├── internal/              # 后端核心代码（DDD 分层）
│   ├── bootstrap/         # 依赖注入与初始化
│   ├── config/            # 配置管理
│   ├── dto/               # 数据传输对象（input / resp）
│   ├── enum/              # 枚举定义
│   ├── handler/           # HTTP 接口层
│   ├── i18n/              # 国际化
│   ├── logger/            # 日志模块
│   ├── middleware/        # 中间件
│   ├── migrate/           # 数据库迁移与初始化
│   ├── model/             # 数据模型（GORM 映射）
│   ├── repository/        # 数据访问层
│   ├── response/          # 统一响应封装
│   ├── service/           # 业务逻辑层
│   ├── validation/        # 参数校验
│   └── webui/dist/        # 前端构建产物（嵌入二进制）
├── pkg/                   # 公共工具包
├── web/                   # 前端源码
├── config.example.yaml    # 配置示例
├── Makefile               # 构建脚本
└── docs/                  # API 文档
```

## 开发流程

### 初始化项目

```bash
git clone https://github.com/zxc7563598/bilibili-live-assistant ./oneadmin
cd oneadmin
cp config.example.yaml config.yaml
# 根据实际情况修改 config.yaml（数据库连接等）
make dev
```

### 可访问服务

| 服务 | 地址 | 说明 |
|------|------|------|
| Go API | http://localhost:25443 | 后端 REST API |
| Vite Dev | http://localhost:3200 | 前端开发服务器（热重载，代理 API 到 :25443） |
| Swagger UI | http://localhost:25443/swagger/index.html | 交互式 API 文档（开发环境） |
| 管理后台 | http://localhost:25443/admin/ | 前端页面（嵌入后端服务） |

## 生产部署

### 构建

```bash
make build
# 产物：./bin/BiliLiveAssistant（已包含前端资源，无需额外部署）
```

### 服务器部署

推荐目录结构：
```
/opt/项目名称/
├── BiliLiveAssistant   # 可执行文件
├── config.yaml  # 配置文件
└── logs/        # 日志目录
```

启动：
```bash
GIN_MODE=release ./BiliLiveAssistant -config ./config.yaml -port 25443
```

推荐使用 systemd 管理服务。

## 分层 CLAUDE.md 索引

本项目采用分层 CLAUDE.md 策略，各层级有独立的开发指南：

- [web/CLAUDE.md](web/CLAUDE.md) — 前端 Vue 页面开发
- [internal/CLAUDE.md](internal/CLAUDE.md) — 后端架构总览与开发流程
  - [internal/model/CLAUDE.md](internal/model/CLAUDE.md) — Model 层规范
  - [internal/repository/CLAUDE.md](internal/repository/CLAUDE.md) — Repository 层规范
  - [internal/service/CLAUDE.md](internal/service/CLAUDE.md) — Service 层规范
  - [internal/handler/CLAUDE.md](internal/handler/CLAUDE.md) — Handler 层规范
  - [internal/dto/CLAUDE.md](internal/dto/CLAUDE.md) — DTO 结构规范
  - [internal/enum/CLAUDE.md](internal/enum/CLAUDE.md) — 枚举规范
  - [internal/i18n/CLAUDE.md](internal/i18n/CLAUDE.md) — 多语言与错误码
  - [internal/bootstrap/CLAUDE.md](internal/bootstrap/CLAUDE.md) — 依赖注入
