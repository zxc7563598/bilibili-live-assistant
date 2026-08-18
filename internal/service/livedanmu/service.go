package livedanmu

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
)

type Service struct {
	liveDanmuRepo live_danmu.Repository
}

func New(liveDanmuRepo live_danmu.Repository) *Service {
	return &Service{
		liveDanmuRepo: liveDanmuRepo,
	}
}

// FetchRoomGroups 用于获取不重复的房间ID信息
func (s *Service) FetchRoomGroups(ctx context.Context) ([]FetchRoomGroupsResp, int, error) {
	roomIDs, err := s.liveDanmuRepo.DistinctRoomIDs(ctx, nil)
	if err != nil {
		return []FetchRoomGroupsResp{}, 60601, err
	}
	return toFetchRoomGroupItems(roomIDs), 0, nil
}

// ListPage 用于获取弹幕列表信息
func (s *Service) ListPage(ctx context.Context, req ListPageReq) (ListPageResp, int, error) {
	// 获取列表数据
	offset, limit := req.OffsetLimit()
	listDanmu, total, err := s.liveDanmuRepo.ListPage(ctx, nil, model.LiveDanmuListPageQuery{
		RoomID:      req.RoomID,
		UID:         req.UID,
		Uname:       req.Uname,
		Msg:         req.Msg,
		SendAtStart: req.SendAtStart,
		SendAtEnd:   req.SendAtEnd,
		Offset:      offset,
		Limit:       limit,
	})
	if err != nil {
		return ListPageResp{}, 60601, err
	}
	// 返回数据
	return ListPageResp{
		Total:    total,
		PageData: toListPageItems(listDanmu),
	}, 0, nil
}
