package liveguard

import (
	"context"
	"errors"
	"log"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_guard_record"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/live"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
)

type Service struct {
	liveGuardRepo live_guard_record.Repository
	liveGiftRepo  live_gift.Repository
	liveUserSvc   *liveuser.Service
	liveSvc       *live.Service
}

func New(liveGuardRepo live_guard_record.Repository, liveGiftRepo live_gift.Repository, liveUserSvc *liveuser.Service, liveSvc *live.Service) *Service {
	return &Service{
		liveGuardRepo: liveGuardRepo,
		liveGiftRepo:  liveGiftRepo,
		liveUserSvc:   liveUserSvc,
		liveSvc:       liveSvc,
	}
}

// ListPage 用于获取某月某档位的大航海名单
func (s *Service) ListPage(ctx context.Context, req ListPageReq) (ListPageResp, int, error) {
	if req.GuardLevel == nil {
		return ListPageResp{}, CodeGuardLevelInvalid, errors.New("guard_level 未传")
	}
	level, ok := parseGuardLevel(*req.GuardLevel)
	if !ok {
		return ListPageResp{}, CodeGuardLevelInvalid, errors.New("guard_level 内容非法")
	}
	from, to, err := monthRange(req.Month)
	if err != nil {
		return ListPageResp{}, CodeParamInvalid, err
	}
	offset, limit, sortField, sortOrder := req.OffsetLimit()
	query := model.LiveGuardRecordListPageQuery{
		GuardLevel: req.GuardLevel,
		MonthFrom:  from,
		MonthTo:    to,
		UID:        req.UID,
		Uname:      req.Uname,
		Offset:     offset,
		Limit:      limit,
		SortField:  sortField,
		SortOrder:  sortOrder,
	}
	list, total, err := s.liveGuardRepo.ListRosterPage(ctx, nil, query)
	if err != nil {
		return ListPageResp{}, CodeQueryFailed, err
	}
	items := toListPageItems(list, level)

	// 当月开通份数与最早开通时间另查一次，只作用于本页用户
	uids := make([]int64, 0, len(items))
	for _, item := range items {
		uids = append(uids, item.UID)
	}
	stats, err := s.liveGuardRepo.ListOpenStats(ctx, nil, model.LiveGuardRecordOpenStatsQuery{
		GuardLevel: req.GuardLevel,
		OpenFrom:   from,
		OpenTo:     to,
		UIDs:       uids,
	})
	if err != nil {
		return ListPageResp{}, CodeQueryFailed, err
	}
	indexByUID := make(map[int64]int, len(items))
	for i, item := range items {
		indexByUID[item.UID] = i
	}
	for _, stat := range stats {
		if i, ok := indexByUID[stat.UID]; ok {
			items[i].OpenNum = stat.OpenNum
			items[i].OpenAt = formatUnix(stat.OpenAt)
		}
	}
	return ListPageResp{Total: total, PageData: items}, 0, nil
}

// Backfill 补录一次大航海开通
//
// 视作系统漏记了这条记录：把礼物记录、累计消费金额、到期时间、开通台账与奖励
// 一并补齐，与直播链路同源同口径。不发弹幕、不走黑名单赎回 —— 补录的是历史，
// 往直播间补发一条感谢弹幕没有意义。
//
// 名单月份区间按台账重排（见 planGuardRestack），不按用户当前到期时间顺延：
// 补录可以插在任意历史位置，拿「现在有效到什么时候」当基准会直接串位。
//
// 全程逐步提交、不开事务：直播链路本身就是非事务的（每个子步骤各自提交），
// 而 sqlite 连接池固定为 1，外层开事务会让内部自开事务的调用互等连接直到超时。
// 中途失败会留下部分状态，与直播链路同等程度的接受，日志里会记清是哪一步。
func (s *Service) Backfill(ctx context.Context, req BackfillReq) (int, error) {
	level, ok := parseGuardLevel(req.GuardLevel)
	if !ok {
		return CodeGuardLevelInvalid, errors.New("guard_level 内容非法")
	}
	giftName, ok := enum.GuardNameByBadgeType(level)
	if !ok {
		return CodeGuardLevelInvalid, errors.New("guard_level 无对应礼物名称")
	}
	// 礼物ID与价格按档位取，不由调用方传：三个航海礼物是固定档位
	giftID, price, ok := enum.GuardGiftByBadgeType(level)
	if !ok {
		return CodeGuardLevelInvalid, errors.New("guard_level 无对应礼物ID与价格")
	}
	num := req.Num
	if num < 1 {
		num = 1
	}
	openAt := req.OpenAt
	if openAt <= 0 {
		return CodeParamInvalid, errors.New("open_at 非法")
	}
	// 注册用户并取内部 ID
	userID, err := s.liveUserSvc.EnsureUser(ctx, req.UID, req.Uname)
	if err != nil {
		return CodeBackfillFailed, err
	}
	// 名单月份区间：并入台账序列重算，同时拿到需要跟着后移的后续记录
	restack, errCode, err := s.planGuardRestack(ctx, req.UID, level, openAt, num)
	if errCode != 0 {
		return errCode, err
	}
	// 真实礼物记录
	if _, err := s.liveGiftRepo.Create(ctx, nil, &model.LiveGift{
		RoomID:   req.RoomID,
		UID:      req.UID,
		Uname:    req.Uname,
		GiftType: enum.GiftTypeGuard,
		GiftID:   giftID,
		GiftName: giftName,
		Price:    price,
		Num:      num,
		SendAt:   openAt,
		Original: enum.Yes,
	}); err != nil {
		return CodeBackfillFailed, err
	}
	// 累计消费金额
	if err := s.liveUserSvc.AddTotalGiftAmount(ctx, userID, price*num); err != nil {
		return CodeBackfillFailed, err
	}
	// 开通台账：先把新记录插进去，再后移受它影响的后续记录。
	// 顺序不能反：中途失败时「多一条区间正确的新记录」比「旧记录被后移但新记录没写进去」好收拾
	if _, err := s.liveGuardRepo.Create(ctx, nil, &model.LiveGuardRecord{
		RoomID:     req.RoomID,
		UID:        req.UID,
		Uname:      req.Uname,
		GuardLevel: level,
		GiftID:     giftID,
		Price:      price,
		Num:        num,
		OpenAt:     openAt,
		CoverFrom:  restack.CoverFrom,
		CoverTo:    restack.CoverTo,
	}); err != nil {
		return CodeBackfillFailed, err
	}
	for _, shift := range restack.Shifts {
		if err := s.liveGuardRepo.UpdateCoverByID(ctx, nil, shift.ID, shift.CoverFrom, shift.CoverTo); err != nil {
			return CodeBackfillFailed, err
		}
	}
	// 到期时间：另外一份口径，只往前推，不回退
	if errCode, err := s.liveUserSvc.ExtendGuardExpireTo(ctx, req.UID, req.Uname, level, restack.ExpireAt); errCode != 0 {
		return CodeBackfillFailed, err
	}
	// 积分/星光奖励，与直播链路同一份金额口径
	if err := s.liveSvc.GrantGiftReward(ctx, userID, giftName, enum.GiftTypeGuard, level, price, num); err != nil {
		return CodeBackfillFailed, err
	}
	log.Printf("[liveguard] 补录大航海成功: uid=%d guardLevel=%d openAt=%d", req.UID, req.GuardLevel, openAt)
	return 0, nil
}
