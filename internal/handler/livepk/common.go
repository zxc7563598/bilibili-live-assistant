package livepk

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livepk"
)

// Handler PK 对战记录 HTTP 接口处理器
type Handler struct {
	livepkSvc *livepk.Service
}

// New 创建 Handler 实例
func New(livepkSvc *livepk.Service) *Handler {
	return &Handler{livepkSvc: livepkSvc}
}

// toLivePkLogListItems PK 对战记录转换：Service 出参 → 响应结构
func toLivePkLogListItems(list []livepk.ListPageItem) []resp.LivePkLogListPageItem {
	res := make([]resp.LivePkLogListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.LivePkLogListPageItem{
			ID:          v.ID,
			RoomID:      v.RoomID,
			PkID:        v.PkID,
			PkStatus:    v.PkStatus,
			BattleType:  v.BattleType,
			MatchType:   v.MatchType,
			RivalUID:    v.RivalUID,
			RivalUname:  v.RivalUname,
			RivalRoomID: v.RivalRoomID,
			SelfVotes:   v.SelfVotes,
			RivalVotes:  v.RivalVotes,
			SelfResult:  v.SelfResult,
			RivalResult: v.RivalResult,
			StartAt:     v.StartAt,
			SettleAt:    v.SettleAt,
		})
	}
	return res
}
