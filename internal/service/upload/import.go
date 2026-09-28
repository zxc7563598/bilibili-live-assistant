package upload

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/logger"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	// importFileLimit .gz 文件本身的大小上限
	importFileLimit int64 = 500 << 20
	// importMaxDecompressed 解压后允许读取的最大字节数
	importMaxDecompressed int64 = 10 << 30
	// importBatchSize 单次批量插入的行数
	importBatchSize = 1000
	// importTimeout 单次导入的整体超时
	importTimeout = 60 * time.Minute
	// importTaskRetention 已结束任务在内存中的保留时长
	importTaskRetention = 60 * time.Minute
	// importSource 导出文件 manifest.source 的取值
	importSource = "BilibiliDanmuji"
	// importBizType 导入流水写入的业务类型
	importBizType = "sync"
	// importRemark 导入流水写入的备注
	importRemark = "数据同步时直接导入"
	// 目标列宽上限，与 internal/model 的 varchar 声明一致
	importMaxUname    = 100
	importMaxMsg      = 200
	importMaxGiftName = 100
)

// MaxImportRequestSize 导入请求体上限：单文件上限 + 1MB（multipart 头部 / 边界开销）
const MaxImportRequestSize int64 = importFileLimit + 1<<20

// 导出文件中的表名
const (
	importTableDanmu = "bl_danmu_logs"
	importTableGift  = "bl_gift_records"
	importTableUser  = "bl_user_vips"
)

var (
	// errImportFormat 文件不是有效的导出数据文件
	errImportFormat = errors.New("导入文件格式不正确")
	// errImportSource 文件来源与导出功能不匹配
	errImportSource = errors.New("导入文件来源不匹配")
	// errImportDataExists 目标表已有数据
	errImportDataExists = errors.New("目标表已存在数据")
)

// importGuardGiftNames 判定大航海礼物的礼物名
var importGuardGiftNames = map[string]struct{}{
	"舰长": {},
	"提督": {},
	"总督": {},
}

// importUIDPrefix 匹配旧系统给 uid 拼上的前缀
var importUIDPrefix = regexp.MustCompile(`(?i)^uid\s*[:：]\s*`)

// importNumStr 兼容 JSON 中字符串与数字两种写法的数值列
type importNumStr string

// UnmarshalJSON 原样保留字符串，数字按原文保留
func (n *importNumStr) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*n = ""
		return nil
	}
	if data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*n = importNumStr(value)
		return nil
	}
	*n = importNumStr(data)
	return nil
}

// importManifest 导出文件的 manifest 节点
type importManifest struct {
	Source string `json:"source"`
}

// importDanmuRow 导出文件 bl_danmu_logs 的一行
type importDanmuRow struct {
	UID         importNumStr `json:"uid"`
	Uname       string       `json:"uname"`
	Msg         string       `json:"msg"`
	BadgeUID    importNumStr `json:"badge_uid"`
	BadgeUname  string       `json:"badge_uname"`
	BadgeRoomID importNumStr `json:"badge_room_id"`
	BadgeName   string       `json:"badge_name"`
	BadgeLevel  int64        `json:"badge_level"`
	BadgeType   int64        `json:"badge_type"`
	SendAt      int64        `json:"send_at"`
	CreatedAt   int64        `json:"created_at"`
	UpdatedAt   int64        `json:"updated_at"`
	DeletedAt   *int64       `json:"deleted_at"`
}

// importGiftRow 导出文件 bl_gift_records 的一行
type importGiftRow struct {
	UID              importNumStr `json:"uid"`
	Uname            string       `json:"uname"`
	GiftID           importNumStr `json:"gift_id"`
	GiftName         string       `json:"gift_name"`
	Price            importNumStr `json:"price"`
	Num              int64        `json:"num"`
	Original         int64        `json:"original"`
	OriginalGiftName string       `json:"original_gift_name"`
	OriginalPrice    importNumStr `json:"original_price"`
	CreatedAt        int64        `json:"created_at"`
	UpdatedAt        int64        `json:"updated_at"`
	DeletedAt        *int64       `json:"deleted_at"`
}

// importUserRow 导出文件 bl_user_vips 的一行
type importUserRow struct {
	UID             importNumStr `json:"uid"`
	Name            string       `json:"name"`
	Point           int64        `json:"point"`
	Coin            int64        `json:"coin"`
	TotalDanmuCount int64        `json:"total_danmu_count"`
	TotalGiftAmount importNumStr `json:"total_gift_amount"`
	CreatedAt       int64        `json:"created_at"`
	UpdatedAt       int64        `json:"updated_at"`
	DeletedAt       *int64       `json:"deleted_at"`
}

// importCounter 统计已读取的压缩字节数
type importCounter struct {
	reader io.Reader
	n      int64
}

// Read 透传读取并累计字节数
func (c *importCounter) Read(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	c.n += int64(n)
	return n, err
}

// importRun 单次导入的执行状态
type importRun struct {
	svc         *Service
	taskID      string
	counter     importCounter
	seenUID     map[int64]struct{}
	creditLog   int64
	skipped     int64
	truncated   int64
	badAmount   int64
	softDeleted int64
}

// importFile 流式解析 .gz 导出文件并分批入库
func (r *importRun) importFile(ctx context.Context, tx *gorm.DB, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开导入文件失败: %w", err)
	}
	defer file.Close()
	r.counter.reader = file
	reader, err := gzip.NewReader(&r.counter)
	if err != nil {
		return fmt.Errorf("%w: 解压失败: %v", errImportFormat, err)
	}
	defer reader.Close()
	decoder := json.NewDecoder(io.LimitReader(reader, importMaxDecompressed))
	if err := expectDelim(decoder, '{'); err != nil {
		return err
	}
	var hasManifest, hasTables bool
	for decoder.More() {
		key, err := readKey(decoder)
		if err != nil {
			return err
		}
		switch key {
		case "manifest":
			var manifest importManifest
			if err := decoder.Decode(&manifest); err != nil {
				return fmt.Errorf("%w: manifest 解析失败: %v", errImportFormat, err)
			}
			if manifest.Source != importSource {
				return fmt.Errorf("%w: source=%q", errImportSource, manifest.Source)
			}
			hasManifest = true
		case "tables":
			if err := r.importTablesNode(ctx, tx, decoder); err != nil {
				return err
			}
			hasTables = true
		default:
			// summary 等其余节点不参与导入
			if err := discardValue(decoder); err != nil {
				return fmt.Errorf("%w: %s 解析失败: %v", errImportFormat, key, err)
			}
		}
	}
	if err := expectDelim(decoder, '}'); err != nil {
		return err
	}
	if !hasManifest || !hasTables {
		return fmt.Errorf("%w: 缺少 manifest 或 tables 节点", errImportFormat)
	}
	return nil
}

// importTablesNode 逐个表数组处理，未知表整体跳过
func (r *importRun) importTablesNode(ctx context.Context, tx *gorm.DB, decoder *json.Decoder) error {
	if err := expectDelim(decoder, '{'); err != nil {
		return err
	}
	for decoder.More() {
		table, err := readKey(decoder)
		if err != nil {
			return err
		}
		switch table {
		case importTableDanmu:
			err = r.danmuRows(ctx, tx, decoder)
		case importTableGift:
			err = r.giftRows(ctx, tx, decoder)
		case importTableUser:
			err = r.userRows(ctx, tx, decoder)
		default:
			err = discardValue(decoder)
		}
		if err != nil {
			return err
		}
	}
	return expectDelim(decoder, '}')
}

// danmuRows 导入弹幕表
func (r *importRun) danmuRows(ctx context.Context, tx *gorm.DB, decoder *json.Decoder) error {
	var imported int64
	return importArray(decoder, func(batch []importDanmuRow) error {
		entities := make([]model.LiveDanmu, 0, len(batch))
		for _, row := range batch {
			entities = append(entities, r.danmuEntity(row))
		}
		if err := r.svc.liveDanmuRepo.CreateBatch(ctx, tx, entities); err != nil {
			return fmt.Errorf("写入弹幕数据失败: %w", err)
		}
		imported += int64(len(entities))
		r.reportProgress(importTableDanmu, imported)
		return nil
	})
}

// giftRows 导入礼物表
func (r *importRun) giftRows(ctx context.Context, tx *gorm.DB, decoder *json.Decoder) error {
	var imported int64
	return importArray(decoder, func(batch []importGiftRow) error {
		entities := make([]model.LiveGift, 0, len(batch))
		for _, row := range batch {
			entities = append(entities, r.giftEntity(row))
		}
		if err := r.svc.liveGiftRepo.CreateBatch(ctx, tx, entities); err != nil {
			return fmt.Errorf("写入礼物数据失败: %w", err)
		}
		imported += int64(len(entities))
		r.reportProgress(importTableGift, imported)
		return nil
	})
}

// userRows 导入用户表，并在同一事务内补录积分与星光流水
func (r *importRun) userRows(ctx context.Context, tx *gorm.DB, decoder *json.Decoder) error {
	var imported int64
	return importArray(decoder, func(batch []importUserRow) error {
		entities := make([]model.LiveUser, 0, len(batch))
		kept := make([]importUserRow, 0, len(batch))
		uids := make([]int64, 0, len(batch))
		for _, row := range batch {
			uid, ok := parseImportUID(string(row.UID))
			if !ok {
				r.skipped++
				continue
			}
			// uid 有唯一索引，同一 uid 只保留导出行序最靠前的一条
			if _, dup := r.seenUID[uid]; dup {
				r.skipped++
				continue
			}
			r.seenUID[uid] = struct{}{}
			entities = append(entities, r.userEntity(uid, row))
			kept = append(kept, row)
			uids = append(uids, uid)
		}
		if err := r.svc.liveUserRepo.CreateBatch(ctx, tx, entities); err != nil {
			return fmt.Errorf("写入用户数据失败: %w", err)
		}
		if err := r.applyCredit(ctx, tx, kept); err != nil {
			return err
		}
		imported += int64(len(entities))
		r.reportProgress(importTableUser, imported)
		return nil
	})
}

// applyCredit 回查主键后给本次真正入库的用户补录积分与星光流水
func (r *importRun) applyCredit(ctx context.Context, tx *gorm.DB, rows []importUserRow) error {
	if len(rows) == 0 {
		return nil
	}
	uids := make([]int64, 0, len(rows))
	for _, row := range rows {
		uid, ok := parseImportUID(string(row.UID))
		if !ok {
			continue
		}
		uids = append(uids, uid)
	}
	users, err := r.svc.liveUserRepo.ListByUIDs(ctx, tx, uids)
	if err != nil {
		return fmt.Errorf("回查导入用户失败: %w", err)
	}
	userIDs := make(map[int64]int64, len(users))
	for _, user := range users {
		userIDs[user.UID] = user.ID
	}
	for _, row := range rows {
		uid, _ := parseImportUID(string(row.UID))
		userID, ok := userIDs[uid]
		if !ok {
			continue
		}
		// 软删除用户无法参与资产变更
		if row.DeletedAt != nil {
			if row.Point != 0 || row.Coin != 0 {
				r.softDeleted++
			}
			continue
		}
		for _, item := range [2]struct {
			creditType enum.CreditType
			amount     int64
		}{
			{enum.CreditTypePoints, row.Point},
			{enum.CreditTypeStars, row.Coin},
		} {
			if item.amount <= 0 {
				continue
			}
			err := r.svc.liveUserSvc.AdjustCredit(ctx, tx, item.creditType, liveuser.AdjustCreditParams{
				UserID:       userID,
				ChangeType:   enum.ChangeTypeIncrease,
				ChangeAmount: item.amount,
				BizType:      importBizType,
				Remark:       importRemark,
				OperatorType: enum.OperatorTypeSystem,
				OperatorID:   0,
			})
			if err != nil {
				return fmt.Errorf("补录用户 %d 资产流水失败: %w", uid, err)
			}
			r.creditLog++
		}
	}
	return nil
}

// danmuEntity 导出文件的一行转弹幕模型，房间与直播场次固定为 0
func (r *importRun) danmuEntity(row importDanmuRow) model.LiveDanmu {
	uid, _ := parseImportUID(string(row.UID))
	badgeUID, _ := parseImportInt(string(row.BadgeUID))
	badgeRoomID, _ := parseImportInt(string(row.BadgeRoomID))
	uname, cutUname := truncateRunes(row.Uname, importMaxUname)
	msg, cutMsg := truncateRunes(row.Msg, importMaxMsg)
	badgeUname, cutBadgeUname := truncateRunes(row.BadgeUname, importMaxUname)
	badgeName, cutBadgeName := truncateRunes(row.BadgeName, importMaxUname)
	if cutUname || cutMsg || cutBadgeUname || cutBadgeName {
		r.truncated++
	}
	return model.LiveDanmu{
		UID:         uid,
		Uname:       uname,
		Msg:         msg,
		BadgeUID:    badgeUID,
		BadgeUname:  badgeUname,
		BadgeRoomID: badgeRoomID,
		BadgeName:   badgeName,
		BadgeLevel:  row.BadgeLevel,
		BadgeType:   enum.BadgeType(row.BadgeType),
		SendAt:      row.SendAt,
		BaseModel:   toBaseModel(row.CreatedAt, row.UpdatedAt, row.DeletedAt),
	}
}

// giftEntity 导出文件的一行转礼物模型，大航海礼物标记为舰长类型
func (r *importRun) giftEntity(row importGiftRow) model.LiveGift {
	uid, _ := parseImportUID(string(row.UID))
	giftID, _ := parseImportInt(string(row.GiftID))
	giftType := enum.GiftTypeNormal
	if _, ok := importGuardGiftNames[row.GiftName]; ok {
		giftType = enum.GiftTypeGuard
	}
	uname, cutUname := truncateRunes(row.Uname, importMaxUname)
	giftName, cutGiftName := truncateRunes(row.GiftName, importMaxGiftName)
	originalGiftName, cutOriginalName := truncateRunes(row.OriginalGiftName, importMaxGiftName)
	if cutUname || cutGiftName || cutOriginalName {
		r.truncated++
	}
	return model.LiveGift{
		UID:               uid,
		Uname:             uname,
		GiftType:          giftType,
		GiftID:            giftID,
		GiftName:          giftName,
		Price:             r.cents(row.Price),
		Num:               row.Num,
		SendAt:            row.CreatedAt,
		Original:          enum.YesNo(row.Original),
		OriginalGiftName:  originalGiftName,
		OriginalGiftPrice: r.cents(row.OriginalPrice),
		BaseModel:         toBaseModel(row.CreatedAt, row.UpdatedAt, row.DeletedAt),
	}
}

// userEntity 导出文件的一行转用户模型，积分与星光改由流水补录
func (r *importRun) userEntity(uid int64, row importUserRow) model.LiveUser {
	uname, cut := truncateRunes(row.Name, importMaxUname)
	if cut {
		r.truncated++
	}
	return model.LiveUser{
		UID:             uid,
		Uname:           uname,
		TotalDanmuCount: row.TotalDanmuCount,
		TotalGiftAmount: r.cents(row.TotalGiftAmount),
		Enable:          enum.EnableEnable,
		BaseModel:       toBaseModel(row.CreatedAt, row.UpdatedAt, row.DeletedAt),
	}
}

// cents 元字符串转分，非法值按 0 处理并计数
func (r *importRun) cents(raw importNumStr) int64 {
	value, ok := parseImportCents(string(raw))
	if !ok {
		r.badAmount++
	}
	return value
}

// reportProgress 回写当前表已导入行数、已读取字节数与流水统计
func (r *importRun) reportProgress(table string, imported int64) {
	bytes := r.counter.n
	r.svc.importTasks.update(r.taskID, func(task *importTask) {
		task.currentTable = table
		task.bytesProcessed = bytes
		switch table {
		case importTableDanmu:
			task.danmuCount = imported
		case importTableGift:
			task.giftCount = imported
		case importTableUser:
			task.userCount = imported
			task.creditLogCount = r.creditLog
			task.skippedUserCount = r.skipped
		}
	})
}

// importArray 逐个解码 JSON 数组元素，攒够一批交给 flush
func importArray[T any](decoder *json.Decoder, flush func(batch []T) error) error {
	if err := expectDelim(decoder, '['); err != nil {
		return err
	}
	batch := make([]T, 0, importBatchSize)
	for decoder.More() {
		var row T
		if err := decoder.Decode(&row); err != nil {
			return fmt.Errorf("%w: 数据行解析失败: %v", errImportFormat, err)
		}
		batch = append(batch, row)
		if len(batch) >= importBatchSize {
			if err := flush(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	if err := expectDelim(decoder, ']'); err != nil {
		return err
	}
	if len(batch) == 0 {
		return nil
	}
	return flush(batch)
}

// checkTablesEmpty 校验三张目标表均无数据（含软删除行）
func (s *Service) checkTablesEmpty(ctx context.Context, tx *gorm.DB) error {
	counts := []struct {
		table string
		count func() (int64, error)
	}{
		{"live_danmus", func() (int64, error) { return s.liveDanmuRepo.CountUnscoped(ctx, tx) }},
		{"live_gifts", func() (int64, error) { return s.liveGiftRepo.CountUnscoped(ctx, tx) }},
		{"live_users", func() (int64, error) { return s.liveUserRepo.CountUnscoped(ctx, tx) }},
	}
	for _, item := range counts {
		total, err := item.count()
		if err != nil {
			return fmt.Errorf("统计 %s 行数失败: %w", item.table, err)
		}
		if total > 0 {
			return fmt.Errorf("%w: %s 已存在 %d 行", errImportDataExists, item.table, total)
		}
	}
	return nil
}

// parseImportUID 归一化 uid：去空格与前缀，非纯数字串视为无效
func parseImportUID(raw string) (int64, bool) {
	value := strings.TrimSpace(importUIDPrefix.ReplaceAllString(strings.TrimSpace(raw), ""))
	if value == "" {
		return 0, false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return 0, false
		}
	}
	uid, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false
	}
	return uid, true
}

// parseImportInt 解析整数字符串，空值与非数字串返回 0
func parseImportInt(raw string) (int64, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, true
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

// parseImportCents 十进制元字符串转分，小数位截断到两位
func parseImportCents(raw string) (int64, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, true
	}
	negative := strings.HasPrefix(value, "-")
	if negative {
		value = value[1:]
	}
	integer, fraction, _ := strings.Cut(value, ".")
	if len(fraction) > 2 {
		fraction = fraction[:2]
	}
	fraction += strings.Repeat("0", 2-len(fraction))
	cents, err := strconv.ParseInt(integer+fraction, 10, 64)
	if err != nil {
		return 0, false
	}
	if negative {
		return -cents, true
	}
	return cents, true
}

// truncateRunes 按字符截断到列宽上限，返回是否发生截断
func truncateRunes(value string, max int) (string, bool) {
	if utf8.RuneCountInString(value) <= max {
		return value, false
	}
	return string([]rune(value)[:max]), true
}

// toBaseModel 按导出时间戳构造基础模型，删除时间戳为 null 表示未删除
func toBaseModel(createdAt, updatedAt int64, deletedAt *int64) model.BaseModel {
	base := model.BaseModel{
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	if deletedAt != nil && *deletedAt > 0 {
		base.DeletedAt = gorm.DeletedAt{Time: time.Unix(*deletedAt, 0), Valid: true}
	}
	return base
}

// discardValue 流式跳过当前值，不把它读进内存
func discardValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("%w: 读取失败: %v", errImportFormat, err)
	}
	if delim, ok := token.(json.Delim); !ok || (delim != '{' && delim != '[') {
		return nil
	}
	for depth := 1; depth > 0; {
		token, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("%w: 读取失败: %v", errImportFormat, err)
		}
		if delim, ok := token.(json.Delim); ok {
			switch delim {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
			}
		}
	}
	return nil
}

// expectDelim 读一个定界符并断言与期望一致
func expectDelim(decoder *json.Decoder, want json.Delim) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("%w: 读取失败: %v", errImportFormat, err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != want {
		return fmt.Errorf("%w: 期望 %q", errImportFormat, want)
	}
	return nil
}

// readKey 读一个对象键
func readKey(decoder *json.Decoder) (string, error) {
	token, err := decoder.Token()
	if err != nil {
		return "", fmt.Errorf("%w: 读取失败: %v", errImportFormat, err)
	}
	key, ok := token.(string)
	if !ok {
		return "", fmt.Errorf("%w: 对象结构不正确", errImportFormat)
	}
	return key, nil
}

// copyUploadedFile 把上传文件落到本地临时文件
func copyUploadedFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	dst, err := os.CreateTemp("", "bililive-import-*.json.gz")
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(dst.Name())
		return "", err
	}
	return dst.Name(), nil
}

// checkImportPrecondition 导入前的空表预检
func (s *Service) checkImportPrecondition(ctx context.Context) (int, error) {
	if err := s.checkTablesEmpty(ctx, nil); err != nil {
		if errors.Is(err, errImportDataExists) {
			return CodeDataExists, err
		}
		return CodeImportFailed, err
	}
	return 0, nil
}

// importErrCode 导入失败原因映射成错误码
func importErrCode(err error) int {
	switch {
	case errors.Is(err, errImportDataExists):
		return CodeDataExists
	case errors.Is(err, errImportSource):
		return CodeImportSourceInvalid
	case errors.Is(err, errImportFormat):
		return CodeImportFileInvalid
	default:
		return CodeImportFailed
	}
}

// runImport 后台执行导入，失败整体回滚
func (s *Service) runImport(ctx context.Context, taskID, path string) {
	// 临时文件只服务于本次导入
	defer os.Remove(path)
	// 协程内的 panic 不经过 gin 的 Recovery
	defer func() {
		if rec := recover(); rec != nil {
			logger.UploadLogger.Error("数据导入协程 panic", zap.String("task_id", taskID), zap.Any("panic", rec))
			s.importTasks.update(taskID, func(task *importTask) {
				task.status = enum.ImportStatusFailed
				task.errorCode = CodeImportFailed
				task.finishedAt = time.Now().Unix()
			})
		}
	}()
	// 请求上下文在响应返回后立即取消，导入必须用独立上下文
	ctx, cancel := context.WithTimeout(ctx, importTimeout)
	defer cancel()
	s.importTasks.update(taskID, func(task *importTask) {
		task.status = enum.ImportStatusRunning
	})
	run := &importRun{svc: s, taskID: taskID, seenUID: make(map[int64]struct{})}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.checkTablesEmpty(ctx, tx); err != nil {
			return err
		}
		return run.importFile(ctx, tx, path)
	})
	task := s.importTasks.get(taskID)
	if err != nil {
		errCode := importErrCode(err)
		logger.UploadLogger.Error("数据导入失败",
			zap.String("task_id", taskID),
			zap.Int64("admin_id", task.adminID),
			zap.String("filename", task.filename),
			zap.Int("code", errCode),
			zap.Error(err),
		)
		s.importTasks.update(taskID, func(task *importTask) {
			task.status = enum.ImportStatusFailed
			task.errorCode = errCode
			task.finishedAt = time.Now().Unix()
		})
		return
	}
	logger.UploadLogger.Info("数据导入完成",
		zap.String("task_id", taskID),
		zap.Int64("admin_id", task.adminID),
		zap.String("filename", task.filename),
		zap.Int64("danmu", task.danmuCount),
		zap.Int64("gift", task.giftCount),
		zap.Int64("user", task.userCount),
		zap.Int64("credit_log", task.creditLogCount),
		zap.Int64("skipped_user", task.skippedUserCount),
		zap.Int64("truncated", run.truncated),
		zap.Int64("bad_amount", run.badAmount),
		zap.Int64("soft_deleted_credit", run.softDeleted),
		zap.Int64("duration", time.Now().Unix()-task.startedAt),
	)
	s.importTasks.update(taskID, func(task *importTask) {
		task.status = enum.ImportStatusFinished
		task.finishedAt = time.Now().Unix()
	})
}
