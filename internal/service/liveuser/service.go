package liveuser

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zxc7563598/bilibili-live-assistant/internal/appconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_session"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user_credit_log"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/crypto"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/jwt"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/tokenizer"
	"gorm.io/gorm"
)

// 站点基础配置键
const (
	keyRegister = "register"
)

type Service struct {
	client                *bilibili.Client
	db                    *gorm.DB
	rdb                   *redis.Client
	appConfigCache        *appconfig.Cache
	liveUserRepo          live_user.Repository
	liveUserCreditLogRepo live_user_credit_log.Repository
	liveDanmuRepo         live_danmu.Repository
	liveGiftRepo          live_gift.Repository
	liveSessionRepo       live_session.Repository
}

const userDanmuAnalysisLimit = 20

func New(db *gorm.DB, rdb *redis.Client, appConfigCache *appconfig.Cache, liveUserRepo live_user.Repository, liveUserCreditLogRepo live_user_credit_log.Repository, liveDanmuRepo live_danmu.Repository, liveGiftRepo live_gift.Repository, liveSessionRepo live_session.Repository) *Service {
	return &Service{
		client:                bilibili.NewClient(),
		db:                    db,
		rdb:                   rdb,
		appConfigCache:        appConfigCache,
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
	offset, limit, sortField, sortOrder := req.OffsetLimit()
	listDanmu, total, err := s.liveUserRepo.ListPage(ctx, nil, model.LiveUserListPageQuery{
		UID:       req.UID,
		Uname:     req.Uname,
		Offset:    offset,
		Limit:     limit,
		SortField: sortField,
		SortOrder: sortOrder,
	})
	if err != nil {
		return ListPageResp{}, CodeQueryFailed, err
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
		return GetUserMonthlyAnalysisResp{}, CodeParamInvalid, fmt.Errorf("非法的年月参数: year=%d, month=%d", year, month)
	}
	// 确定查询时间范围
	start := time.Date(int(year), time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	startTimestamp := start.Unix()
	endTimestamp := end.Unix()
	// 获取数据
	danmu, err := s.liveDanmuRepo.CountDailyByUID(ctx, nil, UID, startTimestamp, endTimestamp)
	if err != nil {
		return GetUserMonthlyAnalysisResp{}, CodeQueryFailed, err
	}
	gift, err := s.liveGiftRepo.CountDailyByUID(ctx, nil, UID, startTimestamp, endTimestamp)
	if err != nil {
		return GetUserMonthlyAnalysisResp{}, CodeQueryFailed, err
	}
	giftCount := make(map[int64]int64, len(gift))
	giftAmount := make(map[int64]int64, len(gift))
	for day, item := range gift {
		giftCount[int64(day)] = item.Num
		giftAmount[int64(day)] = item.Amount
	}
	// 本月开播记录（不区分房间ID/主播，按开播时间 start_at 落在当月统计）
	liveDaySet, err := s.liveSessionRepo.DistinctLiveDays(ctx, nil, startTimestamp, endTimestamp)
	if err != nil {
		return GetUserMonthlyAnalysisResp{}, CodeQueryFailed, err
	}
	// 仓库层返回的是日号集合，响应结构需要 map[日号]bool，这里转换一次
	liveDays := make(map[int64]bool, len(liveDaySet))
	for day := range liveDaySet {
		liveDays[day] = true
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
	danmu, err := s.liveDanmuRepo.ListMessagesByUID(ctx, nil, UID)
	if err != nil {
		return GetUserDanmuAnalysisResp{}, CodeQueryFailed, err
	}
	if len(danmu) == 0 {
		return GetUserDanmuAnalysisResp{}, 0, nil
	}
	tok, err := tokenizer.Get()
	if err != nil {
		return GetUserDanmuAnalysisResp{}, CodeTokenizerInitFailed, err
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
			if err := s.liveUserRepo.UpdateNameByID(ctx, nil, user.ID, uname); err != nil {
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
	return s.liveUserRepo.AdjustField(ctx, nil, userID, "total_danmu_count", 1)
}

// AddTotalGiftAmount 增加用户累计赠送礼物金额
func (s *Service) AddTotalGiftAmount(ctx context.Context, userID int64, amount int64) error {
	return s.liveUserRepo.AdjustField(ctx, nil, userID, "total_gift_amount", amount)
}

// ExtendGuardExpire 延长用户某档大航海的到期时间
func (s *Service) ExtendGuardExpire(ctx context.Context, uid int64, uname string, level enum.BadgeType, recvAt time.Time) (int, error) {
	field, err := guardExpireField(level)
	if err != nil {
		return CodeParamInvalid, err
	}
	userID, err := s.EnsureUser(ctx, uid, uname)
	if err != nil {
		return CodeQueryFailed, err
	}
	user, err := s.liveUserRepo.GetByID(ctx, nil, userID)
	if err != nil {
		return CodeQueryFailed, fmt.Errorf("获取用户信息失败：%w", err)
	}
	if user == nil {
		return CodeUserNotFound, fmt.Errorf("用户记录缺失: id=%d", userID)
	}
	base := time.Date(recvAt.Year(), recvAt.Month(), recvAt.Day(), 0, 0, 0, 0, recvAt.Location())
	// 仍在有效期内则接着原到期时间算
	if current := guardExpireValue(level, user); current != nil && *current > base.Unix() {
		base = time.Unix(*current, 0).In(recvAt.Location())
	}
	if err := s.liveUserRepo.UpdateField(ctx, nil, userID, field, base.AddDate(0, 0, guardValidDays).Unix()); err != nil {
		return CodeQueryFailed, fmt.Errorf("记录大航海到期时间失败：%w", err)
	}
	return 0, nil
}

// AdjustPoints 增减用户积分并写资产流水。tx 为 nil 时自开事务
func (s *Service) AdjustPoints(ctx context.Context, tx *gorm.DB, params AdjustCreditParams) error {
	return s.addCreditLog(ctx, tx, params, enum.CreditTypePoints, live_user.CreditFieldPoints)
}

// AdjustStars 增减用户星光并写资产流水。tx 为 nil 时自开事务
func (s *Service) AdjustStars(ctx context.Context, tx *gorm.DB, params AdjustCreditParams) error {
	return s.addCreditLog(ctx, tx, params, enum.CreditTypeStars, live_user.CreditFieldStars)
}

// AdjustCredit 按资产类型增减用户资产并写流水，供持有 creditType 而非具体资产名的调用方
// （如订单模块）直接透传。tx 为 nil 时自开事务，非 nil 时复用调用方事务。
func (s *Service) AdjustCredit(ctx context.Context, tx *gorm.DB, creditType enum.CreditType, params AdjustCreditParams) error {
	field := live_user.CreditFieldStars
	if creditType == enum.CreditTypePoints {
		field = live_user.CreditFieldPoints
	}
	return s.addCreditLog(ctx, tx, params, creditType, field)
}

// ExistsAccount 获取用户是否存在
func (s *Service) ExistsAccount(ctx context.Context, account int64) (bool, int, error) {
	// 获取用户是否存在
	exists, err := s.liveUserRepo.ExistsByUID(ctx, nil, account)
	if err != nil {
		return false, CodeQueryFailed, err
	}
	return exists, 0, nil
}

// Login 执行登录
func (s *Service) Login(ctx context.Context, account int64, password string) (TokenResp, int, error) {
	// 获取用户信息
	user, err := s.liveUserRepo.GetByUID(ctx, nil, account)
	if err != nil {
		return TokenResp{}, CodeQueryFailed, err
	}
	// 用户不存在且不允许注册，直接结束
	register := ptr.ParseEnumInt[enum.YesNo](s.appConfigCache.GetValue(keyRegister))
	if user == nil && register == enum.No {
		return TokenResp{}, CodeUserNotFound, nil
	}
	// 已存在用户：先校验启用状态与密码，避免对无效请求发起 B站 请求
	if user != nil {
		if user.Enable != enum.EnableEnable {
			return TokenResp{}, CodeAccountDisabled, nil
		}
		if user.Password != "" && !crypto.CheckPassword(user.Password, password) {
			return TokenResp{}, CodeLoginFailed, nil
		}
	}
	// 从B站获取主播信息（注册 / 同步名称头像 / 无密码设置密码都需要）
	// 取不到是上游/网络问题，与「用户不存在」是两回事，不能都报 CodeUserNotFound 未知用户
	master, err := s.client.User.GetMasterInfo(ctx, account)
	if err != nil {
		return TokenResp{}, CodeBilibiliUserFailed, nil
	}
	if master.Name == "" && master.Face == "" {
		return TokenResp{}, CodeBilibiliUserFailed, nil
	}
	// 用户不存在：自动注册后回查完整记录
	if user == nil {
		if _, err := s.EnsureUser(ctx, master.UID, master.Name); err != nil {
			return TokenResp{}, CodeQueryFailed, err
		}
		user, err = s.liveUserRepo.GetByUID(ctx, nil, master.UID)
		if err != nil || user == nil {
			return TokenResp{}, CodeQueryFailed, err
		}
	}
	// 无密码用户：将本次输入的密码作为其密码
	if user.Password == "" {
		hash, err := crypto.HashPassword(password)
		if err != nil {
			return TokenResp{}, CodeUserNotFound, err
		}
		if err := s.liveUserRepo.UpdatePasswordByID(ctx, nil, user.ID, hash); err != nil {
			return TokenResp{}, CodeQueryFailed, err
		}
	}
	// 同步名称与头像（仅在变化时写库）
	if user.Uname != master.Name {
		if err := s.liveUserRepo.UpdateNameByID(ctx, nil, user.ID, master.Name); err != nil {
			return TokenResp{}, CodeQueryFailed, err
		}
	}
	if user.Face != master.Face {
		if err := s.liveUserRepo.UpdateFaceByID(ctx, nil, user.ID, master.Face); err != nil {
			return TokenResp{}, CodeQueryFailed, err
		}
	}
	// 更新token
	return s.updateToken(ctx, user.ID)
}

// RefreshLogin 用于刷新用户登录状态
func (s *Service) RefreshLogin(ctx context.Context, refreshToken string) (TokenResp, int, error) {
	claims, err := jwt.ParseToken(refreshToken)
	if err != nil {
		return TokenResp{}, CodeTokenInvalid, err
	}
	if claims.Type != "refresh" {
		return TokenResp{}, CodeTokenTypeInvalid, nil
	}
	// 获取用户信息
	user, err := s.liveUserRepo.GetByID(ctx, nil, claims.ID)
	if err != nil {
		return TokenResp{}, CodeQueryFailed, err
	}
	// 验证信息
	if user == nil {
		return TokenResp{}, CodeUserNotFound, nil
	}
	if user.Token == nil || *user.Token != refreshToken {
		return TokenResp{}, CodeTokenExpired, nil
	}
	// 更新token
	return s.updateToken(ctx, claims.ID)
}

// Logout 用于退出用户登录
func (s *Service) Logout(ctx context.Context, userID int64) (int, error) {
	// 清空用户token
	if s.rdb != nil {
		err := s.rdb.Del(ctx,
			jwt.UserTokenKey(userID),
			jwt.UserRefreshKey(userID),
		).Err()
		if err != nil {
			return CodeTokenClearFailed, err
		}
	}
	if err := s.liveUserRepo.UpdateTokenByID(ctx, nil, userID, nil); err != nil {
		return CodeTokenPersistFailed, err
	}
	// 返回数据
	return 0, nil
}

// ChangePassword 用于根据用户旧密码修改密码
func (s *Service) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) (int, error) {
	// 根据主键ID获取用户信息
	user, err := s.liveUserRepo.GetByID(ctx, nil, userID)
	if err != nil {
		return CodeQueryFailed, err
	}
	if user == nil {
		return CodeUserNotFound, nil
	}
	// 验证旧密码是否正确
	if !crypto.CheckPassword(user.Password, oldPassword) {
		return CodeLoginFailed, nil
	}
	// 新密码加密并更新
	password, err := crypto.HashPassword(newPassword)
	if err != nil {
		return CodeQueryFailed, err
	}
	if err := s.liveUserRepo.UpdatePasswordByID(ctx, nil, user.ID, password); err != nil {
		return CodeQueryFailed, err
	}
	// 返回结果
	return 0, nil
}

// ResetPassword 用于直接重置用户密码
func (s *Service) ResetPassword(ctx context.Context, userID int64, newPassword string) (int, error) {
	// 根据主键ID获取用户信息
	user, err := s.liveUserRepo.GetByID(ctx, nil, userID)
	if err != nil {
		return CodeQueryFailed, err
	}
	if user == nil {
		return CodeUserNotFound, nil
	}
	// 新密码加密并更新
	password, err := crypto.HashPassword(newPassword)
	if err != nil {
		return CodeQueryFailed, err
	}
	if err := s.liveUserRepo.UpdatePasswordByID(ctx, nil, user.ID, password); err != nil {
		return CodeQueryFailed, err
	}
	// 返回结果
	return 0, nil
}

// GetUserInfo 获取用户基本信息
func (s *Service) GetUserInfo(ctx context.Context, userID int64) (UserInfoResp, int, error) {
	// 根据主键ID获取用户信息
	user, err := s.liveUserRepo.GetByID(ctx, nil, userID)
	if err != nil {
		return UserInfoResp{}, CodeQueryFailed, err
	}
	if user == nil {
		return UserInfoResp{}, CodeUserNotFound, nil
	}
	return UserInfoResp{
		UID:    user.UID,
		Avatar: user.Face,
		Name:   user.Uname,
		Points: user.Points,
		Stars:  user.Stars,
	}, 0, nil
}

// ListUserAssets 分页获取用户账户变更记录
func (s *Service) ListUserAssets(ctx context.Context, userID int64, req UserAssetsPageReq) (UserAssetsPageResp, int, error) {
	// 根据主键ID获取用户信息
	offset, limit, sortField, sortOrder := req.OffsetLimit()
	list, total, err := s.liveUserCreditLogRepo.ListPage(ctx, nil, model.LiveUserCreditLogListPageQuery{
		UID:        req.UID,
		UserID:     &userID,
		Uname:      req.Uname,
		CreditType: req.CreditType,
		ChangeType: req.ChangeType,
		Offset:     offset,
		Limit:      limit,
		SortField:  sortField,
		SortOrder:  sortOrder,
	})
	if err != nil {
		return UserAssetsPageResp{}, CodeQueryFailed, err
	}
	// 返回数据
	return UserAssetsPageResp{
		Total:    total,
		PageData: toUserAssetsPageItems(list),
	}, 0, nil
}

// SaveBalance 管理员手动变更用户余额
//
// 与其他余额变更入口一致，走 addCreditLog 原子更新资产并写流水，
// 操作方固定记为管理员，便于后续追溯是谁调整的
func (s *Service) SaveBalance(ctx context.Context, adminID, userID int64, creditType, changeType int, changeAmount int64, remark *string) (int, error) {
	ct := enum.CreditType(creditType)
	if !ct.IsValid() {
		return CodeParamInvalid, fmt.Errorf("非法的资产类型: %d", creditType)
	}
	t := enum.ChangeType(changeType)
	if !t.IsValid() {
		return CodeParamInvalid, fmt.Errorf("非法的变动类型: %d", changeType)
	}
	if changeAmount <= 0 {
		return CodeParamInvalid, fmt.Errorf("变动数值必须大于 0: %d", changeAmount)
	}
	desc := strings.TrimSpace(ptr.Deref(remark))
	if desc == "" {
		desc = fmt.Sprintf("管理员后台手动%s%s %d", t.Text("zh"), ct.Text("zh"), changeAmount)
	}
	// 组装流水参数
	params := AdjustCreditParams{
		UserID:       userID,
		ChangeType:   t,
		ChangeAmount: changeAmount,
		BizType:      "admin",
		Remark:       desc,
		OperatorType: enum.OperatorTypeAdmin,
		OperatorID:   adminID,
	}
	// 执行变更（自开事务）
	err := s.AdjustCredit(ctx, nil, ct, params)
	if err != nil {
		switch {
		case errors.Is(err, live_user.ErrUserNotFound):
			return CodeUserNotFound, err
		case errors.Is(err, live_user.ErrInsufficientBalance):
			return CodeInsufficientBalance, err
		default:
			return CodeQueryFailed, err
		}
	}
	// 返回结果
	return 0, nil
}

// addCreditLog 增减用户资产并写流水
//
// 资产变更交给数据库原子完成，再按其返回的变更前后数值写流水，
// 保证并发场景下流水与用户余额始终对得上。
// tx 为 nil 时自开事务；非 nil 时复用调用方事务（订单模块在自身事务里扣款即走这条）。
func (s *Service) addCreditLog(ctx context.Context, tx *gorm.DB, params AdjustCreditParams, creditType enum.CreditType, field string) error {
	if params.ChangeAmount < 0 {
		return fmt.Errorf("变动数值不能为负数: %d", params.ChangeAmount)
	}
	if params.ChangeAmount == 0 {
		return nil
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
	// 资产变更 + 写流水的整体逻辑，两种事务来源共用
	apply := func(tx *gorm.DB) error {
		// 原子变更用户资产，余额不足会被数据库条件拦下
		beforeValue, afterValue, err := s.liveUserRepo.AdjustCredit(ctx, tx, params.UserID, field, delta)
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
	}
	// 调用方已开启事务时复用，避免嵌套
	if tx != nil {
		return apply(tx)
	}
	return s.db.Transaction(apply)
}
