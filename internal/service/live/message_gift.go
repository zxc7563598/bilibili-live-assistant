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
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_blacklist"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_credit_log"
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
	LiveUserBlacklistRepo live_user_blacklist.Repository
	liveUserCreditLogRepo live_user_credit_log.Repository
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

func newGiftProcessor(liveUserSvc *liveuser.Service, liveGiftRepo live_gift.Repository, LiveUserBlacklistRepo live_user_blacklist.Repository, liveUserCreditLogRepo live_user_credit_log.Repository, roomState *RoomState, configCache *robotconfig.Cache, client *bilibili.Client, getBotUID func() int64, enqueueDanmu func(msg string, kind string)) *giftProcessor {
	return &giftProcessor{
		liveUserSvc:           liveUserSvc,
		liveGiftRepo:          liveGiftRepo,
		LiveUserBlacklistRepo: LiveUserBlacklistRepo,
		liveUserCreditLogRepo: liveUserCreditLogRepo,
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
			GiftType:  enum.GiftTypeNormal,
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
			GiftType:  enum.GiftTypeGuard,
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
			GiftType:  enum.GiftTypeSuperChat,
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
		}
	}
	if thankInfo != nil {
		botUID := p.getBotUID()
		liveStatus := p.roomState.LiveStatus()
		// 礼物答谢
		p.processGiftIn(thankInfo, roomID, botUID, liveStatus)
		// 黑名单赎回
		p.processRedeem(ctx, thankInfo.UID, thankInfo.Price, thankInfo.Num, roomID, botUID)
		// 奖励发放
		p.processReward(ctx, thankInfo, botUID)
	}
	return nil
}

// processGiftIn 处理礼物答谢逻辑
func (p *giftProcessor) processGiftIn(info *giftThankInfo, roomID, botUID int64, liveStatus int) {
	if botUID == info.UID {
		return
	}
	var giftCfg robotconfig.GiftConfig
	if err := p.configCache.UnmarshalGroup("gift", &giftCfg); err != nil {
		log.Printf("[live.Gift] 加载礼物答谢配置失败: %v", err)
		return
	}
	if !ptr.ParseBool(giftCfg.Enabled) {
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
		parts := make([]string, 0, len(infos))
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
	needed := CollectVars(giftCfg.Content)
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
	black, err := p.LiveUserBlacklistRepo.GetActiveByRoomUID(ctx, nil, roomID, uid)
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
	unmuteUser(ctx, p.client, p.LiveUserBlacklistRepo, black, csrf)
}

// processReward 奖励发放：根据系统设置为用户发放奖励
func (p *giftProcessor) processReward(ctx context.Context, thankInfo *giftThankInfo, botUID int64) {
	if botUID == thankInfo.UID {
		return
	}
	var roomCfg robotconfig.RoomConfig
	if err := p.configCache.UnmarshalGroup("room", &roomCfg); err != nil {
		log.Printf("[live.Gift] 加载房间配置失败: %v", err)
		return
	}
	switch ptr.ParseEnumInt[enum.RewardPolicy](roomCfg.ConsumeRewardEnabled) {
	case enum.RewardPolicyBatteryReward:
		magnification, _ := strconv.ParseInt(roomCfg.ConsumeBatteryRate, 10, 64)
		rewardType := ptr.ParseEnumInt[enum.RewardType](roomCfg.RewardType)
		if err := p.processBatteryReward(ctx, thankInfo.UID, thankInfo.Uname, thankInfo.GiftName, rewardType, thankInfo.Price, thankInfo.Num, magnification); err != nil {
			log.Printf("[live.Gift] 奖励发放失败: %v", err)
		}
	case enum.RewardPolicyVipReward:
		if thankInfo.GiftType == enum.GiftTypeGuard {
			var reward int64
			switch thankInfo.BadgeType {
			case enum.BadgeTypeL1:
				reward, _ = strconv.ParseInt(roomCfg.CaptainRewardAmount, 10, 64)
			case enum.BadgeTypeL2:
				reward, _ = strconv.ParseInt(roomCfg.CommanderRewardAmount, 10, 64)
			case enum.BadgeTypeL3:
				reward, _ = strconv.ParseInt(roomCfg.GovernorRewardAmount, 10, 64)
			}
			rewardType := ptr.ParseEnumInt[enum.RewardType](roomCfg.RewardType)
			if err := p.processVipReward(ctx, thankInfo.UID, thankInfo.Uname, rewardType, thankInfo.BadgeType, reward); err != nil {
				log.Printf("[live.Gift] 奖励发放失败: %v", err)
			}
		}
	}
}

// processBatteryReward 按消费电池发放积分
func (p *giftProcessor) processBatteryReward(ctx context.Context, uid int64, uname, giftName string, rewardType enum.RewardType, price, num, magnification int64) error {
	userID, err := p.liveUserSvc.EnsureUser(ctx, uid, uname)
	if err != nil {
		return err
	}
	battery := price / 10
	params := live_user_credit_log.AddCreditLogParams{
		UserID:       userID,
		ChangeType:   enum.ChangeTypeIncrease,
		ChangeAmount: (battery * num) * magnification,
		BizType:      "gift",
		Remark:       fmt.Sprintf("用户赠送 %d 个价值 %d 电池的 %s (积分倍率 %d)", num, battery, giftName, magnification),
		OperatorType: enum.OperatorTypeSystem,
		OperatorID:   0,
	}
	switch rewardType {
	case enum.RewardTypePoints:
		_, err = p.liveUserCreditLogRepo.AddPointsLog(ctx, nil, params)
	case enum.RewardTypeStars:
		_, err = p.liveUserCreditLogRepo.AddStarsLog(ctx, nil, params)
	}
	return err
}

// processVipReward 按航海类型发放积分
func (p *giftProcessor) processVipReward(ctx context.Context, uid int64, uname string, rewardType enum.RewardType, level enum.BadgeType, reward int64) error {
	userID, err := p.liveUserSvc.EnsureUser(ctx, uid, uname)
	params := live_user_credit_log.AddCreditLogParams{
		UserID:       userID,
		ChangeType:   enum.ChangeTypeIncrease,
		ChangeAmount: reward,
		BizType:      "gift",
		Remark:       fmt.Sprintf("用户开通 %s 奖励 %d", level.Text("zh"), reward),
		OperatorType: enum.OperatorTypeSystem,
		OperatorID:   0,
	}
	switch rewardType {
	case enum.RewardTypePoints:
		_, err = p.liveUserCreditLogRepo.AddPointsLog(ctx, nil, params)
	case enum.RewardTypeStars:
		_, err = p.liveUserCreditLogRepo.AddStarsLog(ctx, nil, params)
	}
	return err
}
