package livegift

import (
	"context"
	"errors"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
)

type Service struct {
	liveGiftRepo live_gift.Repository
}

func New(liveGiftRepo live_gift.Repository) *Service {
	return &Service{
		liveGiftRepo: liveGiftRepo,
	}
}

// FetchRoomGroups 用于获取不重复的房间ID信息
func (s *Service) FetchRoomGroups(ctx context.Context) ([]FetchRoomGroupsResp, int, error) {
	roomIDs, err := s.liveGiftRepo.DistinctRoomIDs(ctx, nil)
	if err != nil {
		return []FetchRoomGroupsResp{}, 60701, err
	}
	return toFetchRoomGroupsItems(roomIDs), 0, nil
}

// ListPage 用于获取礼物列表信息
func (s *Service) ListPage(ctx context.Context, req ListPageReq) (ListPageResp, int, error) {
	// 校验枚举参数合法性，非法值直接返回参数错误
	if req.GiftType != nil && !enum.GiftType(*req.GiftType).IsValid() {
		return ListPageResp{}, 10601, errors.New("gift_type 内容非法")
	}
	if req.Original != nil && !enum.YesNo(*req.Original).IsValid() {
		return ListPageResp{}, 10601, errors.New("original 内容非法")
	}
	// 获取列表数据
	offset, limit := req.OffsetLimit()
	queue := model.LiveGiftListPageQuery{
		RoomID:      req.RoomID,
		UID:         req.UID,
		Uname:       req.Uname,
		GiftName:    req.GiftName,
		GiftType:    req.GiftType,
		Original:    req.Original,
		SendAtStart: req.SendAtStart,
		SendAtEnd:   req.SendAtEnd,
		Offset:      offset,
		Limit:       limit,
	}
	listGift, total, err := s.liveGiftRepo.ListPage(ctx, nil, queue)
	if err != nil {
		return ListPageResp{}, 60701, err
	}
	totalNum, totalAmount, err := s.liveGiftRepo.ListStats(ctx, nil, queue)
	if err != nil {
		return ListPageResp{}, 60701, err
	}
	// 返回数据
	return ListPageResp{
		Total:    total,
		PageData: toListPageItems(listGift),
		Stats: ListPageStats{
			TotalNum:    totalNum,
			TotalAmount: totalAmount,
		},
	}, 0, nil
}

// BlindBoxListPage 用于获取盲盒礼物列表信息
func (s *Service) BlindBoxListPage(ctx context.Context, req BlindBoxListPageReq) (BlindBoxListPageResp, int, error) {
	// 获取列表数据
	offset, limit := req.OffsetLimit()
	queue := model.LiveGiftBlindBoxListPageQuery{
		RoomID:           req.RoomID,
		UID:              req.UID,
		Uname:            req.Uname,
		GiftName:         req.GiftName,
		OriginalGiftName: req.OriginalGiftName,
		SendAtStart:      req.SendAtStart,
		SendAtEnd:        req.SendAtEnd,
		Offset:           offset,
		Limit:            limit,
	}
	listGift, total, err := s.liveGiftRepo.BlindBoxListPage(ctx, nil, queue)
	if err != nil {
		return BlindBoxListPageResp{}, 60701, err
	}
	originalPrice, currentPrice, err := s.liveGiftRepo.BlindBoxListStats(ctx, nil, queue)
	if err != nil {
		return BlindBoxListPageResp{}, 60701, err
	}
	// 返回数据
	return BlindBoxListPageResp{
		Total:    total,
		PageData: toBlindBoxListPageItems(listGift),
		Stats: BlindBoxListPageStats{
			OriginalPrice: originalPrice,
			CurrentPrice:  currentPrice,
		},
	}, 0, nil
}
