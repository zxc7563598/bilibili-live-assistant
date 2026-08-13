package live

import (
	"context"
	"log"
	"strconv"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_interact_word"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
)

// interactProcessor 处理用户互动消息（INTERACT_WORD_V2）
//
// 消息类型（MsgType）：
//
//	1 — 进入直播间
//	2 — 关注
//	3 — 分享
type interactProcessor struct {
	liveInteractWord live_interact_word.Repository
	roomState        *RoomState
	configCache      *robotconfig.Cache
	client           *bilibili.Client
	getBotUID        func() int64
	enqueueDanmu     func(msg string, kind string)
}

func newInteractProcessor(liveInteractWord live_interact_word.Repository, roomState *RoomState, configCache *robotconfig.Cache, client *bilibili.Client, getBotUID func() int64, enqueueDanmu func(msg string, kind string)) *interactProcessor {
	return &interactProcessor{
		liveInteractWord: liveInteractWord,
		roomState:        roomState,
		configCache:      configCache,
		client:           client,
		getBotUID:        getBotUID,
		enqueueDanmu:     enqueueDanmu,
	}
}

func (p *interactProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdInteractWord}
}

func (p *interactProcessor) Process(ctx context.Context, cmd live.Cmd, data any, roomID int64) error {
	info, ok := data.(*live.InteractWordV2Info)
	if !ok {
		log.Printf("[live.Interact] 数据类型断言失败，期望 *live.InteractWordV2Info，实际 %T", data)
		return nil
	}
	botUID := p.getBotUID()
	liveStatus := p.roomState.LiveStatus()
	anchorID := p.roomState.UID()
	// 计入数据库
	_, err := p.liveInteractWord.Create(ctx, nil, &model.LiveInteractWord{
		RoomID:     info.RoomID,
		UID:        info.UID,
		Uname:      info.Uname,
		MsgType:    enum.InteractType(info.MsgType),
		Timestamp:  info.Timestamp,
		BadgeUID:   info.BadgeUID,
		BadgeName:  info.BadgeName,
		BadgeLevel: info.BadgeLevel,
		BadgeType:  enum.BadgeType(info.BadgeType),
	})
	if err != nil {
		log.Printf("[live.Interact] 互动信息存储失败: %v", err)
	}
	// 进房欢迎/感谢关注/感谢分享等自动回复
	kind, ok := p.interactKind(enum.InteractType(info.MsgType))
	if !ok {
		return nil
	}
	p.processInteractIn(ctx, kind, info, anchorID, botUID, liveStatus)
	return nil
}

// interactKind 描述一类互动消息在回复流程中的差异点
type interactKind struct {
	group     string                                                      // 配置分组名：welcome / follow / share
	tag       string                                                      // 日志标签：Welcome / Follow / Share
	label     string                                                      // 中文说明：进房欢迎 / 感谢关注 / 感谢分享
	danmuKind string                                                      // 弹幕发送分类
	count     func(ctx context.Context, uid, roomID int64) (int64, error) // 累计次数
	streak    func(ctx context.Context, uid, roomID int64) (int64, error) // 连续天数
	totalDays func(ctx context.Context, uid, roomID int64) (int64, error) // 累计天数
}

// interactKind 根据互动类型返回对应的差异点描述
func (p *interactProcessor) interactKind(t enum.InteractType) (interactKind, bool) {
	switch t {
	case enum.InteractTypeEnter:
		return interactKind{
			group: "welcome", tag: "Welcome", label: "进房欢迎", danmuKind: "welcome",
			count: func(ctx context.Context, uid, roomID int64) (int64, error) {
				return p.liveInteractWord.CountEnterByUIDAndRoomID(ctx, nil, uid, roomID)
			},
			streak: func(ctx context.Context, uid, roomID int64) (int64, error) {
				return p.liveInteractWord.EnterStreakDaysByUIDAndRoomID(ctx, nil, uid, roomID)
			},
			totalDays: func(ctx context.Context, uid, roomID int64) (int64, error) {
				return p.liveInteractWord.EnterTotalDaysByUIDAndRoomID(ctx, nil, uid, roomID)
			},
		}, true
	case enum.InteractTypeFollow:
		return interactKind{
			group: "follow", tag: "Follow", label: "感谢关注", danmuKind: "follow",
			count: func(ctx context.Context, uid, roomID int64) (int64, error) {
				return p.liveInteractWord.CountFollowByUIDAndRoomID(ctx, nil, uid, roomID)
			},
			streak: func(ctx context.Context, uid, roomID int64) (int64, error) {
				return p.liveInteractWord.FollowStreakDaysByUIDAndRoomID(ctx, nil, uid, roomID)
			},
			totalDays: func(ctx context.Context, uid, roomID int64) (int64, error) {
				return p.liveInteractWord.FollowTotalDaysByUIDAndRoomID(ctx, nil, uid, roomID)
			},
		}, true
	case enum.InteractTypeShare:
		return interactKind{
			group: "share", tag: "Share", label: "感谢分享", danmuKind: "share",
			count: func(ctx context.Context, uid, roomID int64) (int64, error) {
				return p.liveInteractWord.CountShareByUIDAndRoomID(ctx, nil, uid, roomID)
			},
			streak: func(ctx context.Context, uid, roomID int64) (int64, error) {
				return p.liveInteractWord.ShareStreakDaysByUIDAndRoomID(ctx, nil, uid, roomID)
			},
			totalDays: func(ctx context.Context, uid, roomID int64) (int64, error) {
				return p.liveInteractWord.ShareTotalDaysByUIDAndRoomID(ctx, nil, uid, roomID)
			},
		}, true
	}
	return interactKind{}, false
}

// processInteractIn 处理互动消息回复：校验条件后发送答谢弹幕
func (p *interactProcessor) processInteractIn(ctx context.Context, kind interactKind, info *live.InteractWordV2Info, anchorID, botUID int64, liveStatus int) {
	if botUID == info.UID {
		return
	}
	var cfg robotconfig.InteractConfig
	if err := p.configCache.UnmarshalGroup(kind.group, &cfg); err != nil {
		log.Printf("[live.%s] 加载%s配置失败: %v", kind.tag, kind.label, err)
		return
	}
	if !ptr.ParseBool(cfg.Enabled) {
		return
	}
	if !p.checkInteractCondition(kind, cfg, info, anchorID, liveStatus) {
		return
	}
	p.sendInteractReply(ctx, kind, cfg.Content, info)
}

// checkInteractCondition 校验互动回复的 Requirement 和 Scene 是否满足
func (p *interactProcessor) checkInteractCondition(kind interactKind, cfg robotconfig.InteractConfig, info *live.InteractWordV2Info, anchorID int64, liveStatus int) bool {
	req := ptr.ParseEnumInt[enum.Requirement](cfg.Requirement)
	switch req {
	case enum.RequirementUnlimited:
	case enum.RequirementHasBadge:
		if info.BadgeUID != anchorID {
			log.Printf("[live.%s] %s条件不满足: 需要携带当前直播间牌子，用户牌子 anchorID=%d, 当前 anchorID=%d", kind.tag, kind.label, info.BadgeUID, anchorID)
			return false
		}
	case enum.RequirementHasSailBadge:
		if info.BadgeUID != anchorID {
			log.Printf("[live.%s] %s条件不满足: 需要携带当前直播间牌子，用户牌子 anchorID=%d, 当前 anchorID=%d", kind.tag, kind.label, info.BadgeUID, anchorID)
			return false
		}
		if enum.BadgeType(info.BadgeType) == enum.BadgeTypeL0 {
			log.Printf("[live.%s] %s条件不满足: 需要非 L0 勋章", kind.tag, kind.label)
			return false
		}
	default:
		log.Printf("[live.%s] 未知的%s条件: %d", kind.tag, kind.label, req)
		return false
	}
	sce := ptr.ParseEnumInt[enum.Scene](cfg.Scene)
	switch sce {
	case enum.SceneUnlimited:
	case enum.SceneLive:
		if liveStatus != int(enum.LiveStatusLive) {
			log.Printf("[live.%s] 场景不满足: 需要直播中，当前状态=%d", kind.tag, liveStatus)
			return false
		}
	case enum.SceneNotLive:
		if liveStatus == int(enum.LiveStatusLive) {
			log.Printf("[live.%s] 场景不满足: 需要非直播状态，当前为直播中", kind.tag)
			return false
		}
	default:
		log.Printf("[live.%s] 未知的场景: %d", kind.tag, sce)
		return false
	}
	return true
}

// sendInteractReply 从模板列表随机选取一条，渲染变量后发送弹幕
func (p *interactProcessor) sendInteractReply(ctx context.Context, kind interactKind, templates []string, info *live.InteractWordV2Info) {
	tmpl := PickRandom(templates)
	if tmpl == "" {
		return
	}
	needed := CollectVars(templates)
	vars := p.resolveInteractVars(ctx, kind, info, needed)
	msg := RenderTemplate(tmpl, vars)
	p.enqueueDanmu(msg, kind.danmuKind)
}

// resolveInteractVars 按需查询互动相关数据，返回变量名 → 值的映射
func (p *interactProcessor) resolveInteractVars(ctx context.Context, kind interactKind, info *live.InteractWordV2Info, needed map[string]bool) map[string]string {
	vars := make(map[string]string)
	if needed["name"] {
		vars["name"] = info.Uname
	}
	if needed["guard"] {
		vars["guard"] = enum.BadgeType(info.BadgeType).Text("zh")
	}
	if needed["total_times"] {
		if v, err := kind.count(ctx, info.UID, info.RoomID); err != nil {
			log.Printf("[live.%s] 查询累计次数失败: %v", kind.tag, err)
		} else {
			vars["total_times"] = strconv.FormatInt(v, 10)
		}
	}
	if needed["total_days"] {
		if v, err := kind.totalDays(ctx, info.UID, info.RoomID); err != nil {
			log.Printf("[live.%s] 查询累计天数失败: %v", kind.tag, err)
		} else {
			vars["total_days"] = strconv.FormatInt(v, 10)
		}
	}
	if needed["streak"] {
		if v, err := kind.streak(ctx, info.UID, info.RoomID); err != nil {
			log.Printf("[live.%s] 查询连续天数失败: %v", kind.tag, err)
		} else {
			vars["streak"] = strconv.FormatInt(v, 10)
		}
	}
	return vars
}
