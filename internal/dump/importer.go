package dump

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zxc7563598/bilibili-live-assistant/internal/migrate"
	"gorm.io/gorm"
)

// DefaultImportBatchSize 导入时攒够多少行写一次库
const DefaultImportBatchSize = 500

// maxPlaceholders 单条 INSERT 的参数个数上限，取三种数据库里最小的那个量级再留一半余量
// （SQLite 32766、MySQL / PostgreSQL 65535），按列数折算成行数
const maxPlaceholders = 20000

// ImportOptions 导入参数
type ImportOptions struct {
	// Path 导出文件路径，gzip 内容按文件头自动识别
	Path string
	// BatchSize 攒够多少行写一次库，为 0 时取 DefaultImportBatchSize
	BatchSize int
	// Logf 进度输出，为 nil 时静默
	Logf func(format string, args ...any)
}

// ImportResult 导入结果
type ImportResult struct {
	Path     string
	Tables   int
	Rows     int64
	Duration time.Duration
}

// tableState 导入过程中一张表的状态
type tableState struct {
	meta       TableMeta
	kindByName map[string]ColumnMeta
	tx         *gorm.DB
	pending    []map[string]any
	written    int64
	// missing 模型里有、导出文件里没有的列，这些列交给数据库默认值
	missing []string
}

// importer 导入过程中的状态
type importer struct {
	db         *gorm.DB
	source     Dialect
	target     Dialect
	batch      int
	logf       func(format string, args ...any)
	current    *tableState
	result     ImportResult
	started    time.Time
	lastLogAt  int64
	tableTotal int
	// footer 文件尾声明的总表数与总行数，用于收尾时校验文件完整性
	footer *Footer
}

// Import 从导出文件导入数据。
//
// 只往空库里导：非空目标库直接拒绝。这样导入失败时不会留下一半旧数据一半新数据的库，
// 重试也只需要清空重来。
func Import(db *gorm.DB, opts ImportOptions) (*ImportResult, error) {
	if strings.TrimSpace(opts.Path) == "" {
		return nil, fmt.Errorf("导入文件路径不能为空")
	}
	tables, err := Tables(db)
	if err != nil {
		return nil, err
	}
	logf := opts.Logf
	if logf == nil {
		logf = func(string, ...any) {}
	}
	// 先打开文件：路径写错时就别去碰目标库了
	file, err := openDumpFile(opts.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := checkEmptyDatabase(db, tables); err != nil {
		return nil, err
	}
	// 表结构按模型定义重建，不依赖导出文件里的建表语句——跨方言导入靠的就是这一步
	if err := migrate.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("创建表结构失败: %w", err)
	}

	batch := opts.BatchSize
	if batch <= 0 {
		batch = DefaultImportBatchSize
	}
	im := &importer{
		db:         db,
		target:     Dialect{Name: db.Dialector.Name()},
		batch:      batch,
		logf:       logf,
		started:    time.Now(),
		tableTotal: len(tables),
		result:     ImportResult{Path: opts.Path},
	}
	// 任何一步失败都要把还没提交的表回滚掉，别把事务挂在半开状态
	defer im.abort()
	if err := im.run(NewReader(file), tables); err != nil {
		return nil, err
	}
	if err := im.finish(tables); err != nil {
		return nil, err
	}
	im.result.Duration = time.Since(im.started)
	return &im.result, nil
}

// run 逐个事件消费导出文件
func (im *importer) run(reader *Reader, tables []Table) error {
	known := make(map[string]Table, len(tables))
	for _, t := range tables {
		known[t.Name] = t
	}
	sawFooter := false
	for {
		event, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		switch event.Kind {
		case EventHeader:
			// 源方言决定字面量怎么解转义，读取器保证它出现在任何数据行之前
			im.source = reader.dialect
			im.logf("导出文件来自 %s", reader.dialect.Name)
		case EventTable:
			definition, ok := known[event.Table.Name]
			if !ok {
				return fmt.Errorf("导出文件里的表 %s 在当前版本中不存在，无法导入", event.Table.Name)
			}
			if err := im.startTable(event.Table, definition); err != nil {
				return err
			}
		case EventRow:
			if err := im.addRow(event); err != nil {
				return err
			}
		case EventFooter:
			sawFooter = true
			footer := event.Footer
			im.footer = &footer
		}
	}
	if !sawFooter {
		return fmt.Errorf("%w：缺少 BLA-END 结束标记，可能是文件被截断", errBrokenFile)
	}
	return nil
}

// startTable 结束上一张表并开启新表的事务
func (im *importer) startTable(meta TableMeta, definition Table) error {
	if err := im.endTable(); err != nil {
		return err
	}
	state := &tableState{
		meta:       meta,
		kindByName: make(map[string]ColumnMeta, len(meta.Columns)),
	}
	for _, col := range meta.Columns {
		if _, ok := definition.Schema.FieldsByDBName[col.Name]; !ok {
			return fmt.Errorf("导出文件里的表 %s 包含当前版本不认识的列 %s，无法导入", meta.Name, col.Name)
		}
		state.kindByName[col.Name] = col
	}
	for _, col := range definition.Columns {
		if _, ok := state.kindByName[col.Name]; !ok {
			state.missing = append(state.missing, col.Name)
		}
	}
	if len(state.missing) > 0 {
		im.logf("提示：表 %s 的列 %s 在导出文件里没有，将由数据库默认值填充",
			meta.Name, strings.Join(state.missing, "、"))
	}
	im.logf("[%2d/%d] %-26s 开始导入，%d 行", im.result.Tables+1, im.tableTotal, meta.Name, meta.Rows)
	im.lastLogAt = 0
	state.tx = im.db.Begin()
	if state.tx.Error != nil {
		return fmt.Errorf("开启事务失败: %w", state.tx.Error)
	}
	im.current = state
	return nil
}

// addRow 解析一行并攒进批次
func (im *importer) addRow(event Event) error {
	if im.current == nil {
		return fmt.Errorf("导出文件里的数据行出现在任何表定义之前，文件格式不正确")
	}
	if event.TableName != im.current.meta.Name {
		return fmt.Errorf("数据行属于表 %s，但当前表是 %s，文件格式不正确", event.TableName, im.current.meta.Name)
	}
	if len(event.Columns) != len(event.Values) {
		return fmt.Errorf("表 %s 的数据行列数与取值个数不一致，文件格式不正确", event.TableName)
	}
	row := make(map[string]any, len(event.Columns))
	for i, name := range event.Columns {
		col, ok := im.current.kindByName[name]
		if !ok {
			return fmt.Errorf("表 %s 的数据行包含未声明的列 %s", event.TableName, name)
		}
		value, err := parseLiteral(event.Values[i], col.Kind, im.source)
		if err != nil {
			return fmt.Errorf("表 %s 列 %s 解析失败: %w", event.TableName, name, err)
		}
		if err := im.checkStorable(event.TableName, name, value); err != nil {
			return err
		}
		row[name] = value
	}
	im.current.pending = append(im.current.pending, row)
	if len(im.current.pending) >= im.batch {
		return im.flush()
	}
	return nil
}

// checkStorable 拦下目标库存不了的字符串。
//
// SQLite 什么字节都收，但 PostgreSQL 的文本列不接受 NUL 字节、MySQL 与 PostgreSQL
// 都不接受非法 UTF-8。提前拦下来给一句能看懂的话，否则要等写库时才报一个与数据无关的驱动错误，
// 那时前面的表已经提交，用户既不知道是哪张表哪一列，也不知道该怎么办。
func (im *importer) checkStorable(table, column string, value any) error {
	if im.target.Name == DialectSQLite {
		return nil
	}
	text, ok := value.(string)
	if !ok {
		return nil
	}
	if im.target.Name == DialectPostgres && strings.ContainsRune(text, 0) {
		return fmt.Errorf("表 %s 列 %s 的值含 NUL 字节，PostgreSQL 的文本列无法存储；"+
			"请先在源库清理该字段再重新导出", table, column)
	}
	if !utf8.ValidString(text) {
		return fmt.Errorf("表 %s 列 %s 的值不是合法的 UTF-8，%s 无法存储；"+
			"请先在源库清理该字段再重新导出", table, column, im.target.Name)
	}
	return nil
}

// flush 把一个批次写进库
func (im *importer) flush() error {
	state := im.current
	if state == nil || len(state.pending) == 0 {
		return nil
	}
	rows := state.pending
	state.pending = nil
	batch := im.batch
	if limit := maxPlaceholders / max(1, len(state.meta.Columns)); limit < batch {
		batch = max(1, limit)
	}
	if err := state.tx.Table(state.meta.Name).CreateInBatches(rows, batch).Error; err != nil {
		return fmt.Errorf("写入表 %s 失败: %w", state.meta.Name, err)
	}
	state.written += int64(len(rows))
	im.result.Rows += int64(len(rows))
	im.logProgress()
	return nil
}

// logProgress 单表数据量很大时按行数给出中间进度
func (im *importer) logProgress() {
	const step = 50000
	state := im.current
	if state == nil || state.written-im.lastLogAt < step {
		return
	}
	im.lastLogAt = state.written
	im.logf("  %s: 已写入 %d/%d 行", state.meta.Name, state.written, state.meta.Rows)
}

// endTable 冲刷并提交当前表
func (im *importer) endTable() error {
	state := im.current
	if state == nil {
		return nil
	}
	if err := im.flush(); err != nil {
		im.rollback(state)
		return err
	}
	// 声明行数与实际写入对不上说明文件不完整，回滚本表而不是留下半张表的数据
	if state.written != state.meta.Rows {
		im.rollback(state)
		return fmt.Errorf("表 %s 声明 %d 行，实际写入 %d 行：%w", state.meta.Name, state.meta.Rows, state.written, errBrokenFile)
	}
	if state.tx != nil {
		if err := state.tx.Commit().Error; err != nil {
			return fmt.Errorf("提交表 %s 失败: %w", state.meta.Name, err)
		}
	}
	im.result.Tables++
	im.current = nil
	return nil
}

func (im *importer) rollback(state *tableState) {
	if state.tx != nil {
		state.tx.Rollback()
	}
}

// abort 回滚当前还没提交的表。SQLite 连接池只有一条连接，
// 事务挂着不放会把后续所有数据库操作都卡死，所以失败路径必须回滚。
func (im *importer) abort() {
	if im.current == nil {
		return
	}
	im.rollback(im.current)
	im.current = nil
}

// finish 收尾：补跑历史数据补丁、校验文件尾、同步 PostgreSQL 自增序列
func (im *importer) finish(tables []Table) error {
	if err := im.endTable(); err != nil {
		return err
	}
	// 文件尾声明了总表数与总行数，与实际对不上说明文件被动过或写了一半
	if im.footer != nil {
		if im.footer.Tables != im.result.Tables || im.footer.Rows != im.result.Rows {
			return fmt.Errorf("文件尾声明 %d 张表 %d 行，实际导入 %d 张表 %d 行：%w",
				im.footer.Tables, im.footer.Rows, im.result.Tables, im.result.Rows, errBrokenFile)
		}
	}
	// 只补跑历史数据补丁，不调用 Seed：导出文件里本就有源库的种子数据，再插一遍是多余动作；
	// 而且种子数据写死了主键 ID，源库里这些 ID 一旦被别的记录占用就会主键冲突，
	// 那时数据早已落库，整个导入会被这一步反向打成失败
	if err := migrate.Run(im.db); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}
	// 导出文件里是显式 ID，PostgreSQL 的序列仍停在起点，不同步的话之后新增数据必然主键冲突
	names := make([]string, 0, len(tables))
	for _, t := range tables {
		names = append(names, t.Name)
	}
	return migrate.SyncPostgresSequences(im.db, names)
}

// checkEmptyDatabase 目标库里任何一张业务表有数据都拒绝导入
func checkEmptyDatabase(db *gorm.DB, tables []Table) error {
	var occupied []string
	for _, table := range tables {
		if !db.Migrator().HasTable(table.Name) {
			continue
		}
		var count int64
		// 软删除的行也算数据，否则会把"只有软删除数据"的库当成空库
		if err := db.Table(table.Name).Unscoped().Count(&count).Error; err != nil {
			return fmt.Errorf("检查表 %s 是否为空失败: %w", table.Name, err)
		}
		if count > 0 {
			occupied = append(occupied, fmt.Sprintf("%s（%d 行）", table.Name, count))
		}
	}
	if len(occupied) > 0 {
		return fmt.Errorf("目标库不是空库，以下表已有数据：%s\n"+
			"导入只允许写空库；换库请把配置指向一个新库，或先清空这些表",
			strings.Join(occupied, "、"))
	}
	return nil
}

// openDumpFile 打开导出文件，按文件头自动识别 gzip
func openDumpFile(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开导入文件失败: %w", err)
	}
	buffered := bufio.NewReaderSize(file, 256*1024)
	magic, err := buffered.Peek(2)
	if err == nil && magic[0] == 0x1f && magic[1] == 0x8b {
		reader, err := gzip.NewReader(buffered)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("解压导入文件失败: %w", err)
		}
		return &multiCloser{Reader: reader, closers: []io.Closer{reader, file}}, nil
	}
	return &multiCloser{Reader: buffered, closers: []io.Closer{file}}, nil
}

// multiCloser 让解压流与底层文件一起关闭
type multiCloser struct {
	io.Reader
	closers []io.Closer
}

func (m *multiCloser) Close() error {
	var firstErr error
	for _, closer := range m.closers {
		if err := closer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
