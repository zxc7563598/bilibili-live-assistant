# CLAUDE.md — Repository 层

## 职责边界

Repository 层是**唯一可以操作数据库的层**。核心职责：

1. 封装对 Model 的 CRUD 操作
2. 提供标准数据访问接口给 Service 层
3. 每个 Repository 对应一个 Model；多表查询放在主表 Repository

**严禁**：Service 或 Handler 层直接操作数据库字段。所有数据操作必须通过 Repository。

## 文件组织

每个 Repository 模块包含两个文件：

### `interface.go` — 接口定义

```go
type Repository interface {
    base.Repository[model.Role]  // 继承泛型基础方法
    GetByCode(ctx context.Context, tx *gorm.DB, code string) (*model.Role, error)
    ListEnabled(ctx context.Context, tx *gorm.DB) ([]model.Role, error)
    ListPage(ctx context.Context, tx *gorm.DB, query model.RoleListPageQuery) ([]model.RoleListItem, int64, error)
}
```

### `gorm_repo.go` — 实现 + 依赖注入

```go
type gormRepo struct {
    db *gorm.DB
    base.Repository[model.Role]
}

func New(db *gorm.DB) Repository {
    return &gormRepo{
        db:         db,
        Repository: base.New[model.Role](db),
    }
}

// getDB 封装 ctx + tx（每个方法都应该用它）
func (r *gormRepo) getDB(ctx context.Context, tx *gorm.DB) *gorm.DB {
    db := r.db
    if tx != nil {
        db = tx
    }
    if ctx != nil {
        db = db.WithContext(ctx)
    }
    return db
}
```

## Base Repository（泛型基础方法）

```go
type Repository[T any] interface {
    GetByID(ctx, tx, id uint64) (*T, error)
    GetByIDs(ctx, tx, ids []uint64) ([]T, error)
    FindAll(ctx, tx) ([]T, error)
    FindByField(ctx, tx, field string, value any) ([]T, error)
    FindOneByField(ctx, tx, field string, value any) (*T, error)
    Create(ctx, tx, entity *T) (*T, error)
    Update(ctx, tx, entity *T) error
    Delete(ctx, tx, id uint64) error
    Count(ctx, tx) (int64, error)
    Exists(ctx, tx, field string, value any) (bool, error)
}
```

嵌入 `base.Repository[model.Xxx]` 即可自动获得这些方法，无需重复实现。

## 方法编写规范

### 分页查询示例

```go
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.RoleListPageQuery) ([]model.RoleListItem, int64, error) {
    var list []model.RoleListItem
    var total int64
    db := r.getDB(ctx, tx)
    db = db.Model(&model.Role{})
    if v := query.Name; v != nil && *v != "" {
        db = db.Where("name LIKE ?", "%"+*v+"%")
    }
    if v := query.Enable; v != nil {
        e := enum.Enable(*v)
        if e.IsValid() {
            db = db.Where("enable = ?", e)
        }
    }
    if err := db.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    err := db.Order("id asc").Offset(query.Offset).Limit(query.Limit).Find(&list).Error
    return list, total, err
}
```

关键点：
- 使用 `r.getDB(ctx, tx)` 获取 DB 实例（自动处理 context 和事务）
- 每个方法接受 `ctx context.Context` 和 `tx *gorm.DB`（事务支持）
- 枚举值先校验 `IsValid()` 再使用
- LIKE 查询注意防注入：用参数化 `?` 占位符拼接

## 模块注册

新增 Repository 后，在 `internal/bootstrap/repository.go` 中注册：

```go
type Repositories struct {
    Role role.Repository
    // ...
}

func InitRepositories(db *gorm.DB) *Repositories {
    return &Repositories{
        Role: role.New(db),
        // ...
    }
}
```
