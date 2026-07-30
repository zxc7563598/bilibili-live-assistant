# CLAUDE.md — 枚举规范

## 位置

所有枚举统一定义在 `internal/enum/`，作为全局通用模块使用。

## 设计原则

以下场景必须使用枚举：
- 字段取值范围有限（状态、性别、开关等）
- 多个模块共享同一语义的数据
- 需要与多语言展示结合

**严禁在代码中直接使用 `0/1/2` 等魔法值**，统一通过枚举表达语义。

## 标准枚举定义

每个枚举必须实现三个方法：

```go
type YesNo int

const (
    No  YesNo = iota  // 0
    Yes               // 1
)

// Key 返回 i18n key
func (y YesNo) Key() string {
    switch y {
    case No:
        return "no"
    case Yes:
        return "yes"
    default:
        return "unknown"
    }
}

// Text 返回当前语言下的文本
func (y YesNo) Text(lang string) string {
    return i18n.T(lang, y.Key())
}

// IsValid 校验枚举值是否合法（用于参数校验）
func (y YesNo) IsValid() bool {
    switch y {
    case No, Yes:
        return true
    default:
        return false
    }
}
```

### 方法说明

| 方法 | 用途 |
|------|------|
| `Key() string` | 返回 i18n key，与多语言模块解耦 |
| `Text(lang string) string` | 返回当前语言下的展示文本，内部调用 i18n |
| `IsValid() bool` | 校验枚举值合法性，**请求参数校验时必须使用** |

## 类型选择

- 枚举类型与数据库字段类型保持一致
  - 数据库 `tinyint` → 使用 `int` 枚举
  - 数据库 `varchar` → 使用 `string` 枚举

## 使用示例

```go
// Repository 中的查询
if v := query.Enable; v != nil {
    e := enum.Enable(*v)
    if e.IsValid() {  // 先校验再使用
        db = db.Where("enable = ?", e)
    }
}

// 响应中的展示文本
item.StatusText = enum.Status(item.Status).Text(lang)
```

## 核心约定

1. **避免硬编码**：不直接使用 `0/1/2`，统一用枚举
2. **全局统一语义**：同一含义在全系统中编码规则一致（不允许 A 模块 0=女1=男，B 模块 1=男2=女）
3. **参数校验必须用 `IsValid()`**：所有外部输入的枚举值必须校验
4. **不在业务层重复定义**：枚举集中在 `internal/enum/` 统一管理
