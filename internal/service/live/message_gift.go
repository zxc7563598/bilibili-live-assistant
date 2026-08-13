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

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_blacklist"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
)

// giftMergeWindow 礼物合并 debounce 窗口：同一用户在此时间内的多个礼物会被合并为一条答谢消息
const giftMergeWindow = 3 * time.Second

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
	liveGiftRepo          live_gift.Repository
	LiveUserBlacklistRepo live_user_blacklist.Repository
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
	Price             int64          // 价格（分）
	Num               int64          // 数量
	Original          enum.YesNo     // 是否原始礼物
	OriginalGiftPrice int64          // 原始礼物价格
	Badge             bool           // 是否是本直播间牌子
	BadgeType         enum.BadgeType // 牌子类型
}

func newGiftProcessor(liveGiftRepo live_gift.Repository, LiveUserBlacklistRepo live_user_blacklist.Repository, roomState *RoomState, configCache *robotconfig.Cache, client *bilibili.Client, getBotUID func() int64, enqueueDanmu func(msg string, kind string)) *giftProcessor {
	return &giftProcessor{
		liveGiftRepo:          liveGiftRepo,
		LiveUserBlacklistRepo: LiveUserBlacklistRepo,
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
	var gift *model.LiveGift
	var thankInfo *giftThankInfo
	switch cmd {
	case live.CmdSendGift, live.CmdSendGiftV2:
		info, ok := data.(*live.SendGiftInfo)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.SendGiftInfo，实际 %T", data)
			return nil
		}
		gift = &model.LiveGift{
			RoomID:     roomID,
			UID:        info.UID,
			Uname:      info.Uname,
			GiftType:   enum.GiftTypeNormal,
			GiftID:     info.GiftID,
			GiftName:   info.GiftName,
			Price:      info.Price,
			Num:        info.Num,
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
		thankInfo = &giftThankInfo{
			UID:       info.UID,
			Uname:     info.Uname,
			GiftID:    info.GiftID,
			GiftName:  info.GiftName,
			Price:     info.Price,
			Num:       info.Num,
			Original:  enum.Yes,
			Badge:     info.BadgeUID == info.AnchorID,
			BadgeType: enum.BadgeType(info.BadgeType),
		}
		thankInfo.OriginalGiftPrice = info.Price
		if info.BlindGift != nil && info.BlindGift.OriginalGiftPrice > 0 {
			thankInfo.Original = enum.No
			thankInfo.OriginalGiftPrice = info.BlindGift.OriginalGiftPrice
		}
	case live.CmdGuardBuy:
		info, ok := data.(*live.GuardBuyInfo)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.GuardBuyInfo，实际 %T", data)
			return nil
		}
		gift = &model.LiveGift{
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
		thankInfo = &giftThankInfo{
			UID:       info.UID,
			Uname:     info.Uname,
			GiftID:    info.GiftID,
			GiftName:  info.GiftName,
			Price:     info.Price,
			Num:       1,
			Original:  enum.Yes,
			Badge:     true,
			BadgeType: enum.BadgeType(info.GuardLevel),
		}
		thankInfo.OriginalGiftPrice = info.Price
	case live.CmdSuperDanmuMsg:
		info, ok := data.(*live.SuperChatMessage)
		if !ok {
			log.Printf("[live.Gift] 数据类型断言失败，期望 *live.SuperChatMessage，实际 %T", data)
			return nil
		}
		gift = &model.LiveGift{
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
		thankInfo = &giftThankInfo{
			UID:       info.UID,
			Uname:     info.Uname,
			GiftID:    info.GiftID,
			GiftName:  info.GiftName,
			Price:     info.Price,
			Num:       1,
			Original:  enum.Yes,
			Badge:     info.BadgeRoomID == roomID,
			BadgeType: enum.BadgeType(info.BadgeType),
		}
		thankInfo.OriginalGiftPrice = info.Price
	}
	if gift != nil {
		if _, err := p.liveGiftRepo.Create(ctx, nil, gift); err != nil {
			log.Printf("[live.Gift] 礼物存储失败: %v", err)
			return err
		}
	}
	// 礼物答谢
	if thankInfo != nil {
		botUID := p.getBotUID()
		liveStatus := p.roomState.LiveStatus()
		p.processGiftIn(ctx, thankInfo, roomID, botUID, liveStatus)
	}
	return nil
}

func (p *giftProcessor) processGiftIn(ctx context.Context, info *giftThankInfo, roomID, botUID int64, liveStatus int) {
	if botUID == info.UID {
		return
	}
	var cfg robotconfig.GiftConfig
	if err := p.configCache.UnmarshalGroup("gift", &cfg); err != nil {
		log.Printf("[live.Gift] 加载礼物答谢配置失败: %v", err)
		return
	}
	if !ptr.ParseBool(cfg.Enabled) {
		return
	}
	if !p.checkGiftCondition(cfg, info, roomID, liveStatus) {
		return
	}

	if ptr.ParseBool(cfg.MergeGift) {
		// 合并模式：加入缓冲区，由 debounce timer 触发合并发送
		p.mergeBuffer.add(info, func(infos []*giftThankInfo) {
			p.flushUserGifts(infos)
		})
	} else {
		p.sendGiftReply(ctx, cfg.Content, info, cfg.ShowCount, cfg.IncludeBlindbox)
	}
}

// checkGiftCondition 校验礼物答谢的 Requirement, Scene 以及 MinBattery 是否满足
func (p *giftProcessor) checkGiftCondition(cfg robotconfig.GiftConfig, info *giftThankInfo, roomID int64, liveStatus int) bool {
	req := ptr.ParseEnumInt[enum.Requirement](cfg.Requirement)
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
	sce := ptr.ParseEnumInt[enum.Scene](cfg.Scene)
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
	minBatteryValue, _ := strconv.ParseInt(cfg.MinBattery, 10, 64)
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

// resolveGiftVars 按需查询礼物答谢相关数据，返回变量名 → 值的映射
func (p *giftProcessor) resolveGiftVars(ctx context.Context, info *giftThankInfo, needed map[string]bool, showCount bool) map[string]string {
	vars := make(map[string]string)
	// 无 IO：直接从 info 获取
	if needed["name"] {
		vars["name"] = info.Uname
	}
	if needed["price"] {
		vars["price"] = fmt.Sprintf("%.2f", float64(info.Price*info.Num)/100)
	}
	if needed["gift"] {
		vars["gift"] = info.GiftName
		if showCount {
			vars["gift"] = fmt.Sprintf("%d个%s", info.Num, info.GiftName)
		}
	}
	return vars
}

// resolveMergeGiftVars 为合并后的多条礼物信息解析模板变量，功能同 resolveGiftVars
func (p *giftProcessor) resolveMergeGiftVars(infos []*giftThankInfo, needed map[string]bool, showCount bool) map[string]string {
	vars := make(map[string]string)
	if needed["name"] {
		vars["name"] = infos[0].Uname
	}
	if needed["price"] {
		var totalPrice int64
		for _, info := range infos {
			totalPrice += info.Price * info.Num
		}
		vars["price"] = fmt.Sprintf("%.2f", float64(totalPrice)/100)
	}
	if needed["gift"] {
		var parts []string
		for _, info := range infos {
			if showCount {
				parts = append(parts, fmt.Sprintf("%d个%s", info.Num, info.GiftName))
			} else {
				parts = append(parts, info.GiftName)
			}
		}
		vars["gift"] = strings.Join(parts, "、")
	}
	return vars
}

// sendGiftReply 礼物答谢相关从模板列表中随机选取一条，渲染变量后发送弹幕
func (p *giftProcessor) sendGiftReply(ctx context.Context, templates []string, info *giftThankInfo, showCount, includeBlindbox string) {
	tmpl := PickRandom(templates)
	if tmpl == "" {
		return
	}
	needed := CollectVars(templates)
	vars := p.resolveGiftVars(ctx, info, needed, ptr.ParseBool(showCount))
	msg := RenderTemplate(tmpl, vars)
	// 追加盲盒亏损
	includeBlindboxValue := ptr.ParseBool(includeBlindbox)
	if includeBlindboxValue && info.Original == enum.No && info.OriginalGiftPrice > 0 {
		diff := info.Price - info.OriginalGiftPrice
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

// flushUserGifts 将同一用户累积的多个礼物信息合并为一条答谢弹幕并加入发送队列，功能同 sendGiftReply
func (p *giftProcessor) flushUserGifts(infos []*giftThankInfo) {
	if len(infos) == 0 {
		return
	}
	var cfg robotconfig.GiftConfig
	if err := p.configCache.UnmarshalGroup("gift", &cfg); err != nil {
		log.Printf("[live.Gift] 合并发送时加载礼物答谢配置失败: %v", err)
		return
	}
	tmpl := PickRandom(cfg.Content)
	if tmpl == "" {
		return
	}
	needed := CollectVars(cfg.Content)
	vars := p.resolveMergeGiftVars(infos, needed, ptr.ParseBool(cfg.ShowCount))
	msg := RenderTemplate(tmpl, vars)
	// 追加聚合盲盒盈亏
	includeBlindboxValue := ptr.ParseBool(cfg.IncludeBlindbox)
	if includeBlindboxValue {
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
