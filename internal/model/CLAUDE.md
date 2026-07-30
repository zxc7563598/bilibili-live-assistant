# CLAUDE.md — Model 层

## 职责边界

Model 层位于最底层，**只做两件事**：

1. 定义数据库表结构的 GORM 映射
2. 定义 Repository 查询使用的结构体

**严禁**在 Model 层：
- 编写业务逻辑
- 做类型转换
- 定义接口返回结构（那是 DTO 的事）

## 基础模型（BaseModel）

所有表结构统一嵌入 `BaseModel`，自动处理时间戳和软删除：

```go
type BaseModel struct {
    CreatedAt int64          `gorm:"not null;comment:创建时间"`
    UpdatedAt int64          `gorm:"not null;comment:更新时间"`
    DeletedAt gorm.DeletedAt `gorm:"index;comment:软删除"`
}
```

`BaseModel` 通过 GORM Hook 自动维护 `CreatedAt` / `UpdatedAt`：
- `BeforeCreate`：自动写入当前时间戳
- `BeforeUpdate`：自动更新 `UpdatedAt`

## 表模型定义规范

```go
type AdminRole struct {
    ID      uint64 `gorm:"primaryKey"`
    AdminID uint64 `gorm:"not null;default:0;comment:管理员ID"`
    RoleID  uint64 `gorm:"not null;default:0;comment:角色ID"`
    BaseModel
}

func (AdminRole) TableName() string {
    return "admin_roles"
}
```

约定：
- **显式声明 `TableName()`**，避免 GORM 默认命名不一致
- 字段尽量带 `gorm tag`（约束 + 注释）
- 所有表统一嵌入 `BaseModel`
- **一个 model 文件对应一张数据库表**

## 查询结构体（非表结构）

当只需要部分字段时，可以在 Model 文件中定义查询结构体：

```go
// AdminRoleListItem 不对应数据库表，仅用于查询结果接收
type AdminRoleListItem struct {
    ID      uint64
    AdminID uint64
    RoleID  uint64
}
```

适用于表字段很多但只需查部分字段的场景，可提升查询性能。

## 多表查询约定

多表关联查询的结果结构体，应定义在**主表的 Model 文件**中，保证归属清晰。

## 数据迁移

新增 Model 后，在 `internal/migrate/migrate.go` 中注册实现**自动建表**；如需初始数据，在 `internal/migrate/seed.go` 中编写。
