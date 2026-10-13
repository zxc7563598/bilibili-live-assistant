package live

import (
	"context"
	"log"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_interact_word"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/util"
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
	group     string                                                                 // 配置分组名：welcome / follow / share
	tag       string                                                                 // 日志标签：Welcome / Follow / Share
	label     string                                                                 // 中文说明：进房欢迎 / 感谢关注 / 感谢分享
	danmuKind string                                                                 // 弹幕发送分类
	count     func(ctx context.Context, uid, roomID int64) (int64, error)            // 累计次数
	streak    func(ctx context.Context, uid, roomID int64) (int64, error)            // 连续天数
	totalDays func(ctx context.Context, uid, roomID int64) (int64, error)            // 累计天数
	last      func(ctx context.Context, uid, roomID, now int64) (int64, bool, error) // 距上次行为的秒数
	lastVar   string                                                                 // 距上次行为的模板变量名：last_enter / last_follow / last_share
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
			lastVar: "last_enter",
			last: func(ctx context.Context, uid, roomID, now int64) (int64, bool, error) {
				return p.liveInteractWord.SecondsSinceLastEnterByUIDAndRoomID(ctx, nil, uid, roomID, now)
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
			lastVar: "last_follow",
			last: func(ctx context.Context, uid, roomID, now int64) (int64, bool, error) {
				return p.liveInteractWord.SecondsSinceLastFollowByUIDAndRoomID(ctx, nil, uid, roomID, now)
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
			lastVar: "last_share",
			last: func(ctx context.Context, uid, roomID, now int64) (int64, bool, error) {
				return p.liveInteractWord.SecondsSinceLastShareByUIDAndRoomID(ctx, nil, uid, roomID, now)
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
	if !configEnabled(p.configCache, kind.group) {
		return
	}
	var interactCfg robotconfig.InteractConfig
	if err := p.configCache.UnmarshalGroup(kind.group, &interactCfg); err != nil {
		log.Printf("[live.%s] 加载%s配置失败: %v", kind.tag, kind.label, err)
		return
	}
	if p.interactIntervalHit(ctx, kind, interactCfg.Interval, info) {
		return
	}
	if hit, ok := blacklistHit(info.Uname, interactCfg.Blacklist); ok {
		log.Printf("[live.%s] %s黑名单命中，跳过回复: 用户名=%q 命中规则=%q", kind.tag, kind.label, info.Uname, hit)
		return
	}
	if !p.checkInteractCondition(kind, interactCfg, info, anchorID, liveStatus) {
		return
	}
	p.sendInteractReply(ctx, kind, interactCfg.Content, info)
}

// interactIntervalHit 判断用户在该模块的响应间隔内是否刚触发过同类行为, 命中则不发送
//
// 间隔为 0 或解析失败表示不限制；查询失败时放行, 不因读库出错而吞掉答谢。
func (p *interactProcessor) interactIntervalHit(ctx context.Context, kind interactKind, interval string, info *live.InteractWordV2Info) bool {
	seconds, err := strconv.Atoi(interval)
	if err != nil || seconds <= 0 {
		return false
	}
	elapsed, found, err := kind.last(ctx, info.UID, info.RoomID, info.Timestamp)
	if err != nil {
		log.Printf("[live.%s] 查询响应间隔失败: %v", kind.tag, err)
		return false
	}
	if !found || elapsed >= int64(seconds) {
		return false
	}
	log.Printf("[live.%s] %s响应间隔内重复触发，跳过回复: 用户名=%q 间隔=%d秒 距上次=%d秒", kind.tag, kind.label, info.Uname, seconds, elapsed)
	return true
}

// blacklistHit 返回昵称命中的第一条黑名单规则
func blacklistHit(nickname string, rules []string) (string, bool) {
	for _, rule := range rules {
		if nicknameMatched(nickname, rule) {
			return rule, true
		}
	}
	return "", false
}

// nicknameMatched 判断昵称是否命中单条黑名单规则, 规则中的 * 匹配任意长度(含 0)的任意字符
func nicknameMatched(nickname, pattern string) bool {
	if pattern == "" {
		return false
	}
	if !strings.Contains(pattern, "*") {
		return nickname == pattern
	}
	segs := strings.Split(pattern, "*")
	pos := 0
	if segs[0] != "" {
		if !strings.HasPrefix(nickname, segs[0]) {
			return false
		}
		pos = len(segs[0])
	}
	last := len(segs) - 1
	for _, seg := range segs[1:last] {
		if seg == "" {
			continue
		}
		idx := strings.Index(nickname[pos:], seg)
		if idx < 0 {
			return false
		}
		pos += idx + len(seg)
	}
	if segs[last] == "" {
		return true
	}
	return strings.HasSuffix(nickname[pos:], segs[last])
}

// checkInteractCondition 校验互动回复的 Requirement 和 Scene 是否满足
func (p *interactProcessor) checkInteractCondition(kind interactKind, interactCfg robotconfig.InteractConfig, info *live.InteractWordV2Info, anchorID int64, liveStatus int) bool {
	req := ptr.ParseEnumInt[enum.Requirement](interactCfg.Requirement)
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
	sce := ptr.ParseEnumInt[enum.Scene](interactCfg.Scene)
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
	var roomCfg robotconfig.RoomConfig
	if err := p.configCache.UnmarshalGroup("room", &roomCfg); err != nil {
		log.Printf("[live.%s] 加载房间配置失败: %v", kind.tag, err)
		return
	}
	needed := CollectVars([]string{tmpl})
	vars := p.resolveInteractVars(ctx, kind, info, needed, roomCfg)
	msg := RenderTemplate(tmpl, vars)
	p.enqueueDanmu(msg, kind.danmuKind)
}

// resolveInteractVars 按需查询互动相关数据，返回变量名 → 值的映射
func (p *interactProcessor) resolveInteractVars(ctx context.Context, kind interactKind, info *live.InteractWordV2Info, needed map[string]bool, roomCfg robotconfig.RoomConfig) map[string]string {
	vars := make(map[string]string)
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
	if needed[kind.lastVar] {
		if v, ok, err := kind.last(ctx, info.UID, info.RoomID, info.Timestamp); err != nil {
			log.Printf("[live.%s] 查询上次时间失败: %v", kind.tag, err)
		} else if ok {
			vars[kind.lastVar] = formatElapsed(v)
		} else {
			vars[kind.lastVar] = lastNoneText
		}
	}
	return vars
}

// lastNoneText 该用户没有任何历史记录（本次是第一条）时的兜底文案
const lastNoneText = "无记录"

// formatElapsed 将秒数格式化为秒/分钟/小时/天四档文本, 向下取整
func formatElapsed(seconds int64) string {
	switch {
	case seconds < 60:
		return strconv.FormatInt(seconds, 10) + "秒"
	case seconds < 3600:
		return strconv.FormatInt(seconds/60, 10) + "分钟"
	case seconds < 86400:
		return strconv.FormatInt(seconds/3600, 10) + "小时"
	default:
		return strconv.FormatInt(seconds/86400, 10) + "天"
	}
}
