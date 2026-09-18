package livedanmu

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/livedanmu"
)

// Handler 直播控制 HTTP 接口处理器
type Handler struct {
	livedanmuSvc *livedanmu.Service
}

// New 创建 Handler 实例
func New(livedanmuSvc *livedanmu.Service) *Handler {
	return &Handler{livedanmuSvc: livedanmuSvc}
}

func toLiveDanmuListItems(list []livedanmu.ListPageItem) []resp.LiveDanmuListPageItem {
	res := make([]resp.LiveDanmuListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.LiveDanmuListPageItem{
			ID:          v.ID,
			UID:         v.UID,
			Uname:       v.Uname,
			Msg:         v.Msg,
			BadgeRoomID: v.BadgeRoomID,
			BadgeName:   v.BadgeName,
			BadgeLevel:  v.BadgeLevel,
			BadgeType:   v.BadgeType,
			SendAt:      v.SendAt,
		})
	}
	return res
}
