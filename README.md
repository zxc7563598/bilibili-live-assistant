# GoAdminKit

开箱即用的全栈后台管理系统，基于 **Go + Vue 3** 构建，将 API 服务与前端界面整合为​**单一二进制部署**。

内置 RBAC 权限控制与完整后台基础能力，分层清晰、扩展友好，适用于快速搭建各类管理系统与控制面板。

---

## 在线示例

在线体验：[https://gak.hejunjie.life/admin](https://gak.hejunjie.life/admin)

> 建议先体验 Demo，再阅读下方说明，会更直观。

---

## 文档入口

- 部署 / 开发文档（推荐）：[https://hejunjie.life/gak](https://hejunjie.life/gak)
- 在线项目说明：[https://zread.ai/zxc7563598/GoAdminKit](https://zread.ai/zxc7563598/GoAdminKit)

---

## 项目介绍

GoAdminKit 是一个为“​**快速落地后台系统**”而设计的全栈解决方案：

- 后端提供完整的 REST API 与鉴权能力
- 前端提供开箱即用的管理后台界面
- 构建阶段自动将前端资源嵌入后端

最终部署时，你只需要：

> **一个二进制文件，即可同时运行 API + 管理后台**

---

## 项目特点

- ​**单一二进制部署**：无需单独部署前端，极大简化上线流程
- ​**完整 RBAC 权限系统**：管理员 / 角色 / 菜单 / 按钮权限控制
- ​**JWT 鉴权机制**：支持 access / refresh token，Redis 可选增强
- ​**统一错误码体系**：对外仅暴露标准错误码，支持多语言扩展
- ​**后台基础能力齐全**：登录 / 登出 / 权限分配 / 菜单管理全打通
- ​**内置 API 文档**：开发环境集成 Swagger + ReDoc
- ​**清晰分层结构**：低耦合设计，方便扩展业务
- ​**优秀前端开发体验**：通用 CRUD 组件 + composables 减少重复代码

---

## 项目架构

### 后端分层（核心链路）

> DTO → Handler → Service → Repository → Model

依赖统一在 `internal/bootstrap` 中完成装配。

- ​`internal/dto/input`：请求参数结构
- ​`internal/dto/resp`：响应结构
- ​`internal/handler`：HTTP 入口（参数解析 / 返回处理）
- ​`internal/service`：业务逻辑编排（不直接操作数据库）
- ​`internal/repository`：数据访问层（CRUD 封装）
- ​`internal/model`：GORM 模型定义

---

### 前端模块

标准后台管理结构，包含：

- ​`src/views`：页面（登录 / 用户 / 角色 / 菜单等）
- ​`src/components/me`：通用 CRUD 页面组件
- ​`src/composables`：组合式封装（表单 / 弹窗 / CRUD / 缓存）
- 路由权限控制：基于路由守卫 + 权限状态

---

## 技术栈

**后端：**

- Go 1.25+
- Gin
- GORM（MySQL / PostgreSQL）
- JWT
- Redis（可选）
- Swagger / ReDoc
- zap + lumberjack（日志）

**前端：**

- Vue 3 + Vite
- Naive UI
- Pinia（含持久化）
- UnoCSS
- axios

---

## 快速开始

### 环境要求

- Go ≥ 1.25
- Node.js ≥ 18（仅开发构建前端需要）
- MySQL 或 PostgreSQL
- Redis（可选）

---

### 克隆并初始化配置

```bash
git clone https://github.com/zxc7563598/GoAdminKit.git
cd GoAdminKit
cp config.example.yaml config.yaml
```

修改 `config.yaml`（数据库、JWT 密钥等）。

---

### 启动开发环境

```bash
make dev
```

访问地址：

- API：`http://localhost:9000/api/admin`
- Swagger：`http://localhost:9000/swagger/index.html`
- 后台：`http://localhost:9000/admin/`

---

## 部署方式

推荐：**本地构建 → 上传服务器运行**

```bash
make build
GIN_MODE=release ./bin/GoAdminKit -config ./config.yaml -port 9000
```

支持通过 Nginx 做反向代理统一入口（详见文档）。

---
