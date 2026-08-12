package live

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_credit_log"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_sign_log"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
)

// danmuProcessor 处理弹幕消息（DANMU_MSG）
type danmuProcessor struct {
	liveUserRepo          live_user.Repository
	liveDanmuRepo         live_danmu.Repository
	liveUserCreditLogRepo live_user_credit_log.Repository
	liveUserSignLogRepo   live_user_sign_log.Repository
	roomState             *RoomState
	configCache           *robotconfig.Cache
	getBotUID             func() int64
	enqueueDanmu          func(msg string, priority int)
}

func newDanmuProcessor(liveUserRepo live_user.Repository, liveDanmuRepo live_danmu.Repository, liveUserCreditLogRepo live_user_credit_log.Repository, liveUserSignLogRepo live_user_sign_log.Repository, roomState *RoomState, configCache *robotconfig.Cache, getBotUID func() int64, enqueueDanmu func(msg string, priority int)) *danmuProcessor {
	return &danmuProcessor{
		liveUserRepo:          liveUserRepo,
		liveDanmuRepo:         liveDanmuRepo,
		liveUserCreditLogRepo: liveUserCreditLogRepo,
		liveUserSignLogRepo:   liveUserSignLogRepo,
		roomState:             roomState,
		configCache:           configCache,
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
		return err
	}
	botUID := p.getBotUID()
	liveStatus := p.roomState.LiveStatus()
	// 签到检测
	p.processSignIn(ctx, info, roomID, botUID, liveStatus)
	// 自动回复关键词检测
	return nil
}

// processSignIn 处理签到逻辑
// 校验 → 签到 / 查询，任一环节失败仅 log，不返回 error，保证后续弹幕处理不受影响
func (p *danmuProcessor) processSignIn(ctx context.Context, info *live.DanmuMsgInfo, roomID, botUID int64, liveStatus int) {
	if botUID == info.UID {
		return
	}
	var cfg robotconfig.SignConfig
	if err := p.configCache.UnmarshalGroup("sign", &cfg); err != nil {
		log.Printf("[live.Sign] 加载签到配置失败: %v", err)
		return
	}
	if !ptr.ParseBool(cfg.Enabled) {
		return
	}
	if !p.checkSignCondition(cfg, info, roomID, liveStatus) {
		return
	}
	if strings.Contains(info.Msg, cfg.Keyword) {
		p.handleSignIn(ctx, cfg, info)
	}
	if strings.Contains(info.Msg, cfg.QueryKeyword) {
		p.handleSignQuery(ctx, cfg, info)
	}
}

// checkSignCondition 校验签到的 Requirement 和 Scene 是否满足
func (p *danmuProcessor) checkSignCondition(cfg robotconfig.SignConfig, info *live.DanmuMsgInfo, roomID int64, liveStatus int) bool {
	req := ptr.ParseEnumInt[enum.Requirement](cfg.Requirement)
	switch req {
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
	sce := ptr.ParseEnumInt[enum.Scene](cfg.Scene)
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
func (p *danmuProcessor) handleSignIn(ctx context.Context, cfg robotconfig.SignConfig, info *live.DanmuMsgInfo) {
	exists, err := p.liveUserSignLogRepo.ExistsByUIDToday(ctx, nil, info.UID)
	if err != nil {
		log.Printf("[live.Sign] 查询用户签到信息失败: %v", err)
		return
	}
	if exists {
		p.sendSignReply(ctx, cfg.RepeatReply, info)
		return
	}
	// 记录签到
	rewardType := ptr.ParseEnumInt[enum.CreditType](cfg.RewardType)
	rewardAmount, _ := strconv.ParseInt(cfg.RewardAmount, 10, 64)
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
			return
		}
	}
	p.sendSignReply(ctx, cfg.SuccessReply, info)
}

// handleSignQuery 处理签到查询：发送查询回复
func (p *danmuProcessor) handleSignQuery(ctx context.Context, cfg robotconfig.SignConfig, info *live.DanmuMsgInfo) {
	p.sendSignReply(ctx, cfg.QueryReply, info)
}

// resolveSignVars 按需查询签到相关数据，返回变量名 → 值的映射
// 仅对 needed 中标记为 true 的变量发起数据库查询，未使用的变量不会触发额外 IO
func (p *danmuProcessor) resolveSignVars(ctx context.Context, info *live.DanmuMsgInfo, needed map[string]bool) map[string]string {
	vars := make(map[string]string)
	// 无 IO：直接从 info 获取
	if needed["name"] {
		vars["name"] = info.Uname
	}
	if needed["guard"] {
		vars["guard"] = enum.BadgeType(info.BadgeType).Text("zh")
	}
	// 单次 IO：points 和 stars 共享一次 GetByUID
	if needed["points"] || needed["stars"] {
		user, err := p.liveUserRepo.GetByUID(ctx, nil, info.UID)
		if err == nil && user != nil {
			if needed["points"] {
				vars["points"] = strconv.FormatInt(user.Points, 10)
			}
			if needed["stars"] {
				vars["stars"] = strconv.FormatInt(user.Stars, 10)
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

// sendSignReply 从模板列表中随机选取一条，渲染变量后发送弹幕
func (p *danmuProcessor) sendSignReply(ctx context.Context, templates []string, info *live.DanmuMsgInfo) {
	tmpl := PickRandom(templates)
	if tmpl == "" {
		return
	}
	needed := CollectVars(templates)
	vars := p.resolveSignVars(ctx, info, needed)
	msg := RenderTemplate(tmpl, vars)
	p.enqueueDanmu(msg, 100)
}

// grantSignReward 注册用户（如不存在）并发放签到奖励
func (p *danmuProcessor) grantSignReward(ctx context.Context, info *live.DanmuMsgInfo, rewardType enum.CreditType, rewardAmount int64) error {
	user, err := p.liveUserRepo.GetByUID(ctx, nil, info.UID)
	if err != nil {
		return err
	}
	if user == nil {
		user, err = p.liveUserRepo.Create(ctx, nil, &model.LiveUser{
			Uid:   info.UID,
			Uname: info.Uname,
		})
		if err != nil {
			return err
		}
	}
	params := live_user_credit_log.AddCreditLogParams{
		UserID:       user.ID,
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
