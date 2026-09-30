package dump

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// EventKind 导出文件读出的事件类型
type EventKind int

// 事件类型
const (
	// EventHeader 文件头
	EventHeader EventKind = iota
	// EventTable 一张表开始，携带列元信息
	EventTable
	// EventRow 一行数据
	EventRow
	// EventFooter 文件尾
	EventFooter
)

// Event 导出文件中的一个事件
type Event struct {
	Kind EventKind
	// Table 仅 EventTable 有效
	Table TableMeta
	// Footer 仅 EventFooter 有效
	Footer Footer
	// TableName / Columns / Values 仅 EventRow 有效。Values 是未经解析的原始字面量。
	TableName string
	Columns   []string
	Values    []string
}

// 语句与注释的读取错误
var (
	errMissingHeader = errors.New("导出文件缺少 BLA-DUMP 文件头，请使用本系统 `db export` 生成的文件")
	errBrokenFile    = errors.New("导出文件不完整")
)

// scanner 从导出文件中顺序读出注释与语句
type scanner struct {
	reader *bufio.Reader
	// backslashEscapes MySQL 方言下反斜杠是转义符，扫描字符串时必须按转义处理
	backslashEscapes bool
}

// item 扫描出的一个元素
type item struct {
	comment bool
	text    string
}

func newScanner(r io.Reader) *scanner {
	return &scanner{reader: bufio.NewReaderSize(r, 256*1024)}
}

// next 读出下一个元素，返回 ok=false 表示已到文件末尾
func (s *scanner) next() (item, bool, error) {
	for {
		c, err := s.readByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return item{}, false, nil
			}
			return item{}, false, err
		}
		if isSpaceByte(c) {
			continue
		}
		if c == '-' {
			next, err := s.peekByte()
			if err != nil {
				return item{}, false, err
			}
			if next == '-' {
				return s.readLineComment()
			}
		}
		if c == '/' {
			next, err := s.peekByte()
			if err != nil {
				return item{}, false, err
			}
			if next == '*' {
				if err := s.skipBlockComment(); err != nil {
					return item{}, false, err
				}
				continue
			}
		}
		return s.readStatement(c)
	}
}

// readLineComment 读出一整行注释，返回的 text 已去掉前导的 "--"
func (s *scanner) readLineComment() (item, bool, error) {
	if _, err := s.readByte(); err != nil { // 吃掉第二个 '-'
		return item{}, false, err
	}
	var b strings.Builder
	for {
		c, err := s.readByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return item{comment: true, text: strings.TrimSpace(b.String())}, true, nil
			}
			return item{}, false, err
		}
		if c == '\n' {
			return item{comment: true, text: strings.TrimSpace(b.String())}, true, nil
		}
		b.WriteByte(c)
	}
}

// skipBlockComment 跳过块注释
func (s *scanner) skipBlockComment() error {
	if _, err := s.readByte(); err != nil { // 吃掉 '*'
		return err
	}
	prev := byte(0)
	for {
		c, err := s.readByte()
		if err != nil {
			return err
		}
		if prev == '*' && c == '/' {
			return nil
		}
		prev = c
	}
}

// readStatement 读出一条以分号结束的语句。first 是已经读掉的首字符。
func (s *scanner) readStatement(first byte) (item, bool, error) {
	var b strings.Builder
	b.WriteByte(first)
	// 引号内的分号不是语句结束
	var quote byte
	for {
		c, err := s.readByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return item{}, false, fmt.Errorf("%w：语句未以分号结束", errBrokenFile)
			}
			return item{}, false, err
		}
		if quote != 0 {
			b.WriteByte(c)
			if s.backslashEscapes && quote == '\'' && c == '\\' {
				// 转义符后面的字符原样消费，避免把 \" 里的引号当成字符串结尾
				escaped, err := s.readByte()
				if err != nil {
					return item{}, false, err
				}
				b.WriteByte(escaped)
				continue
			}
			if c == quote {
				next, err := s.peekByte()
				if err != nil {
					return item{}, false, err
				}
				if next == quote { // 连写两个表示字面量本身
					again, err := s.readByte()
					if err != nil {
						return item{}, false, err
					}
					b.WriteByte(again)
					continue
				}
				quote = 0
			}
			continue
		}
		switch c {
		case '\'', '"', '`':
			quote = c
			b.WriteByte(c)
		case ';':
			return item{text: strings.TrimSpace(b.String())}, true, nil
		default:
			b.WriteByte(c)
		}
	}
}

// 整个扫描过程按字节而不是按 rune 处理：结构字符全是 ASCII，而值里可能夹着非法 UTF-8 字节
// （拉丁编码残留、脏弹幕数据），按 rune 读会把它们悄悄替换成 U+FFFD，把数据改坏。
func (s *scanner) readByte() (byte, error) {
	return s.reader.ReadByte()
}

func (s *scanner) peekByte() (byte, error) {
	c, err := s.reader.ReadByte()
	if err != nil {
		return 0, err
	}
	if err := s.reader.UnreadByte(); err != nil {
		return 0, err
	}
	return c, nil
}

func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}

// Reader 顺序读取导出文件
type Reader struct {
	scanner *scanner
	dialect Dialect
	// headerRead 未读到文件头之前不接受数据行：方言决定字面量怎么解转义
	headerRead bool
	// pending 一条 INSERT 里的多行数据，逐行吐给调用方
	pending   [][]string
	pendingAt string
	pendingBy []string
}

// NewReader 创建导出文件读取器
func NewReader(r io.Reader) *Reader {
	return &Reader{scanner: newScanner(r)}
}

// Next 读取下一个事件，文件结束时返回 io.EOF
func (r *Reader) Next() (Event, error) {
	if len(r.pending) > 0 {
		values := r.pending[0]
		r.pending = r.pending[1:]
		return Event{Kind: EventRow, TableName: r.pendingAt, Columns: r.pendingBy, Values: values}, nil
	}
	for {
		it, ok, err := r.scanner.next()
		if err != nil {
			return Event{}, err
		}
		if !ok {
			return Event{}, io.EOF
		}
		if it.comment {
			event, matched, err := r.parseMarker(it.text)
			if err != nil {
				return Event{}, err
			}
			if matched {
				return event, nil
			}
			continue
		}
		// 只认 INSERT：建表、索引、方言前缀一律跳过，表结构由模型重建
		if !strings.HasPrefix(strings.ToUpper(it.text), "INSERT") {
			continue
		}
		if !r.headerRead {
			return Event{}, errMissingHeader
		}
		table, columns, rows, err := parseInsert(it.text, r.dialect)
		if err != nil {
			return Event{}, err
		}
		if len(rows) == 0 {
			continue
		}
		r.pendingAt, r.pendingBy = table, columns
		r.pending = rows[1:]
		return Event{Kind: EventRow, TableName: table, Columns: columns, Values: rows[0]}, nil
	}
}

// parseMarker 解析文件里的标记注释，matched 为 false 表示是普通注释
func (r *Reader) parseMarker(text string) (Event, bool, error) {
	switch {
	case strings.HasPrefix(text, headerMarker):
		if err := r.parseHeader(text); err != nil {
			return Event{}, false, err
		}
		return Event{Kind: EventHeader}, true, nil
	case strings.HasPrefix(text, tableMarker):
		if !r.headerRead {
			return Event{}, false, errMissingHeader
		}
		var meta TableMeta
		if err := decodeMarkerPayload(text, tableMarker, &meta); err != nil {
			return Event{}, false, err
		}
		return Event{Kind: EventTable, Table: meta}, true, nil
	case strings.HasPrefix(text, endMarker):
		var footer Footer
		if err := decodeMarkerPayload(text, endMarker, &footer); err != nil {
			return Event{}, false, err
		}
		return Event{Kind: EventFooter, Footer: footer}, true, nil
	}
	return Event{}, false, nil
}

// parseHeader 解析文件头，并据此确定方言
func (r *Reader) parseHeader(text string) error {
	var header Header
	if err := decodeMarkerPayload(text, headerMarker, &header); err != nil {
		return err
	}
	if header.Version != FileVersion {
		return fmt.Errorf("导出文件格式版本为 %d，本程序只支持 %d（请用同版本的 db export 重新导出）",
			header.Version, FileVersion)
	}
	dialect := Dialect{Name: header.Dialect}
	if !dialect.Valid() {
		return fmt.Errorf("导出文件声明的数据库类型无法识别: %q", header.Dialect)
	}
	r.dialect = dialect
	// MySQL 的字符串里反斜杠是转义符，解字面量时要按转义处理
	r.scanner.backslashEscapes = dialect.Name == DialectMySQL
	r.headerRead = true
	return nil
}

func decodeMarkerPayload(text, name string, dest any) error {
	payload := strings.TrimSpace(strings.TrimPrefix(text, name))
	if err := json.Unmarshal([]byte(payload), dest); err != nil {
		return fmt.Errorf("解析 %s 标记失败: %w", name, err)
	}
	return nil
}

// parser 解析一条语句，pos 是当前位置
type parser struct {
	text string
	pos  int
	// backslashEscapes 仅 MySQL 方言为真：它的字符串里反斜杠是转义符，
	// 判断引号是否成对时必须按转义处理，否则以反斜杠结尾的值会把字面量读串位
	backslashEscapes bool
}

// parseInsert 解析一条 INSERT 语句，返回表名、列名与各行原始字面量
func parseInsert(stmt string, d Dialect) (string, []string, [][]string, error) {
	p := &parser{text: stmt, backslashEscapes: d.Name == DialectMySQL}
	if !p.matchKeyword("INSERT") || !p.matchKeyword("INTO") {
		return "", nil, nil, fmt.Errorf("无法解析语句: %s", preview(stmt))
	}
	table, err := p.readIdentifier()
	if err != nil {
		return "", nil, nil, err
	}
	columns, err := p.readIdentifierList()
	if err != nil {
		return "", nil, nil, err
	}
	if !p.matchKeyword("VALUES") && !p.matchKeyword("VALUE") {
		return "", nil, nil, fmt.Errorf("INSERT 语句缺少 VALUES 关键字: %s", preview(stmt))
	}
	rows, err := p.readRows()
	if err != nil {
		return "", nil, nil, err
	}
	return table, columns, rows, nil
}

// readIdentifierList 读取括号内的列名清单
func (p *parser) readIdentifierList() ([]string, error) {
	if !p.consume('(') {
		return nil, fmt.Errorf("INSERT 语句缺少列清单: %s", preview(p.text))
	}
	var columns []string
	for {
		name, err := p.readIdentifier()
		if err != nil {
			return nil, err
		}
		columns = append(columns, name)
		p.skipSpace()
		if p.consume(',') {
			continue
		}
		if p.consume(')') {
			return columns, nil
		}
		return nil, fmt.Errorf("INSERT 语句列清单格式错误: %s", preview(p.text))
	}
}

// readRows 读取 VALUES 后面的全部元组
func (p *parser) readRows() ([][]string, error) {
	var rows [][]string
	for {
		values, err := p.readValueTuple()
		if err != nil {
			return nil, err
		}
		rows = append(rows, values)
		p.skipSpace()
		if p.consume(',') {
			continue
		}
		return rows, nil
	}
}

// readValueTuple 读取一个括号包裹的元组
func (p *parser) readValueTuple() ([]string, error) {
	if !p.consume('(') {
		return nil, fmt.Errorf("VALUES 格式错误: %s", preview(p.text))
	}
	var values []string
	for {
		value, err := p.readValue()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		p.skipSpace()
		if p.consume(',') {
			continue
		}
		if p.consume(')') {
			return values, nil
		}
		return nil, fmt.Errorf("VALUES 元组格式错误: %s", preview(p.text))
	}
}

// readValue 读取一个字面量，原样返回（含引号），不做任何解释
func (p *parser) readValue() (string, error) {
	p.skipSpace()
	if p.pos >= len(p.text) {
		return "", fmt.Errorf("VALUES 提前结束: %s", preview(p.text))
	}
	start := p.pos
	switch c := p.text[p.pos]; {
	case c == '\'':
		if err := p.skipQuoted('\''); err != nil {
			return "", err
		}
	case (c == 'X' || c == 'x') && p.pos+1 < len(p.text) && p.text[p.pos+1] == '\'':
		p.pos++
		if err := p.skipQuoted('\''); err != nil {
			return "", err
		}
	case p.matchFunctionCall():
		// 含 NUL 的字符串导出成了 CAST(X'..' AS TEXT)，整段当作一个字面量
		if err := p.skipFunctionCall(); err != nil {
			return "", err
		}
	default:
		for p.pos < len(p.text) && !isValueDelimiter(p.text[p.pos]) {
			p.pos++
		}
		if p.pos == start {
			return "", fmt.Errorf("VALUES 中出现无法识别的字面量: %s", preview(p.text[start:]))
		}
	}
	// PostgreSQL 的 '\x..'::bytea 这类带类型转换的字面量，把后缀一起带上
	if strings.HasPrefix(p.text[p.pos:], "::") {
		p.pos += 2
		for p.pos < len(p.text) && (isIdentChar(p.text[p.pos])) {
			p.pos++
		}
	}
	return p.text[start:p.pos], nil
}

// skipQuoted 跳过一个以 quote 包裹的字面量，内部同样遵循连写转义
func (p *parser) skipQuoted(quote byte) error {
	p.pos++ // 起始引号
	for p.pos < len(p.text) {
		c := p.text[p.pos]
		if c == '\\' && quote == '\'' && p.backslashEscapes {
			p.pos += 2 // 转义符连同后一个字符一起跳过
			continue
		}
		if c == quote {
			if p.pos+1 < len(p.text) && p.text[p.pos+1] == quote {
				p.pos += 2
				continue
			}
			p.pos++
			return nil
		}
		p.pos++
	}
	return fmt.Errorf("字面量缺少结束引号: %s", preview(p.text))
}

// functionCallPrefixes 会出现在取值位置上的函数调用，目前只有含 NUL 字符串的转换写法
var functionCallPrefixes = []string{"CAST(", "CONVERT_FROM("}

// matchFunctionCall 判断当前位置是否是函数调用
func (p *parser) matchFunctionCall() bool {
	for _, prefix := range functionCallPrefixes {
		if p.pos+len(prefix) <= len(p.text) && strings.EqualFold(p.text[p.pos:p.pos+len(prefix)], prefix) {
			return true
		}
	}
	return false
}

// skipFunctionCall 跳过一个完整的函数调用（含嵌套括号与字符串参数）
func (p *parser) skipFunctionCall() error {
	depth := 0
	for p.pos < len(p.text) {
		switch c := p.text[p.pos]; {
		case c == '\'':
			if err := p.skipQuoted('\''); err != nil {
				return err
			}
			continue
		case c == '(':
			depth++
		case c == ')':
			depth--
			if depth == 0 {
				p.pos++
				return nil
			}
		}
		p.pos++
	}
	return fmt.Errorf("函数调用缺少右括号: %s", preview(p.text))
}

// readIdentifier 读取标识符，支持反引号、双引号与裸标识符
func (p *parser) readIdentifier() (string, error) {
	p.skipSpace()
	if p.pos >= len(p.text) {
		return "", fmt.Errorf("语句提前结束: %s", preview(p.text))
	}
	switch c := p.text[p.pos]; c {
	case '`', '"':
		p.pos++
		start := p.pos
		for p.pos < len(p.text) {
			if p.text[p.pos] == c {
				if p.pos+1 < len(p.text) && p.text[p.pos+1] == c {
					p.pos += 2
					continue
				}
				name := p.text[start:p.pos]
				p.pos++
				return name, nil
			}
			p.pos++
		}
		return "", fmt.Errorf("标识符缺少结束引号: %s", preview(p.text))
	default:
		start := p.pos
		for p.pos < len(p.text) && isIdentChar(p.text[p.pos]) {
			p.pos++
		}
		if p.pos == start {
			return "", fmt.Errorf("无法识别的标识符: %s", preview(p.text[start:]))
		}
		return p.text[start:p.pos], nil
	}
}

// matchKeyword 匹配关键字（大小写不敏感），匹配后连同后面的空白一起跳过
func (p *parser) matchKeyword(keyword string) bool {
	p.skipSpace()
	if p.pos+len(keyword) > len(p.text) {
		return false
	}
	if !strings.EqualFold(p.text[p.pos:p.pos+len(keyword)], keyword) {
		return false
	}
	p.pos += len(keyword)
	return true
}

func (p *parser) skipSpace() {
	for p.pos < len(p.text) && (p.text[p.pos] == ' ' || p.text[p.pos] == '\t' ||
		p.text[p.pos] == '\n' || p.text[p.pos] == '\r') {
		p.pos++
	}
}

func (p *parser) consume(c byte) bool {
	p.skipSpace()
	if p.pos < len(p.text) && p.text[p.pos] == c {
		p.pos++
		return true
	}
	return false
}

func isValueDelimiter(c byte) bool {
	return c == ',' || c == ')' || c == '(' || c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isIdentChar(c byte) bool {
	return c == '_' || c == '$' ||
		(c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		c >= 0x80
}
