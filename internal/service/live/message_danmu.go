package live

import (
	"context"
	"log"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/live"
)

// danmuProcessor 处理弹幕消息（DANMU_MSG）
type danmuProcessor struct {
	liveDanmuRepo live_danmu.Repository
	roomState     *RoomState
	getBotUID     func() int64
}

func newDanmuProcessor(liveDanmuRepo live_danmu.Repository, roomState *RoomState, getBotUID func() int64) *danmuProcessor {
	return &danmuProcessor{
		liveDanmuRepo: liveDanmuRepo,
		roomState:     roomState,
		getBotUID:     getBotUID,
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
	// 签到检测
	botUID := p.getBotUID()
	liveStatus := p.roomState.LiveStatus()
	_ = botUID
	_ = liveStatus

	// 自动回复关键词检测
	return nil
}
