package dump

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// DefaultBatchSize 单次读取并写出的行数
const DefaultBatchSize = 500

// ExportOptions 导出参数
type ExportOptions struct {
	// Path 输出文件路径，以 .gz 结尾时写出 gzip 压缩内容
	Path string
	// Force 为真时覆盖已存在的输出文件
	Force bool
	// BatchSize 单次读取并写出的行数，为 0 时取 DefaultBatchSize
	BatchSize int
	// Logf 进度输出，为 nil 时静默
	Logf func(format string, args ...any)
}

// ExportResult 导出结果
type ExportResult struct {
	Path     string
	Size     int64
	Tables   int
	Rows     int64
	Duration time.Duration
}

// exporter 导出过程中的状态
type exporter struct {
	db      *gorm.DB
	dialect Dialect
	batch   int
	out     *bufio.Writer
	closers []io.Closer
	logf    func(format string, args ...any)
}

// Export 把当前连接的数据库导出成 SQL 文件。
//
// 先写临时文件、成功后再改名：中途失败不会留下一个看起来完整的半截备份。
func Export(db *gorm.DB, opts ExportOptions) (*ExportResult, error) {
	started := time.Now()
	dialect := Dialect{Name: db.Dialector.Name()}
	if !dialect.Valid() {
		return nil, fmt.Errorf("不支持的数据库类型: %s", dialect.Name)
	}
	if strings.TrimSpace(opts.Path) == "" {
		return nil, fmt.Errorf("导出文件路径不能为空")
	}
	tables, err := Tables(db)
	if err != nil {
		return nil, err
	}
	if !opts.Force {
		if _, err := os.Stat(opts.Path); err == nil {
			return nil, fmt.Errorf("导出文件已存在: %s（如需覆盖请加 -force）", opts.Path)
		}
	}
	batch := opts.BatchSize
	if batch <= 0 {
		batch = DefaultBatchSize
	}
	logf := opts.Logf
	if logf == nil {
		logf = func(string, ...any) {}
	}

	dir := filepath.Dir(opts.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建导出目录失败: %w", err)
	}
	tmpPath := opts.Path + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("创建导出文件失败: %w", err)
	}
	// 任何一步失败都删掉临时文件，不留下半截产物
	success := false
	defer func() {
		if !success {
			file.Close()
			os.Remove(tmpPath)
		}
	}()

	e := &exporter{db: db, dialect: dialect, batch: batch, logf: logf}
	sink := io.Writer(file)
	if strings.HasSuffix(strings.ToLower(opts.Path), ".gz") {
		gz := gzip.NewWriter(file)
		e.closers = append(e.closers, gz)
		sink = gz
	}
	e.out = bufio.NewWriterSize(sink, 256*1024)

	result := &ExportResult{Path: opts.Path}
	if err := e.writeBody(tables, result); err != nil {
		return nil, err
	}
	// 先冲刷缓冲区，再关 gzip，最后关文件，顺序反了会丢掉尾部数据
	if err := e.out.Flush(); err != nil {
		return nil, fmt.Errorf("写出导出文件失败: %w", err)
	}
	for _, closer := range e.closers {
		if err := closer.Close(); err != nil {
			return nil, fmt.Errorf("关闭压缩流失败: %w", err)
		}
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("关闭导出文件失败: %w", err)
	}
	if err := os.Rename(tmpPath, opts.Path); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("重命名导出文件失败: %w", err)
	}
	success = true

	if info, err := os.Stat(opts.Path); err == nil {
		result.Size = info.Size()
	}
	result.Duration = time.Since(started)
	return result, nil
}

// writeBody 依次写出文件头、各表数据与文件尾
func (e *exporter) writeBody(tables []Table, result *ExportResult) error {
	if err := e.writeHeader(); err != nil {
		return err
	}
	if err := e.writePreamble(); err != nil {
		return err
	}
	for i, table := range tables {
		rows, err := e.writeTable(table)
		if err != nil {
			return err
		}
		result.Rows += rows
		e.logf("[%2d/%d] %-26s %d 行", i+1, len(tables), table.Name, rows)
	}
	result.Tables = len(tables)
	line, err := marker(endMarker, Footer{Tables: len(tables), Rows: result.Rows})
	if err != nil {
		return err
	}
	return e.printf("%s\n", line)
}

// writeHeader 写出文件头与说明
func (e *exporter) writeHeader() error {
	line, err := marker(headerMarker, Header{
		Version:   FileVersion,
		Dialect:   e.dialect.Name,
		CreatedAt: time.Now().Unix(),
	})
	if err != nil {
		return err
	}
	return e.printf("-- BiliLiveAssistant database dump\n%s\n-- %s\n\n", line, headerNote)
}

// writePreamble 写出方言相关的前置语句
func (e *exporter) writePreamble() error {
	var prelude string
	switch e.dialect.Name {
	case DialectMySQL:
		prelude = "SET NAMES utf8mb4;"
	case DialectPostgres:
		prelude = "SET client_encoding = 'UTF8';"
	default:
		prelude = "PRAGMA foreign_keys = OFF;"
	}
	return e.printf("%s\n\n", prelude)
}

// writeTable 写出一张表的建表语句、索引与全部数据，返回实际写出的行数
func (e *exporter) writeTable(t Table) (int64, error) {
	total, err := e.countRows(t)
	if err != nil {
		return 0, err
	}
	if err := e.printf("-- ---------------------------------------------\n%s\n", createTableSQL(e.db, t, e.dialect)); err != nil {
		return 0, err
	}
	for _, sql := range createIndexSQLs(t, e.dialect) {
		if err := e.printf("%s\n", sql); err != nil {
			return 0, err
		}
	}
	meta := TableMeta{Name: t.Name, Rows: total, Columns: make([]ColumnMeta, 0, len(t.Columns))}
	for _, col := range t.Columns {
		meta.Columns = append(meta.Columns, ColumnMeta{Name: col.Name, Kind: col.Kind, Nullable: col.Nullable})
	}
	line, err := marker(tableMarker, meta)
	if err != nil {
		return 0, err
	}
	if err := e.printf("\n%s\n\n", line); err != nil {
		return 0, err
	}

	written, err := e.writeRows(t)
	if err != nil {
		return 0, err
	}
	// 导出期间有并发写入时，实际行数与表头声明会对不上，此时宁可报错也不要一个不一致的备份
	if written != total {
		return 0, fmt.Errorf("表 %s 写出 %d 行，与导出时的 %d 行不一致："+
			"导出期间数据库仍在写入，请停掉服务后重新导出", t.Name, written, total)
	}
	return written, nil
}

// countRows 统计行数（含软删除行），用于进度与导出后的一致性校验
func (e *exporter) countRows(t Table) (int64, error) {
	// 表不存在通常是连到了空库（配置文件里的数据库还没跑过服务），单独给一句能看懂的提示
	if !e.db.Migrator().HasTable(t.Name) {
		return 0, fmt.Errorf("数据库里没有表 %s：请确认配置指向的库里已经有本系统的数据（空库请先启动一次服务建表）", t.Name)
	}
	var total int64
	if err := e.db.Table(t.Name).Unscoped().Count(&total).Error; err != nil {
		return 0, fmt.Errorf("统计表 %s 行数失败: %w", t.Name, err)
	}
	return total, nil
}

// writeRows 按主键游标分块读取并写出数据。
//
// 用主键游标而非 OFFSET：大表上 OFFSET 会越翻越慢，SQLite 单连接下还容易长时间占着连接。
func (e *exporter) writeRows(t Table) (int64, error) {
	primaryKey, ok := t.PrimaryColumn()
	if !ok {
		return 0, fmt.Errorf("表 %s 没有主键，无法分块导出", t.Name)
	}
	columns := make([]string, 0, len(t.Columns))
	for _, col := range t.Columns {
		columns = append(columns, e.dialect.quote(col.Name))
	}
	selectSQL := strings.Join(columns, ", ")

	var written int64
	cursor := int64(math.MinInt64)
	for {
		rows := make([]map[string]any, 0, e.batch)
		err := e.db.Table(t.Name).Unscoped().
			Select(selectSQL).
			Where(e.dialect.quote(primaryKey)+" > ?", cursor).
			Order(e.dialect.quote(primaryKey) + " ASC").
			Limit(e.batch).
			Find(&rows).Error
		if err != nil {
			return written, fmt.Errorf("读取表 %s 失败: %w", t.Name, err)
		}
		if len(rows) == 0 {
			return written, nil
		}
		if err := e.writeInsert(t, rows); err != nil {
			return written, err
		}
		next, err := toInt(rows[len(rows)-1][primaryKey])
		if err != nil {
			return written, fmt.Errorf("表 %s 主键 %s 取值失败: %w", t.Name, primaryKey, err)
		}
		cursor = next
		written += int64(len(rows))
		if len(rows) < e.batch {
			return written, nil
		}
	}
}

// writeInsert 把一块数据写成一条多行 INSERT，一行一个元组
func (e *exporter) writeInsert(t Table, rows []map[string]any) error {
	columns := make([]string, 0, len(t.Columns))
	for _, col := range t.Columns {
		columns = append(columns, e.dialect.quote(col.Name))
	}
	tuples := make([]string, 0, len(rows))
	for _, row := range rows {
		values := make([]string, 0, len(t.Columns))
		for _, col := range t.Columns {
			value, ok := row[col.Name]
			if !ok {
				return fmt.Errorf("表 %s 的查询结果缺少列 %s，无法导出", t.Name, col.Name)
			}
			literal, err := renderValue(value, col.Kind, e.dialect)
			if err != nil {
				return fmt.Errorf("表 %s 列 %s 渲染失败: %w", t.Name, col.Name, err)
			}
			values = append(values, literal)
		}
		tuples = append(tuples, "("+strings.Join(values, ",")+")")
	}
	return e.printf("INSERT INTO %s (%s) VALUES\n%s;\n", e.dialect.quote(t.Name), strings.Join(columns, ","), strings.Join(tuples, ",\n"))
}

func (e *exporter) printf(format string, args ...any) error {
	if _, err := fmt.Fprintf(e.out, format, args...); err != nil {
		return fmt.Errorf("写出导出文件失败: %w", err)
	}
	return nil
}
