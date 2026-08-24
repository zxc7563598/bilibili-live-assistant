package liveuser

import (
	"context"
	"fmt"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_session"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_credit_log"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/tokenizer"
	"gorm.io/gorm"
)

type Service struct {
	db                    *gorm.DB
	liveUserRepo          live_user.Repository
	liveUserCreditLogRepo live_user_credit_log.Repository
	liveDanmuRepo         live_danmu.Repository
	liveGiftRepo          live_gift.Repository
	liveSessionRepo       live_session.Repository
}

const userDanmuAnalysisLimit = 20

func New(db *gorm.DB, liveUserRepo live_user.Repository, liveUserCreditLogRepo live_user_credit_log.Repository, liveDanmuRepo live_danmu.Repository, liveGiftRepo live_gift.Repository, liveSessionRepo live_session.Repository) *Service {
	return &Service{
		db:                    db,
		liveUserRepo:          liveUserRepo,
		liveUserCreditLogRepo: liveUserCreditLogRepo,
		liveDanmuRepo:         liveDanmuRepo,
		liveGiftRepo:          liveGiftRepo,
		liveSessionRepo:       liveSessionRepo,
	}
}

// ListPage 用于获取用户列表信息
func (s *Service) ListPage(ctx context.Context, req ListPageReq) (ListPageResp, int, error) {
	// 获取列表数据
	offset, limit := req.OffsetLimit()
	listDanmu, total, err := s.liveUserRepo.ListPage(ctx, nil, model.LiveUserListPageQuery{
		UID:    req.UID,
		Uname:  req.Uname,
		Offset: offset,
		Limit:  limit,
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

// GetUserMonthlyAnalysis 获取用户每月数据
func (s *Service) GetUserMonthlyAnalysis(ctx context.Context, UID, year, month int64) (GetUserMonthlyAnalysisResp, int, error) {
	// 校验年月参数，避免 time.Date 对非法值静默归一化
	if year < 1970 || year > 2100 || month < 1 || month > 12 {
		return GetUserMonthlyAnalysisResp{}, 10801, fmt.Errorf("非法的年月参数: year=%d, month=%d", year, month)
	}
	// 确定查询时间范围
	start := time.Date(int(year), time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	startTimestamp := start.Unix()
	endTimestamp := end.Unix()
	// 获取数据
	danmu, err := s.liveDanmuRepo.CountDailyByUID(ctx, nil, UID, startTimestamp, endTimestamp)
	if err != nil {
		return GetUserMonthlyAnalysisResp{}, 60801, err
	}
	gift, err := s.liveGiftRepo.CountDailyByUID(ctx, nil, UID, startTimestamp, endTimestamp)
	if err != nil {
		return GetUserMonthlyAnalysisResp{}, 60801, err
	}
	giftCount := make(map[int64]int64, len(gift))
	giftAmount := make(map[int64]int64, len(gift))
	for day, item := range gift {
		giftCount[int64(day)] = item.Num
		giftAmount[int64(day)] = item.Amount
	}
	// 本月开播记录（不区分房间ID/主播，按开播时间 start_at 落在当月统计）
	liveDays, err := s.liveSessionRepo.DistinctLiveDays(ctx, nil, startTimestamp, endTimestamp)
	if err != nil {
		return GetUserMonthlyAnalysisResp{}, 60801, err
	}
	return GetUserMonthlyAnalysisResp{
		DanmuCount: danmu,
		GiftCount:  giftCount,
		GiftAmount: giftAmount,
		LiveDays:   liveDays,
	}, 0, nil
}

// GetUserDanmuAnalysis 获取用户弹幕分析数据
func (s *Service) GetUserDanmuAnalysis(ctx context.Context, UID int64) (GetUserDanmuAnalysisResp, int, error) {
	danmu, err := s.liveDanmuRepo.GetMessagesByUID(ctx, nil, UID)
	if err != nil {
		return GetUserDanmuAnalysisResp{}, 60801, err
	}
	if len(danmu) == 0 {
		return GetUserDanmuAnalysisResp{}, 0, nil
	}
	tok, err := tokenizer.Get()
	if err != nil {
		return GetUserDanmuAnalysisResp{}, 50801, err
	}
	// 单词
	wordsData := tok.CutAndFilterAll(danmu)
	if len(wordsData) > userDanmuAnalysisLimit {
		wordsData = wordsData[:userDanmuAnalysisLimit]
	}
	words := make([]WordFrequency, len(wordsData))
	for i, item := range wordsData {
		words[i] = WordFrequency{
			Word:  item.Word,
			Count: item.Count,
		}
	}
	// 双词
	bigramsData := tok.CountNGram(danmu, 2)
	if len(bigramsData) > userDanmuAnalysisLimit {
		bigramsData = bigramsData[:userDanmuAnalysisLimit]
	}
	bigrams := make([]WordFrequency, len(bigramsData))
	for i, item := range bigramsData {
		bigrams[i] = WordFrequency{
			Word:  item.Phrase,
			Count: item.Count,
		}
	}
	// 三词
	trigramsData := tok.CountNGram(danmu, 3)
	if len(trigramsData) > userDanmuAnalysisLimit {
		trigramsData = trigramsData[:userDanmuAnalysisLimit]
	}
	trigrams := make([]WordFrequency, len(trigramsData))
	for i, item := range trigramsData {
		trigrams[i] = WordFrequency{
			Word:  item.Phrase,
			Count: item.Count,
		}
	}
	// 短句
	messagesData := tok.CountMessages(danmu)
	if len(messagesData) > userDanmuAnalysisLimit {
		messagesData = messagesData[:userDanmuAnalysisLimit]
	}
	messages := make([]WordFrequency, len(messagesData))
	for i, item := range messagesData {
		messages[i] = WordFrequency{
			Word:  item.Message,
			Count: item.Count,
		}
	}
	// 返回数据
	return GetUserDanmuAnalysisResp{
		Words:    words,
		Bigrams:  bigrams,
		Trigrams: trigrams,
		Messages: messages,
	}, 0, nil
}

// EnsureUser 获取用户 ID，如果用户不存在则自动注册
func (s *Service) EnsureUser(ctx context.Context, uid int64, uname string) (int64, error) {
	user, err := s.liveUserRepo.GetByUID(ctx, nil, uid)
	if err != nil {
		return 0, fmt.Errorf("获取用户ID信息失败：%w", err)
	}
	if user != nil {
		if user.Uname != uname {
			if err := s.liveUserRepo.UpdateName(ctx, nil, user.ID, uname); err != nil {
				return 0, fmt.Errorf("更新用户名称失败：%w", err)
			}
		}
		return user.ID, nil
	}
	// 用户注册
	danmuCount, err := s.liveDanmuRepo.CountByUID(ctx, nil, uid)
	if err != nil {
		return 0, fmt.Errorf("获取用户弹幕总数失败：%w", err)
	}
	giftTotalAmount, err := s.liveGiftRepo.SumTotalGiftAmountByUID(ctx, nil, uid)
	if err != nil {
		return 0, fmt.Errorf("获取用户消费金额失败：%w", err)
	}
	user, err = s.liveUserRepo.CreateIfNotExist(ctx, nil, &model.LiveUser{
		UID:             uid,
		Uname:           uname,
		TotalDanmuCount: danmuCount,
		TotalGiftAmount: giftTotalAmount,
	})
	if err != nil {
		return 0, fmt.Errorf("用户注册失败：%w", err)
	}
	// 并发下可能出现"插入冲突后回查也没查到"，此时 user 为 nil，直接返回错误避免空指针
	if user == nil {
		return 0, fmt.Errorf("用户注册后未查询到记录：uid=%d", uid)
	}
	return user.ID, nil
}

// GetUserBalance 获取用户余额信息，如果用户不存在则返回空信息
func (s *Service) GetUserBalance(ctx context.Context, uid int64) (*UserBalance, error) {
	user, err := s.liveUserRepo.GetByUID(ctx, nil, uid)
	if err != nil {
		return nil, fmt.Errorf("获取用户余额失败：%w", err)
	}
	if user == nil {
		return nil, nil
	}
	return &UserBalance{
		Points: user.Points,
		Stars:  user.Stars,
	}, nil
}

// AddTotalDanmuCount 增加用户累计发送弹幕数
func (s *Service) AddTotalDanmuCount(ctx context.Context, userID int64) error {
	return s.liveUserRepo.IncrementField(ctx, nil, userID, "total_danmu_count", 1)
}

// AddTotalGiftAmount 增加用户累计赠送礼物金额
func (s *Service) AddTotalGiftAmount(ctx context.Context, userID int64, amount int64) error {
	return s.liveUserRepo.IncrementField(ctx, nil, userID, "total_gift_amount", amount)
}

// AddPointsLog 增加用户积分记录（增加或减少）
func (s *Service) AddPointsLog(ctx context.Context, params AddCreditLogParams) error {
	return s.addCreditLog(ctx, params, enum.CreditTypePoints, live_user.CreditFieldPoints)
}

// AddStarsLog 增加用户星光记录（增加或减少）
func (s *Service) AddStarsLog(ctx context.Context, params AddCreditLogParams) error {
	return s.addCreditLog(ctx, params, enum.CreditTypeStars, live_user.CreditFieldStars)
}

// addCreditLog 增加用户资产记录（增加或减少）
//
// 资产变更交给数据库原子完成，再按其返回的变更前后数值写流水，
// 保证并发场景下流水与用户余额始终对得上
func (s *Service) addCreditLog(ctx context.Context, params AddCreditLogParams, creditType enum.CreditType, field string) error {
	if params.ChangeAmount < 0 {
		return fmt.Errorf("变动数值不能为负数: %d", params.ChangeAmount)
	}
	// 变动类型换算成增量，扣减为负数
	var delta int64
	switch params.ChangeType {
	case enum.ChangeTypeIncrease:
		delta = params.ChangeAmount
	case enum.ChangeTypeReduce:
		delta = -params.ChangeAmount
	default:
		return fmt.Errorf("未知的变动类型: %v", params.ChangeType)
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 原子变更用户资产，余额不足会被数据库条件拦下
		beforeValue, afterValue, err := s.liveUserRepo.AddCredit(ctx, tx, params.UserID, field, delta)
		if err != nil {
			return fmt.Errorf("更新用户资产失败：%w", err)
		}
		// 创建变动记录
		if _, err := s.liveUserCreditLogRepo.Create(ctx, tx, &model.LiveUserCreditLog{
			UserID:       params.UserID,
			CreditType:   creditType,
			ChangeType:   params.ChangeType,
			ChangeAmount: params.ChangeAmount,
			BeforeValue:  beforeValue,
			AfterValue:   afterValue,
			BizType:      params.BizType,
			Remark:       params.Remark,
			OperatorType: params.OperatorType,
			OperatorID:   params.OperatorID,
		}); err != nil {
			return fmt.Errorf("创建记录失败：%w", err)
		}
		return nil
	})
}
