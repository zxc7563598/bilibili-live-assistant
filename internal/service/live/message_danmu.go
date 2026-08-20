package live

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_blacklist"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_credit_log"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_sign_log"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/util"
)

// danmuProcessor 处理弹幕消息（DANMU_MSG）
type danmuProcessor struct {
	liveUserSvc           *liveuser.Service
	liveDanmuRepo         live_danmu.Repository
	liveGiftRepo          live_gift.Repository
	liveUserCreditLogRepo live_user_credit_log.Repository
	liveUserSignLogRepo   live_user_sign_log.Repository
	LiveUserBlacklistRepo live_user_blacklist.Repository
	roomState             *RoomState
	configCache           *robotconfig.Cache
	client                *bilibili.Client
	getBotUID             func() int64
	enqueueDanmu          func(msg string, kind string)
}

func newDanmuProcessor(liveUserSvc *liveuser.Service, liveDanmuRepo live_danmu.Repository, liveGiftRepo live_gift.Repository, liveUserCreditLogRepo live_user_credit_log.Repository, liveUserSignLogRepo live_user_sign_log.Repository, LiveUserBlacklistRepo live_user_blacklist.Repository, roomState *RoomState, configCache *robotconfig.Cache, client *bilibili.Client, getBotUID func() int64, enqueueDanmu func(msg string, kind string)) *danmuProcessor {
	return &danmuProcessor{
		liveUserSvc:           liveUserSvc,
		liveDanmuRepo:         liveDanmuRepo,
		liveGiftRepo:          liveGiftRepo,
		liveUserCreditLogRepo: liveUserCreditLogRepo,
		liveUserSignLogRepo:   liveUserSignLogRepo,
		LiveUserBlacklistRepo: LiveUserBlacklistRepo,
		roomState:             roomState,
		configCache:           configCache,
		client:                client,
		getBotUID:             getBotUID,
		enqueueDanmu:          enqueueDanmu,
	}
}

func (p *danmuProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdDanmuMsg}
}

func (p *danmuProcessor) Process(ctx context.Context, cmd live.Cmd, data any, roomID int64) error {
	info, ok := data.(*live.DanmuMsgInfo)
	if !ok {
		log.Printf("[live.Danmu] 数据类型断言失败，期望 *live.DanmuMsgInfo，实际 %T", data)
		return nil
	}
	// 存储弹幕记录到数据库
	danmu := &model.LiveDanmu{
		RoomID:      roomID,
		UID:         info.UID,
		Uname:       info.Uname,
		Msg:         info.Msg,
		LiveID:      0,
		BadgeUID:    info.BadgeUID,
		BadgeUname:  info.BadgeUname,
		BadgeRoomID: info.BadgeRoomID,
		BadgeName:   info.BadgeName,
		BadgeLevel:  info.BadgeLevel,
		BadgeType:   enum.BadgeType(info.BadgeType),
		SendAt:      time.Now().Unix(),
	}
	if _, err := p.liveDanmuRepo.Create(ctx, nil, danmu); err != nil {
		log.Printf("[live.Danmu] 弹幕存储失败: %v", err)
	}
	botUID := p.getBotUID()
	liveStatus := p.roomState.LiveStatus()
	// 签到检测
	p.processSignIn(ctx, info, roomID, botUID, liveStatus)
	// 自动回复关键词检测
	p.processReplyIn(ctx, info, roomID, botUID, liveStatus)
	return nil
}

// processSignIn 处理签到逻辑
func (p *danmuProcessor) processSignIn(ctx context.Context, info *live.DanmuMsgInfo, roomID, botUID int64, liveStatus int) {
	if botUID == info.UID {
		return
	}
	var signCfg robotconfig.SignConfig
	if err := p.configCache.UnmarshalGroup("sign", &signCfg); err != nil {
		log.Printf("[live.Sign] 加载签到配置失败: %v", err)
		return
	}
	if !ptr.ParseBool(signCfg.Enabled) {
		return
	}
	if !p.checkSignCondition(signCfg, info, roomID, liveStatus) {
		return
	}
	if strings.Contains(info.Msg, signCfg.Keyword) {
		p.handleSignIn(ctx, signCfg, info)
	}
	if strings.Contains(info.Msg, signCfg.QueryKeyword) {
		p.handleSignQuery(ctx, signCfg, info)
	}
}

// checkSignCondition 校验签到的 Requirement 和 Scene 是否满足
func (p *danmuProcessor) checkSignCondition(signCfg robotconfig.SignConfig, info *live.DanmuMsgInfo, roomID int64, liveStatus int) bool {
	req := ptr.ParseEnumInt[enum.Requirement](signCfg.Requirement)
	switch req {
	case enum.RequirementUnlimited:
	case enum.RequirementHasBadge:
		if info.BadgeRoomID != roomID {
			log.Printf("[live.Sign] 签到条件不满足: 需要携带当前直播间牌子，用户牌子 roomID=%d, 当前 roomID=%d", info.BadgeRoomID, roomID)
			return false
		}
	case enum.RequirementHasSailBadge:
		if info.BadgeRoomID != roomID {
			log.Printf("[live.Sign] 签到条件不满足: 需要携带当前直播间牌子，用户牌子 roomID=%d, 当前 roomID=%d", info.BadgeRoomID, roomID)
			return false
		}
		if enum.BadgeType(info.BadgeType) == enum.BadgeTypeL0 {
			log.Printf("[live.Sign] 签到条件不满足: 需要非 L0 勋章")
			return false
		}
	default:
		log.Printf("[live.Sign] 未知的签到条件: %d", req)
		return false
	}
	sce := ptr.ParseEnumInt[enum.Scene](signCfg.Scene)
	switch sce {
	case enum.SceneUnlimited:
	case enum.SceneLive:
		if liveStatus != int(enum.LiveStatusLive) {
			log.Printf("[live.Sign] 场景不满足: 需要直播中，当前状态=%d", liveStatus)
			return false
		}
	case enum.SceneNotLive:
		if liveStatus == int(enum.LiveStatusLive) {
			log.Printf("[live.Sign] 场景不满足: 需要非直播状态，当前为直播中")
			return false
		}
	default:
		log.Printf("[live.Sign] 未知的场景: %d", sce)
		return false
	}
	return true
}

// handleSignIn 处理签到：查重 → 已签到发重复回复 / 未签到则记录、发奖励、发成功回复
func (p *danmuProcessor) handleSignIn(ctx context.Context, signCfg robotconfig.SignConfig, info *live.DanmuMsgInfo) {
	exists, err := p.liveUserSignLogRepo.ExistsByUIDToday(ctx, nil, info.UID)
	if err != nil {
		log.Printf("[live.Sign] 查询用户签到信息失败: %v", err)
		return
	}
	if exists {
		p.sendSignReply(ctx, signCfg.RepeatReply, info)
		return
	}
	// 记录签到
	rewardType := ptr.ParseEnumInt[enum.CreditType](signCfg.RewardType)
	rewardAmount, _ := strconv.ParseInt(signCfg.RewardAmount, 10, 64)
	if _, err := p.liveUserSignLogRepo.Create(ctx, nil, &model.LiveUserSignLog{
		UID:          info.UID,
		Uname:        info.Uname,
		Msg:          info.Msg,
		BadgeUID:     info.BadgeUID,
		BadgeUname:   info.BadgeUname,
		BadgeRoomID:  info.BadgeRoomID,
		BadgeName:    info.BadgeName,
		BadgeLevel:   info.BadgeLevel,
		BadgeType:    enum.BadgeType(info.BadgeType),
		RewardType:   rewardType,
		RewardAmount: rewardAmount,
	}); err != nil {
		log.Printf("[live.Sign] 签到信息添加失败: %v", err)
	}
	// 发放奖励
	if rewardAmount > 0 {
		if err := p.grantSignReward(ctx, info, rewardType, rewardAmount); err != nil {
			log.Printf("[live.Sign] 签到奖励发放失败: %v", err)
		}
	}
	p.sendSignReply(ctx, signCfg.SuccessReply, info)
}

// handleSignQuery 处理签到查询：发送查询回复
func (p *danmuProcessor) handleSignQuery(ctx context.Context, signCfg robotconfig.SignConfig, info *live.DanmuMsgInfo) {
	p.sendSignReply(ctx, signCfg.QueryReply, info)
}

// resolveSignVars 按需查询签到相关数据，返回变量名 → 值的映射
func (p *danmuProcessor) resolveSignVars(ctx context.Context, info *live.DanmuMsgInfo, needed map[string]bool, roomCfg robotconfig.RoomConfig) map[string]string {
	vars := make(map[string]string)
	// 无 IO：直接从 info 获取
	if needed["name"] {
		vars["name"] = info.Uname
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
	if needed["guard"] {
		vars["guard"] = enum.BadgeType(info.BadgeType).Text("zh")
	}
	// 单次 IO：points 和 stars 共享一次 GetByUID
	if needed["points"] || needed["stars"] {
		balance, err := p.liveUserSvc.GetUserBalance(ctx, info.UID)
		if err == nil && balance != nil {
			if needed["points"] {
				vars["points"] = strconv.FormatInt(balance.Points, 10)
			}
			if needed["stars"] {
				vars["stars"] = strconv.FormatInt(balance.Stars, 10)
			}
		}
	}
	// IO：总签到次数
	if needed["days"] {
		count, err := p.liveUserSignLogRepo.CountByUID(ctx, nil, info.UID)
		if err == nil {
			vars["days"] = strconv.FormatInt(count, 10)
		}
	}
	// IO：连续签到天数
	if needed["streak"] {
		streak, err := p.liveUserSignLogRepo.StreakByUID(ctx, nil, info.UID)
		if err == nil {
			vars["streak"] = strconv.FormatInt(streak, 10)
		}
	}
	return vars
}

// sendSignReply 签到相关从模板列表中随机选取一条，渲染变量后发送弹幕
func (p *danmuProcessor) sendSignReply(ctx context.Context, templates []string, info *live.DanmuMsgInfo) {
	tmpl := PickRandom(templates)
	if tmpl == "" {
		return
	}
	var roomCfg robotconfig.RoomConfig
	if err := p.configCache.UnmarshalGroup("room", &roomCfg); err != nil {
		log.Printf("[live.Sign] 加载房间配置失败: %v", err)
		return
	}
	needed := CollectVars(templates)
	vars := p.resolveSignVars(ctx, info, needed, roomCfg)
	msg := RenderTemplate(tmpl, vars)
	p.enqueueDanmu(msg, "sign")
}

// grantSignReward 注册用户（如不存在）并发放签到奖励
func (p *danmuProcessor) grantSignReward(ctx context.Context, info *live.DanmuMsgInfo, rewardType enum.CreditType, rewardAmount int64) error {
	userID, err := p.liveUserSvc.EnsureUser(ctx, info.UID, info.Uname)
	if err != nil {
		return err
	}
	params := live_user_credit_log.AddCreditLogParams{
		UserID:       userID,
		ChangeType:   enum.ChangeTypeIncrease,
		ChangeAmount: rewardAmount,
		BizType:      "sign",
		Remark:       "用户签到成功系统自动增加",
		OperatorType: enum.OperatorTypeSystem,
		OperatorID:   0,
	}
	switch rewardType {
	case enum.CreditTypePoints:
		_, err = p.liveUserCreditLogRepo.AddPointsLog(ctx, nil, params)
	case enum.CreditTypeStars:
		_, err = p.liveUserCreditLogRepo.AddStarsLog(ctx, nil, params)
	}
	return err
}

// processReplyIn 处理自动回复逻辑
func (p *danmuProcessor) processReplyIn(ctx context.Context, info *live.DanmuMsgInfo, roomID, botUID int64, liveStatus int) {
	if botUID == info.UID {
		return
	}
	var replyCfg robotconfig.ReplyConfig
	if err := p.configCache.UnmarshalGroup("reply", &replyCfg); err != nil {
		log.Printf("[live.Reply] 加载签到配置失败: %v", err)
		return
	}
	if !ptr.ParseBool(replyCfg.Enabled) {
		return
	}
	if !p.checkReplyCondition(replyCfg, info, roomID, liveStatus) {
		return
	}
	for _, reply := range replyCfg.Content {
		keywordMatchPolicyValue := parseMatchPolicy(reply.KeywordMatchPolicy)
		safeWordMatchPolicyValue := parseMatchPolicy(reply.SafeWordMatchPolicy)
		if containsMatch(info.Msg, reply.Keyword, keywordMatchPolicyValue) {
			if !containsMatch(info.Msg, reply.SafeWord, safeWordMatchPolicyValue) {
				// 加入黑名单
				if ptr.ParseBool(reply.MuteSender) {
					if err := p.blockUserForReply(info.UID, roomID, info.Uname, info.Msg, reply.MuteDuration, reply.RansomAmount); err != nil {
						log.Printf("[live.Reply] 加入黑名单失败: %v", err)
						return
					}
				}
				// 回复信息
				p.sendReply(ctx, reply.ReplyContent, info, roomID)
			}
		}
	}
}

// checkReplyCondition 校验自动回复的 Requirement 和 Scene 是否满足
func (p *danmuProcessor) checkReplyCondition(replyCfg robotconfig.ReplyConfig, info *live.DanmuMsgInfo, roomID int64, liveStatus int) bool {
	req := ptr.ParseEnumInt[enum.Requirement](replyCfg.Requirement)
	switch req {
	case enum.RequirementUnlimited:
	case enum.RequirementHasBadge:
		if info.BadgeRoomID != roomID {
			log.Printf("[live.Reply] 回复条件不满足: 需要携带当前直播间牌子，用户牌子 roomID=%d, 当前 roomID=%d", info.BadgeRoomID, roomID)
			return false
		}
	case enum.RequirementHasSailBadge:
		if info.BadgeRoomID != roomID {
			log.Printf("[live.Reply] 回复条件不满足: 需要携带当前直播间牌子，用户牌子 roomID=%d, 当前 roomID=%d", info.BadgeRoomID, roomID)
			return false
		}
		if enum.BadgeType(info.BadgeType) == enum.BadgeTypeL0 {
			log.Printf("[live.Reply] 回复条件不满足: 需要非 L0 勋章")
			return false
		}
	default:
		log.Printf("[live.Reply] 未知的回复条件: %d", req)
		return false
	}
	sce := ptr.ParseEnumInt[enum.Scene](replyCfg.Scene)
	switch sce {
	case enum.SceneUnlimited:
	case enum.SceneLive:
		if liveStatus != int(enum.LiveStatusLive) {
			log.Printf("[live.Reply] 场景不满足: 需要直播中，当前状态=%d", liveStatus)
			return false
		}
	case enum.SceneNotLive:
		if liveStatus == int(enum.LiveStatusLive) {
			log.Printf("[live.Reply] 场景不满足: 需要非直播状态，当前为直播中")
			return false
		}
	default:
		log.Printf("[live.Reply] 未知的场景: %d", sce)
		return false
	}
	return true
}

// blockUserForReply 由自动回复触发的加入黑名单
func (p *danmuProcessor) blockUserForReply(uid, roomID int64, uname, msg, muteDuration, ransomAmount string) error {
	csrf, err := p.client.CSRF()
	if err != nil {
		return fmt.Errorf("用户 CSRF 获取失败: %w", err)
	}
	err = p.client.Room.AddSilentUser(context.Background(), roomID, uid, msg, csrf)
	if err != nil {
		return fmt.Errorf("用户加入黑名单失败: %w", err)
	}
	ransomAmountValue, _ := strconv.ParseInt(ransomAmount, 10, 64)
	muteDurationValue, _ := strconv.ParseInt(muteDuration, 10, 64)
	_, err = p.LiveUserBlacklistRepo.Create(context.Background(), nil, &model.LiveUserBlacklist{
		RoomID:          roomID,
		UID:             uid,
		Uname:           uname,
		Msg:             msg,
		RansomAmount:    ransomAmountValue,
		MuteDuration:    muteDurationValue,
		MuteExpiresAt:   getExpireTimestamp(muteDurationValue),
		UnmuteFailCount: 0,
		Status:          enum.MuteStatusMuted,
	})
	if err != nil {
		return fmt.Errorf("添加黑名单记录失败: %w", err)
	}
	return nil
}

// sendReply 自动回复相关从模板列表中随机选取一条，渲染变量后发送弹幕
func (p *danmuProcessor) sendReply(ctx context.Context, templates []string, info *live.DanmuMsgInfo, roomID int64) {
	tmpl := PickRandom(templates)
	if tmpl == "" {
		return
	}
	var roomCfg robotconfig.RoomConfig
	if err := p.configCache.UnmarshalGroup("room", &roomCfg); err != nil {
		log.Printf("[live.Reply] 加载房间配置失败: %v", err)
		return
	}
	needed := CollectVars(templates)
	vars := p.resolveReplyVars(ctx, info, needed, roomID, roomCfg)
	msg := RenderTemplate(tmpl, vars)
	p.enqueueDanmu(msg, "reply")
}

// resolveReplyVars 按需查询自动回复相关数据，返回变量名 → 值的映射
func (p *danmuProcessor) resolveReplyVars(ctx context.Context, info *live.DanmuMsgInfo, needed map[string]bool, roomID int64, roomCfg robotconfig.RoomConfig) map[string]string {
	vars := make(map[string]string)
	// 无 IO：直接从 info 获取
	if needed["name"] {
		vars["name"] = info.Uname
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
	if needed["guard"] {
		vars["guard"] = enum.BadgeType(info.BadgeType).Text("zh")
	}
	// IO：计算盲盒变量
	needUserProfit := needed["daily_net"] || needed["weekly_net"] || needed["monthly_net"] || needed["total_net"]
	needRoomProfit := needed["room_daily_net"] || needed["room_weekly_net"] || needed["room_monthly_net"] || needed["room_total_net"]
	if needUserProfit || needRoomProfit {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		// 处理时间变量
		day := live_gift.TimeRange{Start: today.Unix(), End: now.Unix()}
		daysSinceMonday := int(now.Weekday()) - int(time.Monday)
		if daysSinceMonday < 0 {
			daysSinceMonday += 7
		}
		weekStart := time.Date(now.Year(), now.Month(), now.Day()-daysSinceMonday, 0, 0, 0, 0, now.Location())
		week := live_gift.TimeRange{Start: weekStart.Unix(), End: now.Unix()}
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		month := live_gift.TimeRange{Start: monthStart.Unix(), End: now.Unix()}
		// 获取用户纬度数据（如果需要）
		if needUserProfit {
			profit, err := p.liveGiftRepo.SumBlindBoxProfit(ctx, nil, info.UID, roomID, day, week, month)
			if err != nil {
				log.Printf("[live.Reply] 查询用户盲盒盈利失败: %v", err)
			} else {
				if needed["daily_net"] {
					vars["daily_net"] = strconv.FormatInt(profit.Daily, 10)
				}
				if needed["weekly_net"] {
					vars["weekly_net"] = strconv.FormatInt(profit.Weekly, 10)
				}
				if needed["monthly_net"] {
					vars["monthly_net"] = strconv.FormatInt(profit.Monthly, 10)
				}
				if needed["total_net"] {
					vars["total_net"] = strconv.FormatInt(profit.Total, 10)
				}
			}
		}
		// 获取房间纬度数据（如果需要）
		if needRoomProfit {
			profit, err := p.liveGiftRepo.SumBlindBoxProfit(ctx, nil, 0, roomID, day, week, month)
			if err != nil {
				log.Printf("[live.Reply] 查询直播间盲盒盈利失败: %v", err)
			} else {
				if needed["room_daily_net"] {
					vars["room_daily_net"] = strconv.FormatInt(profit.Daily, 10)
				}
				if needed["room_weekly_net"] {
					vars["room_weekly_net"] = strconv.FormatInt(profit.Weekly, 10)
				}
				if needed["room_monthly_net"] {
					vars["room_monthly_net"] = strconv.FormatInt(profit.Monthly, 10)
				}
				if needed["room_total_net"] {
					vars["room_total_net"] = strconv.FormatInt(profit.Total, 10)
				}
			}
		}
	}
	return vars
}

// parseMatchPolicy 解析匹配策略，空值（老版本配置）回退为 MatchAny
func parseMatchPolicy(s string) enum.MatchPolicy {
	if s == "" {
		return enum.MatchPolicyMatchAny
	}
	return ptr.ParseEnumInt[enum.MatchPolicy](s)
}

// containsMatch 自动回复相关关键词匹配
func containsMatch(s string, items []string, matchPolicy enum.MatchPolicy) bool {
	switch matchPolicy {
	case enum.MatchPolicyMatchAny:
		for _, item := range items {
			if strings.Contains(s, item) {
				return true
			}
		}
		return false
	case enum.MatchPolicyMatchAll:
		for _, item := range items {
			if !strings.Contains(s, item) {
				return false
			}
		}
		return len(items) > 0
	default:
		return false
	}
}

// getExpireTimestamp 自动回复相关禁言获取过期时间戳
func getExpireTimestamp(muteDurationValue int64) int64 {
	if muteDurationValue <= 0 {
		return time.Now().Add(99 * 365 * 24 * time.Hour).Unix()
	}
	return time.Now().Add(time.Duration(muteDurationValue) * time.Minute).Unix()
}
