# CLAUDE.md — Repository 层

## 职责边界

Repository 层是**唯一可以操作数据库的层**。核心职责：

1. 封装对 Model 的 CRUD 操作
2. 提供标准数据访问接口给 Service 层
3. 每个 Repository 对应一个 Model；多表查询放在主表 Repository

**严禁**：Service 或 Handler 层直接操作数据库字段。所有数据操作必须通过 Repository。

## 文件组织

每个 Repository 模块包含两个文件：

### `interface.go` — 接口定义 + 自定义方法实现

接口和它的**自定义方法实现**都写在这个文件里：上方是接口块（声明契约），
下方是接口各方法的实现。泛型通用方法在 `base` 里，不要在这里重复实现。

```go
type Repository interface {
    base.Repository[model.Role]  // 继承泛型基础方法
    GetByCode(ctx context.Context, tx *gorm.DB, code string) (*model.Role, error)
    ListEnabled(ctx context.Context, tx *gorm.DB) ([]model.Role, error)
    ListPage(ctx context.Context, tx *gorm.DB, query model.RoleListPageQuery) ([]model.RoleListItem, int64, error)
}

// GetByCode 根据角色编码查询单条记录，不存在返回 nil
func (r *gormRepo) GetByCode(ctx context.Context, tx *gorm.DB, code string) (*model.Role, error) {
    return r.GetByField(ctx, tx, "code", code)
}
```

方法注释写两处：接口块里写契约（边界条件、时间区间开闭、返回 nil 的含义等），
实现上方写一句简述即可。

### `gorm_repo.go` — 结构体与依赖注入

只放结构体定义与构造函数，**不在这里写业务方法**。嵌入 `base.Repo[model.X]`
即可同时获得 base 的全部通用方法与 `ResolveDB`，不必再自己复制一份 `*gorm.DB`
的解析逻辑：

```go
type gormRepo struct {
    *base.Repo[model.Role]
}

func New(db *gorm.DB) Repository {
    return &gormRepo{Repo: base.NewRepo[model.Role](db)}
}
```

## Base Repository（泛型基础方法）

```go
type Repository[T any] interface {
    GetByID(ctx, tx, id int64) (*T, error)
    GetByIDs(ctx, tx, ids []int64) ([]T, error)
    ListAll(ctx, tx) ([]T, error)
    ListByField(ctx, tx, field string, value any) ([]T, error)
    GetByField(ctx, tx, field string, value any) (*T, error)
    Create(ctx, tx, entity *T) (*T, error)
    CreateBatch(ctx, tx, entities []T) error
    Save(ctx, tx, entity *T) error
    UpdateMap(ctx, tx, field string, value any, updates map[string]any) error
    UpdateField(ctx, tx, id int64, field string, value any) error
    Delete(ctx, tx, id int64) error
    DeleteByIDs(ctx, tx, ids []int64) error
    Count(ctx, tx) (int64, error)
    Exists(ctx, tx, field string, value any) (bool, error)
    AdjustField(ctx, tx, id int64, field string, delta int64) error
}
```

嵌入 `base.Repository[model.Xxx]` 即可自动获得这些方法，无需重复实现。

几处容易踩的点：

- `GetByID` / `GetByField` 查不到时返回 `(nil, nil)`，不是 error
- `Save` 走的是 `db.Save`：主键为零时插入，否则按主键**整行覆盖**，不是部分更新
- `UpdateMap` 按条件更新，命中多行就全部更新
- `Delete` / `DeleteByIDs` 是**软删除**（模型内嵌 `BaseModel` 的 `DeletedAt`）
- `AdjustField` 的 `delta` 可为负；没命中任何行时返回 `gorm.ErrRecordNotFound`

各模块仓库用 `r.ResolveDB(ctx, tx)` 取数据库句柄（事务优先，其次附加 context）。

## 方法命名约定

新增方法时照着这里选名字，让「名字 + 签名」就足以看懂行为。

**读取**

| 前缀 | 返回 | 例子 |
|------|------|------|
| `Get*` | 单条或标量 | `GetByUID`、`GetByField`、`GetActiveByRoomUID` |
| `List*` | 集合 | `ListPage`、`ListByUID`、`ListEnabled`、`ListExpiredMuted` |
| `Count*` / `Sum*` / `Distinct*` | 聚合结果，不返回记录行 | `CountByUID`、`SumNumAndAmount`、`DistinctRoomIDs` |

分页查询统一叫 `ListPage`，返回 `(列表, 总数, error)`。

**写入**

- `Create*` —— 含 upsert 语义的（`CreateIfNotExist`）也归这一族
- `Update*` —— 部分更新；只更新单列的用 `UpdateField` 或 `Update<列名>ByID`
- `Save` —— 整行覆盖
- 单行操作加 `ByID` 后缀：`UpdateNameByID`、`ToggleEnableByID`
- 按外键批量操作加 `By<外键>` 后缀：`DeleteByProductID`。**不要用 `ByID`**，
  那个后缀读起来像单行操作，而它实际会命中多行
- 接受 `delta` 的方法用 `Adjust*`，不承诺方向：`AdjustCredit`、`AdjustStock`
- 只有带方向校验的才用 `Increment*` / `Decrement*`，例如 `DecrementStock`
  会带上 `stock >= delta` 守卫
- 领域动作可以直接用领域动词：`ToggleEnableByID`（翻转而非设置）、`CancelActiveByID`
- `Delete*` 默认是软删除，注释里写明。**例外：关联表用物理删除**

### 关联表为什么必须物理删除

`role_menus`、`admin_roles` 这种纯关联表，一行只表示「A 拥有 B」这个事实，
软删除不保留任何有用信息，却会留下两个坑：

1. **占着唯一索引**：`admin_roles` 有 `uk_admin_role(admin_id, role_id)`，
   唯一索引不含 `deleted_at`，被软删除的行照样占着这个键。而绑定走的是
   「先删后插」，于是重新绑定同一组合时插入会撞唯一键，
   报 `gorm.ErrDuplicatedKey`（文案是 `duplicated key not allowed`）
2. **堆积死行**：没有唯一索引的 `role_menus` 不报错，但每次重新授权都会
   把旧行标成已删除却留在表里

所以这类表的 `Delete*` 要写 `db.Unscoped().Where(...).Delete(&model.X{})`。
写的时候注意：`Unscoped()` 只加在删除上，查询仍走默认的软删除过滤。

## 方法编写规范

### 分页查询示例

```go
func (r *gormRepo) ListPage(ctx context.Context, tx *gorm.DB, query model.RoleListPageQuery) ([]model.RoleListItem, int64, error) {
    var list []model.RoleListItem
    var total int64
    db := r.ResolveDB(ctx, tx)
    db = db.Model(&model.Role{})
    if v := query.Name; v != nil && *v != "" {
        db = db.Where("name LIKE ? ESCAPE '!'", "%"+sqlutil.EscapeLike(*v)+"%")
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
    err := db.Order("created_at desc").Offset(query.Offset).Limit(query.Limit).Find(&list).Error
    return list, total, err
}
```

关键点：

- 使用 `r.ResolveDB(ctx, tx)` 获取 DB 实例（自动处理 context 和事务）
- 每个方法接受 `ctx context.Context` 和 `tx *gorm.DB`（事务支持）
- 枚举值先校验 `IsValid()` 再使用，非法值忽略该条件
- LIKE 查询注意防注入：用参数化 `?` 占位符拼接
- 模糊搜索统一用 `sqlutil.EscapeLike` 转义搜索词，**并给 SQL 补上 `ESCAPE '!'` 子句**。
  只转义不写 `ESCAPE` 是错的：SQLite 没有默认转义字符，会把转义符当成普通字符去匹配，
  搜索词里含 `_` `%` `\` 时一条都查不到（MySQL/PostgreSQL 默认转义符恰好是反斜杠，
  漏写在那两家上碰巧是对的，所以这个坑只在 SQLite 上暴露）。参考实现见
  [pkg/sqlutil/like.go](../../pkg/sqlutil/like.go)
- 排序白名单用 `sortColumns` / `sortOrder` 两个 map 兜底，非法参数静默回退默认排序

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
