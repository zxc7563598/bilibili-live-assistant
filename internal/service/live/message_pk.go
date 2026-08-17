package live

import (
	"context"
	"log"
	"strconv"
	"unicode/utf8"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/util"
)

// pkProcessor 处理 PK 相关消息（PK_BATTLE_PRE_NEW）
type pkProcessor struct {
	configCache  *robotconfig.Cache
	client       *bilibili.Client
	getBotUID    func() int64
	enqueueDanmu func(msg string, kind string)
}

func newPkProcessor(configCache *robotconfig.Cache, client *bilibili.Client, getBotUID func() int64, enqueueDanmu func(msg string, kind string)) *pkProcessor {
	return &pkProcessor{
		configCache:  configCache,
		client:       client,
		getBotUID:    getBotUID,
		enqueueDanmu: enqueueDanmu,
	}
}

func (p *pkProcessor) Cmds() []live.Cmd {
	return []live.Cmd{live.CmdPkStart}
}

// Process 处理PK播报逻辑
func (p *pkProcessor) Process(ctx context.Context, cmd live.Cmd, data any, roomID int64) error {
	info, ok := data.(*live.PkBattlePreNewInfo)
	if !ok {
		log.Printf("[live.PK] 数据类型断言失败，期望 *live.PkBattlePreNewInfo，实际 %T", data)
		return nil
	}
	botUID := p.getBotUID()
	p.processPkIn(ctx, info, botUID)
	return nil
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
