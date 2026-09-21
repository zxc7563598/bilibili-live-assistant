# CLAUDE.md — 导出模块（Export）

## 职责

提供与业务无关的「大表导出」能力：**短时效下载凭证 + CSV 流式写出**。
本包不 import 任何业务模块，各业务模块实现 `Source` 接口后由 `bootstrap/export.go` 注册进来。

对外只有两个接口（`internal/handler/export`）：

```
POST /api/admin/export/ticket      (AdminAuth，JSON 信封，所有校验在这里)
  → { url, filename, total, expires_at }
GET  /api/admin/export/download?ticket=xxx    (无 AdminAuth，凭证即鉴权)
  → text/csv 流式响应
```

## 四条决定实现的约束

改动本模块前请先理解这四条 —— 它们各自否掉了一个「更直觉」的写法。

### 1. 分块主键 seek，不用 `Rows()` 长游标

默认 sqlite 驱动的连接池被**硬编码为 1**（`internal/config/database.go`，注释：SQLite 仅支持单写者），
且未开 WAL。一个长生命周期的 `Rows()` 游标会独占那唯一一条连接，把整个后端冻住整场导出 ——
其他管理页请求全部排队，直播监听也无法写入。

因此改为每块一条独立短查询（`WHERE id < :lastID ... ORDER BY id DESC LIMIT :batch`），
块与块之间连接归还，其他请求能在间隙正常穿插。

**代价**：分块的排序键必须是**有索引的唯一列**。所以导出行序固定为主键倒序，
**不跟随页面排序** —— 跟随排序要按 `send_at` 这类无索引列排，每块都得全表扫描重排。

### 2. 所有校验必须在获取凭证时做完

下载接口走浏览器原生下载，响应体会被直接存成文件。在那里返回 JSON 错误，
用户只会得到一个内容不对的 `.csv`。所以模块/列/行数上限/并发数**全部在 `CreateTicket` 判定**，
走正常信封让前端能弹提示。

下载阶段只可能因凭证失效（竞态）或中途 DB 出错而失败。因此 handler 的约定是：
**第一块数据成功取到之前不写任何字节**（`Sink.Start` 只在那之后调用），在那之前失败还能用
错误状态码表达；之后失败只能追加一行 `#导出中断` 标记后中断 —— 静默截断比可见标记更糟。

### 3. `write_timeout` 要滚动续期，不能清零

`cmd/server/main.go` 全局设了 `WriteTimeout: 15s`，且 `net/http` 是在**读完请求头后**
一次性设下绝对期限的 —— 几分钟的导出必然被切断。handler 用
`http.NewResponseController(c.Writer).SetWriteDeadline(now + write_idle_timeout)` 处理，
**每写一块续期一次**。

不要改成 `time.Time{}`（等于永不超时）：一个卡死的客户端能就此永久占住 goroutine 与连接。
也不要改全局 `write_timeout`，其它接口的 15 秒保护必须保留。

### 4. 凭证无状态签名，不落库

`base64url(payload) . base64url(HMAC-SHA256(派生key, payload))`，两段式（与三段式 JWT 形状不同，
不可能混淆）。payload 含管理员ID、模块、列、筛选条件、语言、过期时间。

- **不依赖 Redis**：Redis 在本项目是可选的（未配置时 `rdb == nil`），拿它存凭证就得再写内存兜底。
- **不把 accessToken 放进 URL**：那是 2 小时有效、全权限的凭证，会落进访问日志与浏览器历史。
  凭证只有 120 秒、限单模块、只授予「重下这份 CSV」。
- 密钥由 `jwt.secret` 派生（`HMAC("export-ticket:v1", secret)`）——**绝不能用同一把密钥
  直接签两种凭证**，那样两边的签名可以互相换算。

## 各业务模块如何接入

1. `internal/service/<模块>/export.go` 实现 `Source`：

   | 方法 | 说明 |
   |------|------|
   | `Module()` | 模块名，与前端 `MeCrud` 的 `export-module` 一致 |
   | `Columns()` | 允许导出的列（`Key` + i18n 的 `TitleKey`）。**不含裸 `id`** |
   | `Normalize()` | 解析校验前端筛选条件，回写规范化 JSON；失败返回**本模块**的错误码 |
   | `Count()` | 命中行数，必须用有界的子查询（`LIMIT limit`）早停，不要全表 `COUNT(*)` |
   | `FetchChunk()` | 按主键倒序取一块，游标用上一块最后一行的主键 |

2. 仓储加两个方法：有界计数 + 分块读取，写法与 `ListPage` 一致（`ResolveDB` + 白名单排序）。
   筛选条件务必与 `ListPage` **共用同一份拼装函数**，否则导出结果与页面对不上。

3. `internal/bootstrap/export.go` 的注册列表里加一行。

4. `internal/i18n/locales/{zh,en}.yaml` 补 `export.module.<模块>`（文件名）与
   `export.column.<key>`（表头）。**中英两份都要加**，漏加不会编译失败，只会渲染成 key。

5. 前端给对应页面的 `<MeCrud>` 加 `export-module="<模块>"`。

6. 配一个单测遍历 `Columns()`，对零值 model 逐个调取值函数 —— 列声明与取值 switch 是两份
   手工对齐的字面量，漂移了要等线上点导出才会暴露，让它变成一次 `go test` 就能发现的事。

## 取值与格式化

模块的 `FetchChunk` 返回 `[][]any`，渲染由本包的 `formatCell` 统一处理，
模块只做「列 key → 字段」的映射：

| 返回类型 | 输出 |
|----------|------|
| `export.UnixTime` | `2006-01-02 15:04:05`，0 → 空 |
| `export.Money`（分） | 两位小数的元 |
| 实现了 `Text(lang)` 的枚举 | 当前语言的文案 |
| `bool` | yes/no 文案 |
| `string` | 过一遍公式注入防护 |
| 其它 | `fmt.Sprint` 兜底，并往 CSV 追加一条未支持类型提示 |

CSV 细节：首字节写 UTF-8 BOM（Windows Excel 中文不乱码）、`UseCRLF`、
**公式注入防护**（`= + - @` 开头加单引号；加引号不能替代这层防护）。

## 错误码

MM=17，见 `code.go`。各模块自己的筛选校验错误用**本模块**的码（如弹幕模块的 `10601`）。
文案里不能带参数（`i18n.E` 只查表，不支持插值）。
