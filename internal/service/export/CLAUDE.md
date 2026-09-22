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
   | `Columns()` | 允许导出的列，直接用 `export.ColumnsOf(exportColumns)` 映射下面的字面量 |
   | `Normalize()` | 解析校验前端筛选条件，回写规范化 JSON；失败返回**本模块**的错误码 |
   | `Count()` | 命中行数，必须用有界的子查询（`LIMIT limit`）早停，不要全表 `COUNT(*)` |
   | `FetchChunk()` | 按主键倒序取一块，游标用上一块最后一行的主键；取值查 `export.ValueMapOf(exportColumns)` |

   **列声明与取值写在同一个 `[]export.ColumnSpec[行类型]` 字面量里**，不要拆成
   「列清单 + 取值 switch」两份手工对齐的表 —— 后者加一列忘了加分支不会编译失败，
   要等点导出才暴露，还得为每个模块补一个同步单测。合成一份后这类漂移结构上不会发生。
   列里**不含裸 `id`**（内部主键，各模块含义不一致）。

2. 仓储加两个方法：有界计数 + 分块读取，写法与 `ListPage` 一致（`ResolveDB` + 白名单排序）。
   筛选条件务必与 `ListPage` **共用同一份拼装函数**，否则导出结果与页面对不上。
   注意 `ListPage` 里那些**不在共享构建器内**的额外条件也要一并复现
   （例如盲盒列表额外带的 `original = 0`、订单列表的 `LEFT JOIN` 与 `Select` 投影）——
   漏掉会静默导出范围不对的数据。

3. `internal/bootstrap/export.go` 的注册列表里加一行。

4. `internal/i18n/locales/{zh,en}.yaml` 补 `export.module.<模块>`（文件名）与
   `export.column.<key>`（表头）。**中英两份都要加**，漏加不会编译失败，只会渲染成 key。

5. 前端给对应页面的 `<MeCrud>` 加 `export-module="<模块>"`。
   页面上**每个带 `title` 且未标 `hideInExcel` 的列**的 key 都必须在白名单里，否则领票会被
   `11701` 拒绝（动作列尤其容易漏标 `hideInExcel`）；列的 key 与后端字段不一致时用列上的
   `exportKey` 覆盖。

6. 若单元格显示的内容**多于该列 key 对应的那一个字段**（如订单的「商品信息」一格里有名称、数量、
   规格），在模块里声明一个拼接列（如 `product_info`）并在前端用 `exportKey` 指过去，
   文案对齐前端单元格。这是唯一会把前端展示逻辑复制到后端的地方，两处注释都要互相指认。

## 命名与文件结构

各模块的 `export.go` 逐字对齐，照着最接近的一个改即可（单模块看 `livedanmu`，
一个 service 挂两个模块看 `livegift`）。文件内固定按这个顺序：

```go
// exportColumns 允许导出的列
var exportColumns = []export.ColumnSpec[行类型]{...}

var columnValue = export.ValueMapOf(exportColumns)

// exportFilterInput 前端传来的筛选条件，字段与列表接口同口径
type exportFilterInput struct{...}

// exportFilterQuery 规范化后的筛选条件：时间区间已换算成秒级的闭区间
type exportFilterQuery struct{...}   // 只做透传的模块不需要这个类型

// Module / Columns / Normalize / Count / FetchChunk —— 注释与其它模块逐字相同
// parseExportQuery 把规范化的筛选条件还原为仓储查询结构
// 模块私有的辅助函数（如拼接列）放最后
```

几条约定：

- 名字一律用 `exportColumns` / `columnValue` / `exportFilterInput` / `exportFilterQuery` /
  `parseExportQuery`。一个 service 挂两个导出模块时（目前只有 `livegift`）加列表名前缀、
  基名不变：`giftExportColumns`、`blindBoxExportFilterInput`、`parseGiftExportQuery`。
- 注释只写一行结论，**不写「为什么」**：`// exportColumns 允许导出的列`、
  `// exportFilterInput 前端传来的筛选条件，字段与列表接口同口径` 就是完整形态。
  取舍与背景写进本文件的「各模块的设计取舍」一节，不要写回代码里 —— 代码上看不出来
  为什么这么写的地方，都在那一节有交代。
- 文件内只有一个模块时不加分组标记，用不上分隔线；`livegift` 因为一个文件放了两个模块，
  才用 `// ---------- 礼物列表 ----------` 这类标记隔开。
- `FetchChunk` 结尾统一走
  `export.BuildRecords(rows, keys, columnValue, func(r 行类型) int64 { return r.ID })`，
  不要各写一遍取值循环。

## 各模块的设计取舍

代码里只剩一行注释，为什么这么做都记在这里。加新模块或改这些模块前先扫一眼。

### 框架层（本包）

- **列声明与取值写在同一个 `ColumnSpec` 里**：早期是「列清单 + 取值 switch」两份手工对齐的
  字面量，加一列忘了加分支不会编译失败，要等点导出才暴露，每个模块还得配一个同步单测兜着。
  合成一份后这类漂移在结构上不可能出现。
- `ValueMapOf` 的结果在包级初始化一次即可；导出是只读操作，之后并发调用安全。
- `BuildRecords` 的第二个返回值是下一块的游标（本块最后一行的主键，由 `idOf` 取出）。
  各模块的 `FetchChunk` 因此一行收尾，取数与游标的写法不会有出入。

### liveuser（管理端「用户列表」+ 商城端「用户管理」，同一个接口）

- 白名单**不含 `password` / `token`**：model 里有这两列，它们是用户凭证，不能出现在导出文件里。
- `total_gift_amount` 落库单位是分，用 `export.Money` 渲染成两位小数的元，
  与页面 `row.total_gift_amount / 100` 的展示口径一致。

### livegift / livegiftblindbox（一个 service，两个接口）

- 两个接口的列与筛选条件都不一样，所以注册成两个模块名；Service 只有一个，
  用两个薄包装类型各自挂 `Source` 方法。
- `total`、`profit` 是页面上算出来的展示值（后端没有对应字段），导出侧按同样的算式算一遍；
  `profit` 可为负（收到的礼物比转赠出去的便宜就是亏）。
- 「是不是盲盒」的判据 `original = 0` 在仓储的 `blindBoxBase` 里，不在本文件 —— 因为它是
  列表的身份条件，列表 / 计数 / 导出三个入口都要带，收在仓储只写一份。
- `gift_type` / `original` 在 `Normalize` 里显式校验：仓储构建器对非法枚举是**静默忽略**的，
  不拦下来就会导出一份筛选条件没生效的全量数据。

### livepk

- `pk_status` / `battle_type` 在库里是裸 `int64`，值来自 B 站协议、后端没有对应枚举，
  页面上的中文标签也来自前端本地的选项数组，所以导出**直接出原始数值**，不在后端造第二份文案。
- 前端的 `pkStatusOptions` 里 404 = 异常结束（这个值曾与「即将开始」的 101 重复，已修）。

### order（管理端「订单列表」+「发货」，同一个接口）

- 白名单是**两个页面列的并集**：各页只发自己显示的那几列。
- `product_info` / `receiver_info` 是拼接列，文案对齐前端单元格。**这是唯一把前端展示逻辑
  复制到后端的地方**：改前端 `renderProduct`（`order/list` 与 `order/delivery` 各一份）、
  `renderReceiver`（`order/delivery`）时，要同步改本模块的 `productInfo` / `receiverInfo`，
  以及复刻 `web/src/utils/common.js` 的 `formatProductSpecs` 的那份 `formatSpecProperties`。
  落库的规格快照是按规格定义顺序序列化的 `[{"规格名":"规格值"},…]`，且每个元素恰好一对
  （见 `internal/service/product/common.go` 的 `marshalSpecProperties`），所以按元素顺序
  遍历即与页面同序。
- 状态筛选字段用 `*int`：裸 `int` 分不清「没传」与「传 0」，而 0 是合法筛选值
  （发货页的默认筛选就是 `order_status=1 & ship_status=0`）。
- `emptyCell`（`—`）对齐前端 `dash()` 的空值占位。

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
