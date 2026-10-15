package liveguard

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveguard"
)

// Handler 大航海名单 HTTP 接口处理器
type Handler struct {
	liveguardSvc *liveguard.Service
}

// New 创建 Handler 实例
func New(liveguardSvc *liveguard.Service) *Handler {
	return &Handler{liveguardSvc: liveguardSvc}
}

func toLiveGuardListItems(list []liveguard.ListPageItem) []resp.LiveGuardListPageItem {
	res := make([]resp.LiveGuardListPageItem, 0, len(list))
	for _, v := range list {
		res = append(res, resp.LiveGuardListPageItem{
			UID:            v.UID,
			Uname:          v.Uname,
			GuardLevel:     v.GuardLevel,
			CoverFromMonth: v.CoverFromMonth,
			CoverToMonth:   v.CoverToMonth,
			OpenNum:        v.OpenNum,
			OpenAt:         v.OpenAt,
		})
	}
	return res
}
