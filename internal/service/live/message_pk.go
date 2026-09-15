package live

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"unicode/utf8"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_pk_log"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/util"
)

// pkProcessor 处理 PK 相关消息（PK_BATTLE_PRE_NEW / PK_BATTLE_END / PK_BATTLE_SETTLE_NEW）
type pkProcessor struct {
	roomState    *RoomState
	pkLogRepo    live_pk_log.Repository
	configCache  *robotconfig.Cache
	client       *bilibili.Client
	getBotUID    func() int64
	enqueueDanmu func(msg string, kind string)
}

func newPkProcessor(roomState *RoomState, pkLogRepo live_pk_log.Repository, configCache *robotconfig.Cache, client *bilibili.Client, getBotUID func() int64, enqueueDanmu func(msg string, kind string)) *pkProcessor {
	return &pkProcessor{
		roomState:    roomState,
		pkLogRepo:    pkLogRepo,
		configCache:  configCache,
		client:       client,
		getBotUID:    getBotUID,
		enqueueDanmu: enqueueDanmu,
	}
}

func (p *pkProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdPkStart, live.CmdPkBattleEnd, live.CmdPkSettleNew}
}

// pkBattleSnapshot 是 PK_BATTLE_END / PK_BATTLE_SETTLE_NEW 两个 payload 的共有字段
type pkBattleSnapshot struct {
	Cmd        live.Cmd         // 来源事件
	PkID       int64            // PK ID
	PkStatus   int64            // PK 状态
	Timestamp  int64            // 事件时间（秒级）
	BattleType int64            // 对战类型
	InitInfo   *live.PkSideInfo // 发起方
	MatchInfo  *live.PkSideInfo // 被匹配方
}

// Process 处理PK相关消息
func (p *pkProcessor) Process(ctx context.Context, cmd live.Cmd, data any, roomID int64) error {
	var snapshot *pkBattleSnapshot
	switch cmd {
	case live.CmdPkStart:
		info, ok := data.(*live.PkBattlePreNewInfo)
		if !ok {
			log.Printf("[live.PK] 数据类型断言失败，期望 *live.PkBattlePreNewInfo，实际 %T", data)
			return nil
		}
		p.recordPkStart(ctx, info)
		botUID := p.getBotUID()
		p.processPkIn(ctx, info, botUID)
		return nil
	case live.CmdPkBattleEnd:
		info, ok := data.(*live.PkBattleEndInfo)
		if !ok {
			log.Printf("[live.PK] 数据类型断言失败，期望 *live.PkBattleEndInfo，实际 %T", data)
			return nil
		}
		snapshot = &pkBattleSnapshot{
			Cmd:        cmd,
			PkID:       info.PkID,
			PkStatus:   info.PkStatus,
			Timestamp:  info.Timestamp,
			BattleType: info.BattleType,
			InitInfo:   info.InitInfo,
			MatchInfo:  info.MatchInfo,
		}
	case live.CmdPkSettleNew:
		info, ok := data.(*live.PkBattleSettleNewInfo)
		if !ok {
			log.Printf("[live.PK] 数据类型断言失败，期望 *live.PkBattleSettleNewInfo，实际 %T", data)
			return nil
		}
		snapshot = &pkBattleSnapshot{
			Cmd:        cmd,
			PkID:       info.PkID,
			PkStatus:   info.PkStatus,
			Timestamp:  info.Timestamp,
			BattleType: info.BattleType,
			InitInfo:   info.InitInfo,
			MatchInfo:  info.MatchInfo,
		}
	}
	if snapshot != nil {
		p.processPkSettle(ctx, snapshot)
	}
	return nil
}

// recordPkStart PK 开始事件落库：一场 PK 一条记录，结束/结算事件按 pk_id 回填
func (p *pkProcessor) recordPkStart(ctx context.Context, info *live.PkBattlePreNewInfo) {
	entry := &model.LivePkLog{
		RoomID:      p.roomState.RoomID(),
		PkID:        info.PkID,
		PkStatus:    info.PkStatus,
		BattleType:  info.BattleType,
		MatchType:   info.MatchType,
		RivalUID:    info.UID,
		RivalUname:  info.Uname,
		RivalRoomID: info.RoomID,
		StartAt:     info.Timestamp,
	}
	if _, err := p.pkLogRepo.Create(ctx, nil, entry); err != nil {
		log.Printf("[live.PK] PK 记录写入失败 (pk_id=%d): %v", info.PkID, err)
	}
}

// processPkSettle PK 结束/结算事件：按 pk_id 找到开始记录，回填状态与双方比分
func (p *pkProcessor) processPkSettle(ctx context.Context, s *pkBattleSnapshot) {
	ourRoomID := p.roomState.RoomID()
	ourSide, rivalSide, ok := locatePkSides(ourRoomID, s.InitInfo, s.MatchInfo)
	entry, err := p.pkLogRepo.GetByPkID(ctx, nil, s.PkID)
	if err != nil {
		log.Printf("[live.PK] [%s] 查询 PK 记录失败 (pk_id=%d): %v", s.Cmd, s.PkID, err)
		return
	}
	if entry == nil {
		log.Printf("[live.PK] [%s] 未找到 PK 开始记录，跳过 (pk_id=%d)", s.Cmd, s.PkID)
		return
	}
	entry.PkStatus = s.PkStatus
	entry.BattleType = s.BattleType
	entry.SettleAt = s.Timestamp
	if ok {
		entry.SelfVotes = ourSide.Votes
		entry.SelfResult = pkSideResult(ourSide)
		if rivalSide != nil {
			entry.RivalVotes = rivalSide.Votes
			entry.RivalResult = pkSideResult(rivalSide)
		}
	}
	if err := p.pkLogRepo.Update(ctx, nil, entry); err != nil {
		log.Printf("[live.PK] [%s] PK 记录回填失败 (pk_id=%d): %v", s.Cmd, s.PkID, err)
		return
	}
	if !ok {
		log.Printf("[live.PK] [%s] pk_id=%d our_room=%d 未能定位本直播间，只回填了状态，原始 init(%s) match(%s)",
			s.Cmd, s.PkID, ourRoomID, pkSideFields(s.InitInfo), pkSideFields(s.MatchInfo))
		return
	}
	log.Printf("[live.PK] [%s] pk_id=%d 已回填 我方(%s) 对方(%s)", s.Cmd, s.PkID, pkSideFields(ourSide), pkSideFields(rivalSide))
}

// locatePkSides 依据己方真实房间号，从 init_info / match_info 中定位我方与对方
func locatePkSides(ourRoomID int64, initInfo, matchInfo *live.PkSideInfo) (ourSide, rivalSide *live.PkSideInfo, ok bool) {
	if ourRoomID <= 0 {
		return nil, nil, false
	}
	if initInfo != nil && initInfo.RoomID == ourRoomID {
		return initInfo, matchInfo, true
	}
	if matchInfo != nil && matchInfo.RoomID == ourRoomID {
		return matchInfo, initInfo, true
	}
	return nil, nil, false
}

// pkSideResult 取单方的胜负标记：2=获胜 -1=落败，没有则返回 0
//
// 两个事件对它的命名不同 —— 经典PK（PK_BATTLE_END）用 winner_type，大乱斗（PK_BATTLE_SETTLE_NEW）用 result_type，
// 实测两者互斥、另一个恒为 0，所以这里取到谁算谁，收敛成一个值再落库。
func pkSideResult(s *live.PkSideInfo) int64 {
	if s == nil {
		return 0
	}
	if s.ResultType != 0 {
		return s.ResultType
	}
	return s.WinnerType
}

// pkSideFields 把单方数据格式化成日志片段
func pkSideFields(s *live.PkSideInfo) string {
	if s == nil {
		return "room=0 votes=0 winner_type=0 result_type=0"
	}
	return fmt.Sprintf("room=%d votes=%d winner_type=%d result_type=%d",
		s.RoomID, s.Votes, s.WinnerType, s.ResultType)
}

func (p *pkProcessor) processPkIn(ctx context.Context, info *live.PkBattlePreNewInfo, botUID int64) {
	if botUID == info.UID {
		return
	}
	var pkCfg robotconfig.PkConfig
	if err := p.configCache.UnmarshalGroup("pk", &pkCfg); err != nil {
		log.Printf("[live.PK] 加载PK播报配置失败: %v", err)
		return
	}
	if !ptr.ParseBool(pkCfg.Enabled) {
		return
	}
	p.sendPkReply(ctx, pkCfg.Content, info)
}

// sendPkReply PK播报不随机抽取，渲染变量后发送全部模板
func (p *pkProcessor) sendPkReply(ctx context.Context, templates []string, info *live.PkBattlePreNewInfo) {
	var roomCfg robotconfig.RoomConfig
	if err := p.configCache.UnmarshalGroup("room", &roomCfg); err != nil {
		log.Printf("[live.PK] 加载房间配置失败: %v", err)
		return
	}
	needed := CollectVars(templates)
	vars := p.resolvePkVars(ctx, info, needed, roomCfg)
	msg := make([]string, 0, len(templates))
	for _, template := range templates {
		msg = append(msg, RenderTemplate(template, vars))
	}
	for _, message := range msg {
		p.enqueueDanmu(message, "pk")
	}
}

// resolvePkVars 按需查询PK播报相关数据，返回变量名 → 值的映射
func (p *pkProcessor) resolvePkVars(ctx context.Context, info *live.PkBattlePreNewInfo, needed map[string]bool, roomCfg robotconfig.RoomConfig) map[string]string {
	vars := make(map[string]string)
	if needed["anchor"] {
		vars["anchor"] = info.Uname
		maxNameLengthValue, _ := strconv.ParseInt(roomCfg.MaxNameLength, 10, 64)
		if maxNameLengthValue > 0 {
			if utf8.RuneCountInString(vars["anchor"]) > int(maxNameLengthValue) {
				switch ptr.ParseEnumInt[enum.NameTrimMode](roomCfg.NameTrimMode) {
				case enum.NameTrimModeTrimEnd:
					vars["anchor"] = util.TrimFromBack(vars["anchor"], int(maxNameLengthValue))
				case enum.NameTrimModeTrimStart:
					vars["anchor"] = util.TrimFromFront(vars["anchor"], int(maxNameLengthValue))
				}
			}
		}
	}
	if needed["online_num"] || needed["online_score"] || needed["top3_score"] {
		onlineGoldRank, err := p.client.Room.GetOnlineGoldRank(ctx, info.UID, info.RoomID)
		if err != nil {
			log.Printf("[live.PK] 获取直播间在线金瓜子榜失败: %v", err)
		} else {
			if needed["online_num"] {
				vars["online_num"] = strconv.Itoa(onlineGoldRank.OnlineNum)
			}
			if needed["online_score"] {
				var sum int64
				for _, item := range onlineGoldRank.Items {
					sum += item.Score
				}
				vars["online_score"] = strconv.FormatInt(sum, 10)
			}
			if needed["top3_score"] {
				items := onlineGoldRank.Items
				if len(items) > 3 {
					items = items[:3]
				}
				var sum int64
				for _, item := range items {
					sum += item.Score
				}
				vars["top3_score"] = strconv.FormatInt(sum, 10)
			}
		}
	}
	if needed["vip_num"] {
		vipNumbers, err := p.client.Room.GetVipNumbers(ctx, info.UID, info.RoomID)
		if err != nil {
			log.Printf("[live.PK] 获取直播间大航海总人数失败: %v", err)
		} else {
			vars["vip_num"] = strconv.Itoa(vipNumbers)
		}
	}
	return vars
}
