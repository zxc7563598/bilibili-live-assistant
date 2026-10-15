package live

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_guard_record"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_blacklist"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/util"
)

// giftMergeWindow 礼物合并 debounce 窗口：同一用户在此时间内的多个礼物会被合并为一条答谢消息
const giftMergeWindow = 3 * time.Second

// unmuteFailLimit 解禁失败次数阈值，累计达到该次数后标记为解禁失败
const unmuteFailLimit int64 = 3

// giftMergeBuffer 礼物合并缓冲区，在 MergeGift 开启时将同一用户短时间内赠送的
type giftMergeBuffer struct {
	mu       sync.Mutex
	entries  map[int64]*mergeEntry
	interval time.Duration
}

type mergeEntry struct {
	infos []*giftThankInfo
	timer *time.Timer
}

func newGiftMergeBuffer(interval time.Duration) *giftMergeBuffer {
	return &giftMergeBuffer{
		entries:  make(map[int64]*mergeEntry),
		interval: interval,
	}
}

// giftProcessor 处理礼物相关消息（SEND_GIFT / SEND_GIFT_V2 / GUARD_BUY / SUPER_CHAT_MESSAGE）
type giftProcessor struct {
	liveUserSvc           *liveuser.Service
	liveGiftRepo          live_gift.Repository
	liveUserBlacklistRepo live_user_blacklist.Repository
	liveGuardRecordRepo   live_guard_record.Repository
	roomState             *RoomState
	configCache           *robotconfig.Cache
	client                *bilibili.Client
	getBotUID             func() int64
	enqueueDanmu          func(msg string, kind string)
	mergeBuffer           *giftMergeBuffer
}

// giftThankInfo 礼物答谢所需的共有信息，从 SendGiftInfo / GuardBuyInfo / SuperChatMessage 中提取
type giftThankInfo struct {
	UID               int64          // 用户ID
	Uname             string         // 用户名
	GiftID            int64          // 礼物ID
	GiftName          string         // 礼物名称
	GiftType          enum.GiftType  // 礼物类型
	Price             int64          // 价格（分）
	Num               int64          // 数量
	Original          enum.YesNo     // 是否原始礼物
	OriginalGiftPrice int64          // 原始礼物价格
	Badge             bool           // 是否是本直播间牌子
	BadgeType         enum.BadgeType // 牌子类型
}

func newGiftProcessor(liveUserSvc *liveuser.Service, liveGiftRepo live_gift.Repository, liveUserBlacklistRepo live_user_blacklist.Repository, liveGuardRecordRepo live_guard_record.Repository, roomState *RoomState, configCache *robotconfig.Cache, client *bilibili.Client, getBotUID func() int64, enqueueDanmu func(msg string, kind string)) *giftProcessor {
	return &giftProcessor{
		liveUserSvc:           liveUserSvc,
		liveGiftRepo:          liveGiftRepo,
		liveUserBlacklistRepo: liveUserBlacklistRepo,
		liveGuardRecordRepo:   liveGuardRecordRepo,
		roomState:             roomState,
		configCache:           configCache,
		client:                client,
		getBotUID:             getBotUID,
		enqueueDanmu:          enqueueDanmu,
		mergeBuffer:           newGiftMergeBuffer(giftMergeWindow),
	}
}

func (p *giftProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdSendGift, live.CmdSendGiftV2, live.CmdGuardBuy, live.CmdSuperDanmuMsg}
}

func (p *giftProcessor) Process(ctx context.Context, cmd live.Cmd, data any, roomID int64) error {
	switch cmd {
	case live.CmdSendGift, live.CmdSendGiftV2:
		info, ok := data.(*live.SendGiftInfo)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.SendGiftInfo，实际 %T", data)
			return nil
		}
		// 一次广播可能携带多份礼物（盲盒爆出时可达数十种），每一份都当作独立的礼物推送处理
		pushes := make([]giftPush, 0, len(info.Gifts))
		for _, item := range info.Gifts {
			gift, thankInfo := buildSendGift(roomID, info, item)
			pushes = append(pushes, giftPush{gift: gift, thank: thankInfo})
		}
		p.processGifts(ctx, roomID, pushes)
	case live.CmdGuardBuy:
		info, ok := data.(*live.GuardBuyInfo)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.GuardBuyInfo，实际 %T", data)
			return nil
		}
		gift := &model.LiveGift{
			RoomID:   roomID,
			UID:      info.UID,
			Uname:    info.Uname,
			GiftType: enum.GiftTypeGuard,
			GiftID:   info.GiftID,
			GiftName: info.GiftName,
			Price:    info.Price,
			Num:      info.Num,
			LiveID:   0,
			SendAt:   time.Now().Unix(),
			Original: enum.Yes,
		}
		thankInfo := &giftThankInfo{
			UID:       info.UID,
			Uname:     info.Uname,
			GiftID:    info.GiftID,
			GiftName:  info.GiftName,
			GiftType:  enum.GiftTypeGuard,
			Price:     info.Price,
			Num:       1,
			Original:  enum.Yes,
			Badge:     true,
			BadgeType: enum.BadgeType(info.GuardLevel),
		}
		thankInfo.OriginalGiftPrice = info.Price
		p.recordGuardExpire(ctx, roomID, info)
		p.processGifts(ctx, roomID, []giftPush{{gift: gift, thank: thankInfo}})
	case live.CmdSuperDanmuMsg:
		info, ok := data.(*live.SuperChatMessage)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.SuperChatMessage，实际 %T", data)
			return nil
		}
		gift := &model.LiveGift{
			RoomID:     roomID,
			UID:        info.UID,
			Uname:      info.Uname,
			GiftType:   enum.GiftTypeSuperChat,
			GiftID:     info.GiftID,
			GiftName:   info.GiftName,
			Price:      info.Price,
			Num:        1,
			Message:    info.Message,
			BadgeName:  info.BadgeName,
			BadgeLevel: info.BadgeLevel,
			BadgeType:  enum.BadgeType(info.BadgeType),
			LiveID:     0,
			SendAt:     time.Now().Unix(),
			Original:   enum.Yes,
		}
		thankInfo := &giftThankInfo{
			UID:       info.UID,
			Uname:     info.Uname,
			GiftID:    info.GiftID,
			GiftName:  info.GiftName,
			GiftType:  enum.GiftTypeSuperChat,
			Price:     info.Price,
			Num:       1,
			Original:  enum.Yes,
			Badge:     info.BadgeRoomID == roomID,
			BadgeType: enum.BadgeType(info.BadgeType),
		}
		thankInfo.OriginalGiftPrice = info.Price
		p.processGifts(ctx, roomID, []giftPush{{gift: gift, thank: thankInfo}})
	}
	return nil
}

// buildSendGift 把 SEND_GIFT / SEND_GIFT_V2 中的一份礼物明细组装成入库记录与答谢信息
//
// 盲盒爆出时送礼方与勋章信息取自整条广播，标价与数量取自该份明细；
// 原始礼物信息（OriginalGiftID/Name/Price）取自整条广播的 blind_gift。
func buildSendGift(roomID int64, info *live.SendGiftInfo, item live.GiftItemInfo) (*model.LiveGift, *giftThankInfo) {
	gift := &model.LiveGift{
		RoomID:     roomID,
		UID:        info.UID,
		Uname:      info.Uname,
		GiftType:   enum.GiftTypeNormal,
		GiftID:     item.GiftID,
		GiftName:   item.GiftName,
		Price:      item.Price,
		Num:        item.Num,
		BadgeUID:   info.BadgeUID,
		BadgeName:  info.BadgeName,
		BadgeLevel: info.BadgeLevel,
		BadgeType:  enum.BadgeType(info.BadgeType),
		LiveID:     0,
		SendAt:     time.Now().Unix(),
		Original:   enum.Yes,
	}
	if info.BlindGift != nil {
		gift.Original = enum.No
		gift.OriginalGiftID = info.BlindGift.OriginalGiftID
		gift.OriginalGiftName = info.BlindGift.OriginalGiftName
		gift.OriginalGiftPrice = info.BlindGift.OriginalGiftPrice
	}
	thankInfo := &giftThankInfo{
		UID:       info.UID,
		Uname:     info.Uname,
		GiftID:    item.GiftID,
		GiftName:  item.GiftName,
		GiftType:  enum.GiftTypeNormal,
		Price:     item.Price,
		Num:       item.Num,
		Original:  enum.Yes,
		Badge:     info.BadgeUID == info.AnchorID,
		BadgeType: enum.BadgeType(info.BadgeType),
	}
	thankInfo.OriginalGiftPrice = item.Price
	if info.BlindGift != nil && info.BlindGift.OriginalGiftPrice > 0 {
		thankInfo.Original = enum.No
		thankInfo.OriginalGiftPrice = info.BlindGift.OriginalGiftPrice
	}
	return gift, thankInfo
}

// giftPush 一次推送中的一份礼物：待入库的记录与对应的答谢信息
type giftPush struct {
	gift  *model.LiveGift
	thank *giftThankInfo
}

// recordGuardExpire 延长用户的大航海到期时间，并写入开通台账
//
// 到期时间与台账各管一摊、口径不同，都要写但互不推导：
// live_users 的到期列是「身份有效到什么时候」（30 天 × 份数），台账的月份区间是
// 「哪几个月的名单里有他」（月份 × 份数）。7 月 29 日开通一份，有效期到 8 月 28 日，
// 但只在 7 月名单里。
func (p *giftProcessor) recordGuardExpire(ctx context.Context, roomID int64, info *live.GuardBuyInfo) {
	level := enum.BadgeType(info.GuardLevel)
	now := time.Now()
	if _, errCode, err := p.liveUserSvc.ExtendGuardExpire(ctx, info.UID, info.Uname, level, now, info.Num); errCode != 0 {
		log.Printf("[live.Gift] 记录大航海到期时间失败: uid=%d guardLevel=%d code=%d err=%v", info.UID, info.GuardLevel, errCode, err)
		return
	}
	// 顺延基准取该用户该档位最晚的一条台账：live_users 的到期时间是天数口径，替代不了它
	latest, err := p.liveGuardRecordRepo.LatestByUIDAndLevel(ctx, nil, info.UID, level)
	if err != nil {
		log.Printf("[live.Gift] 查询大航海顺延基准失败: uid=%d guardLevel=%d err=%v", info.UID, info.GuardLevel, err)
		return
	}
	var prevCoverTo int64
	if latest != nil {
		prevCoverTo = latest.CoverTo
	}
	coverFrom, coverTo := liveuser.GuardCoverWindow(prevCoverTo, now, info.Num)
	if _, err := p.liveGuardRecordRepo.Create(ctx, nil, &model.LiveGuardRecord{
		RoomID:     roomID,
		UID:        info.UID,
		Uname:      info.Uname,
		GuardLevel: level,
		GiftID:     info.GiftID,
		Price:      info.Price,
		Num:        info.Num,
		OpenAt:     now.Unix(),
		CoverFrom:  coverFrom,
		CoverTo:    coverTo,
	}); err != nil {
		log.Printf("[live.Gift] 记录大航海开通台账失败: uid=%d guardLevel=%d err=%v", info.UID, info.GuardLevel, err)
	}
}

// processGifts 处理一次礼物推送（普通送礼只有 1 份，盲盒爆出时可达数十份）
//
// 入库走批量接口，拼成单条多行 INSERT，避免写一半失败时落库数据永久缺斤少两；
// 入库、累计金额属于同一用户的一次广播，只做一遍，答谢等后续事件再逐份触发。
func (p *giftProcessor) processGifts(ctx context.Context, roomID int64, pushes []giftPush) {
	if len(pushes) == 0 {
		return
	}
	// 注册用户 —— 同一广播内的多份礼物属于同一用户，只需注册一次
	userID, err := p.liveUserSvc.EnsureUser(ctx, pushes[0].thank.UID, pushes[0].thank.Uname)
	if err != nil {
		log.Printf("[live.Gift] 注册并获取用户信息失败: %v", err)
		return
	}
	entities := make([]model.LiveGift, 0, len(pushes))
	var totalAmount int64
	for _, push := range pushes {
		entities = append(entities, *push.gift)
		totalAmount += push.gift.Price * push.gift.Num
	}
	if err := p.liveGiftRepo.CreateBatch(ctx, nil, entities); err != nil {
		log.Printf("[live.Gift] 礼物存储失败: %v", err)
		return
	}
	// 追加用户累计赠送礼物金额
	if userID > 0 {
		if err := p.liveUserSvc.AddTotalGiftAmount(ctx, userID, totalAmount); err != nil {
			log.Printf("[live.Gift] 追加用户累计赠送礼物金额失败: %v", err)
			return
		}
	}
	// 处理后续事件
	botUID := p.getBotUID()
	liveStatus := p.roomState.LiveStatus()
	for _, push := range pushes {
		// 礼物答谢
		p.processGiftIn(push.thank, roomID, botUID, liveStatus)
		// 黑名单赎回
		p.processRedeem(ctx, push.thank.UID, push.thank.Price, push.thank.Num, roomID, botUID)
		// 奖励发放
		p.processReward(ctx, userID, push.thank, botUID)
	}
}

// processGiftIn 处理礼物答谢逻辑
func (p *giftProcessor) processGiftIn(info *giftThankInfo, roomID, botUID int64, liveStatus int) {
	if botUID == info.UID {
		return
	}
	if !configEnabled(p.configCache, "gift") {
		return
	}
	var giftCfg robotconfig.GiftConfig
	if err := p.configCache.UnmarshalGroup("gift", &giftCfg); err != nil {
		log.Printf("[live.Gift] 加载礼物答谢配置失败: %v", err)
		return
	}
	if !p.checkGiftCondition(giftCfg, info, roomID, liveStatus) {
		return
	}
	if ptr.ParseBool(giftCfg.MergeGift) {
		p.mergeBuffer.add(info, p.sendGiftReply)
	} else {
		p.sendGiftReply([]*giftThankInfo{info})
	}
}

// checkGiftCondition 校验礼物答谢的 Requirement, Scene 以及 MinBattery 是否满足
func (p *giftProcessor) checkGiftCondition(giftCfg robotconfig.GiftConfig, info *giftThankInfo, roomID int64, liveStatus int) bool {
	req := ptr.ParseEnumInt[enum.Requirement](giftCfg.Requirement)
	switch req {
	case enum.RequirementUnlimited:
	case enum.RequirementHasBadge:
		if !info.Badge {
			log.Printf("[live.Gift] 答谢条件不满足: 需要携带当前直播间牌子")
			return false
		}
	case enum.RequirementHasSailBadge:
		if !info.Badge {
			log.Printf("[live.Gift] 答谢条件不满足: 需要携带当前直播间牌子")
			return false
		}
		if enum.BadgeType(info.BadgeType) == enum.BadgeTypeL0 {
			log.Printf("[live.Gift] 答谢条件不满足: 需要非 L0 勋章")
			return false
		}
	default:
		log.Printf("[live.Gift] 未知的答谢条件: %d", req)
		return false
	}
	sce := ptr.ParseEnumInt[enum.Scene](giftCfg.Scene)
	switch sce {
	case enum.SceneUnlimited:
	case enum.SceneLive:
		if liveStatus != int(enum.LiveStatusLive) {
			log.Printf("[live.Gift] 场景不满足: 需要直播中，当前状态=%d", liveStatus)
			return false
		}
	case enum.SceneNotLive:
		if liveStatus == int(enum.LiveStatusLive) {
			log.Printf("[live.Gift] 场景不满足: 需要非直播状态，当前为直播中")
			return false
		}
	default:
		log.Printf("[live.Gift] 未知的场景: %d", sce)
		return false
	}
	minBatteryValue, _ := strconv.ParseInt(giftCfg.MinBattery, 10, 64)
	priceValue := (info.Price * info.Num) / 10
	if priceValue < minBatteryValue {
		log.Printf("[live.Gift] 礼物价值低于起始感谢电池数: 礼物电池=%d, 起始感谢电池数=%d", priceValue, minBatteryValue)
		return false
	}
	return true
}

// add 将一条礼物答谢信息加入合并缓冲区
func (b *giftMergeBuffer) add(info *giftThankInfo, flushFn func([]*giftThankInfo)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	uid := info.UID
	entry, ok := b.entries[uid]
	if !ok {
		entry = &mergeEntry{}
		b.entries[uid] = entry
	}
	entry.infos = append(entry.infos, info)
	// 重置 debounce timer
	if entry.timer != nil {
		entry.timer.Stop()
	}
	entry.timer = time.AfterFunc(b.interval, func() {
		b.mu.Lock()
		entry := b.entries[uid]
		if entry != nil {
			delete(b.entries, uid)
		}
		b.mu.Unlock()

		if entry != nil && entry.timer != nil {
			entry.timer.Stop()
		}
		if entry != nil && len(entry.infos) > 0 {
			flushFn(entry.infos)
		}
	})
}

// resolveGiftVars 按需解析礼物答谢模板变量，返回变量名 → 值的映射。支持单条与合并后的多条礼物。
func (p *giftProcessor) resolveGiftVars(infos []*giftThankInfo, needed map[string]bool, showCount bool, roomCfg robotconfig.RoomConfig) map[string]string {
	vars := make(map[string]string)
	if len(infos) == 0 {
		return vars
	}
	if needed["name"] {
		vars["name"] = infos[0].Uname
		maxNameLengthValue, _ := strconv.ParseInt(roomCfg.MaxNameLength, 10, 64)
		if maxNameLengthValue > 0 {
			if utf8.RuneCountInString(vars["name"]) > int(maxNameLengthValue) {
				switch ptr.ParseEnumInt[enum.NameTrimMode](roomCfg.NameTrimMode) {
				case enum.NameTrimModeTrimEnd:
					vars["name"] = util.TrimFromBack(vars["name"], int(maxNameLengthValue))
				case enum.NameTrimModeTrimStart:
					vars["name"] = util.TrimFromFront(vars["name"], int(maxNameLengthValue))
				}
			}
		}
	}
	if needed["price"] {
		var totalPrice int64
		for _, info := range infos {
			totalPrice += info.Price * info.Num
		}
		vars["price"] = fmt.Sprintf("%.2f", float64(totalPrice)/100)
	}
	if needed["gift"] {
		vars["gift"] = buildGiftText(infos, showCount)
	}
	return vars
}

// buildGiftText 将缓冲区内若干条礼物答谢信息聚合成「礼物列表」文案：
// 同名礼物只出现一次、数量累加，不同礼物按首次出现顺序以「、」分隔。
func buildGiftText(infos []*giftThankInfo, showCount bool) string {
	type namedGift struct {
		name  string
		count int64
	}
	index := make(map[string]int, len(infos))
	merged := make([]namedGift, 0, len(infos))
	for _, info := range infos {
		if idx, ok := index[info.GiftName]; ok {
			merged[idx].count += info.Num
			continue
		}
		index[info.GiftName] = len(merged)
		merged = append(merged, namedGift{name: info.GiftName, count: info.Num})
	}
	parts := make([]string, 0, len(merged))
	for _, g := range merged {
		if showCount {
			parts = append(parts, fmt.Sprintf("%d个%s", g.count, g.name))
		} else {
			parts = append(parts, g.name)
		}
	}
	return strings.Join(parts, "、")
}

// sendGiftReply 渲染礼物答谢模板并发送弹幕，支持单条与合并后的多条礼物。
func (p *giftProcessor) sendGiftReply(infos []*giftThankInfo) {
	if len(infos) == 0 {
		return
	}
	var giftCfg robotconfig.GiftConfig
	if err := p.configCache.UnmarshalGroup("gift", &giftCfg); err != nil {
		log.Printf("[live.Gift] 加载礼物答谢配置失败: %v", err)
		return
	}
	var roomCfg robotconfig.RoomConfig
	if err := p.configCache.UnmarshalGroup("room", &roomCfg); err != nil {
		log.Printf("[live.Gift] 加载房间配置失败: %v", err)
		return
	}
	tmpl := PickRandom(giftCfg.Content)
	if tmpl == "" {
		return
	}
	needed := CollectVars([]string{tmpl})
	vars := p.resolveGiftVars(infos, needed, ptr.ParseBool(giftCfg.ShowCount), roomCfg)
	msg := RenderTemplate(tmpl, vars)
	// 追加盲盒盈亏（按数量累加）
	if ptr.ParseBool(giftCfg.IncludeBlindbox) {
		var totalPrice, totalOriginalPrice int64
		for _, info := range infos {
			if info.Original == enum.No && info.OriginalGiftPrice > 0 {
				totalPrice += info.Price * info.Num
				totalOriginalPrice += info.OriginalGiftPrice * info.Num
			}
		}
		diff := totalPrice - totalOriginalPrice
		if diff != 0 {
			revenue := math.Abs(float64(diff)) / 100
			status := "赚了"
			if diff < 0 {
				status = "亏了"
			}
			msg = fmt.Sprintf("%s | %s%.2f元", msg, status, revenue)
		}
	}
	p.enqueueDanmu(msg, "gift")
}

// processRedeem 黑名单赎回：用户在禁言中赠送礼物后，尝试解除直播间禁言并更新解禁结果
func (p *giftProcessor) processRedeem(ctx context.Context, uid, price, num, roomID, botUID int64) {
	if botUID == uid {
		return
	}
	// 获取用户是否正在禁言中
	black, err := p.liveUserBlacklistRepo.GetActiveByRoomUID(ctx, nil, roomID, uid, time.Now().Unix())
	if err != nil {
		log.Printf("[live.Gift] 获取黑名单数据失败: %v", err)
		return
	}
	if black == nil {
		return
	}
	// 验证是否满足解禁条件
	priceValue := (price * num) / 10
	if black.RansomAmount > priceValue {
		log.Printf("[live.Gift] 用户 %d 不满足解除黑名单的条件，需要赠送电池 %d 当前赠送电池 %d", uid, black.RansomAmount, priceValue)
		return
	}
	csrf, err := p.client.CSRF()
	if err != nil {
		log.Printf("[live.Gift] 获取黑名单列表需要的 CSRF 获取失败: %v", err)
		return
	}
	unmuteUser(ctx, p.client, p.liveUserBlacklistRepo, black, csrf)
}

// processReward 奖励发放：根据系统设置为用户发放奖励
func (p *giftProcessor) processReward(ctx context.Context, userID int64, thankInfo *giftThankInfo, botUID int64) {
	if botUID == thankInfo.UID {
		return
	}
	if err := grantGiftReward(ctx, p.configCache, p.liveUserSvc, userID,
		thankInfo.GiftName, thankInfo.GiftType, thankInfo.BadgeType, thankInfo.Price, thankInfo.Num); err != nil {
		log.Printf("[live.Gift] 奖励发放失败: %v", err)
	}
}
