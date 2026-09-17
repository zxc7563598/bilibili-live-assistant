package livepk

import (
	"context"
	"errors"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_pk_log"
)

type Service struct {
	livePkLogRepo live_pk_log.Repository
}

func New(livePkLogRepo live_pk_log.Repository) *Service {
	return &Service{
		livePkLogRepo: livePkLogRepo,
	}
}

// FetchRoomGroups 用于获取不重复的房间ID信息
func (s *Service) FetchRoomGroups(ctx context.Context) ([]FetchRoomGroupsResp, int, error) {
	roomIDs, err := s.livePkLogRepo.DistinctRoomIDs(ctx, nil)
	if err != nil {
		return []FetchRoomGroupsResp{}, 60701, err
	}
	return toFetchRoomGroupsItems(roomIDs), 0, nil
}

// ListPage 用于获取 PK 对战记录列表信息
func (s *Service) ListPage(ctx context.Context, req ListPageReq) (ListPageResp, int, error) {
	// 校验筛选参数合法性，非法值直接返回参数错误
	if req.Result != nil {
		switch *req.Result {
		case resultLose, resultWin:
		default:
			return ListPageResp{}, 11501, errors.New("result 内容非法")
		}
	}
	// 获取列表数据
	offset, limit, sortField, sortOrder := req.OffsetLimit()
	query := model.LivePkLogListPageQuery{
		RoomID:       req.RoomID,
		RivalUID:     req.RivalUID,
		RivalUname:   req.RivalUname,
		SelfResult:   req.Result,
		StartAtStart: req.StartAtStart,
		StartAtEnd:   req.StartAtEnd,
		Offset:       offset,
		Limit:        limit,
		SortField:    sortField,
		SortOrder:    sortOrder,
	}
	list, total, err := s.livePkLogRepo.ListPage(ctx, nil, query)
	if err != nil {
		return ListPageResp{}, 61501, err
	}
	totalNum, winNum, loseNum, err := s.livePkLogRepo.ListStats(ctx, nil, query)
	if err != nil {
		return ListPageResp{}, 61501, err
	}
	// 返回数据
	return ListPageResp{
		Total:    total,
		PageData: toListPageItems(list),
		Stats: ListPageStats{
			TotalNum: totalNum,
			WinNum:   winNum,
			LoseNum:  loseNum,
		},
	}, 0, nil
}
