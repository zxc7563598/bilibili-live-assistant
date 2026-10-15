package liveuser

import (
	"context"
	"fmt"
	"time"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/jwt"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// guardValidDays 一份大航海的有效天数
const guardValidDays = 30

// GuardExpireWindow 依据上一段的结束时间与本次开通时间，算出本次开通的有效期窗口 [startAt, endAt)
//
// 仍在有效期内就接着原到期时间往后堆叠，否则从当天零点重新起算；
// 一次开通多份时按每份 30 天连着排。
//
// 这份算术算的是「身份有效到什么时候」，与大航海名单的月份归属无关（见 GuardCoverWindow）：
// 7 月 29 日开通一份，有效期到 8 月 28 日，但他只在 7 月的名单里。
func GuardExpireWindow(prevEnd int64, recvAt time.Time, num int64) (startAt, endAt int64) {
	base := timeutil.StartOfDay(recvAt)
	if prevEnd > base.Unix() {
		base = time.Unix(prevEnd, 0).In(recvAt.Location())
	}
	if num < 1 {
		num = 1
	}
	return base.Unix(), base.AddDate(0, 0, guardValidDays*int(num)).Unix()
}

// GuardCoverWindow 依据上一段覆盖到的月份与本次开通时间，算出本次开通覆盖的名单月份区间 [from, to)
//
// 两端都是本地月首：一次开通买 num 份就是连着排 num 个月；仍在覆盖期内就接着上一段的
// 终点往后排，同一个月开通多次才能把后面的月份顶出去。
//
// 名单只认月份，不认天数：7 月 29 日开通一份，覆盖区间是 [7/1, 8/1)，他进 7 月名单、
// 不进 8 月名单，尽管有效期到 8 月 28 日。
func GuardCoverWindow(prevCoverTo int64, openAt time.Time, num int64) (from, to int64) {
	if num < 1 {
		num = 1
	}
	start := timeutil.StartOfMonth(openAt)
	if prevCoverTo > start.Unix() {
		start = time.Unix(prevCoverTo, 0).In(openAt.Location())
	}
	return start.Unix(), timeutil.AddMonths(start, int(num)).Unix()
}

// guardExpireOwner 注册并取回用户记录，大航海到期时间的读写都要先拿到它
func (s *Service) guardExpireOwner(ctx context.Context, uid int64, uname string) (int64, *model.LiveUser, int, error) {
	userID, err := s.EnsureUser(ctx, uid, uname)
	if err != nil {
		return 0, nil, CodeQueryFailed, err
	}
	user, err := s.liveUserRepo.GetByID(ctx, nil, userID)
	if err != nil {
		return 0, nil, CodeQueryFailed, fmt.Errorf("获取用户信息失败：%w", err)
	}
	if user == nil {
		return 0, nil, CodeUserNotFound, fmt.Errorf("用户记录缺失: id=%d", userID)
	}
	return userID, user, 0, nil
}

// guardExpireField 大航海等级 → 到期时间列名
func guardExpireField(level enum.BadgeType) (string, error) {
	switch level {
	case enum.BadgeTypeL1:
		return "captain_expire_at", nil
	case enum.BadgeTypeL2:
		return "admiral_expire_at", nil
	case enum.BadgeTypeL3:
		return "governor_expire_at", nil
	default:
		return "", fmt.Errorf("非大航海等级: %d", level)
	}
}

// guardExpireValue 用户当前该档的到期时间，未记录返回 nil
func guardExpireValue(level enum.BadgeType, user *model.LiveUser) *int64 {
	switch level {
	case enum.BadgeTypeL1:
		return user.CaptainExpireAt
	case enum.BadgeTypeL2:
		return user.AdmiralExpireAt
	case enum.BadgeTypeL3:
		return user.GovernorExpireAt
	default:
		return nil
	}
}

// updateToken 用于更新用户token
//
// 写库顺序说明（有意为之）：先写 DB 再写 Redis。
// DB 是 refresh_token 的唯一权威来源，先写 DB 保证刷新校验永远基于最新值；
// Redis 仅用于单点登录校验，若写入失败最坏情况是新 access_token 无法通过校验、用户重新登录，不会产生安全泄漏。
func (s *Service) updateToken(ctx context.Context, userID int64) (TokenResp, int, error) {
	accessToken, err := jwt.GenerateAccessToken(userID, "user", 0, "")
	if err != nil {
		return TokenResp{}, CodeAccessTokenGenFailed, err
	}
	newRefreshToken, err := jwt.GenerateRefreshToken(userID, "user", 0, "")
	if err != nil {
		return TokenResp{}, CodeRefreshTokenGenFailed, err
	}
	if err := s.liveUserRepo.UpdateTokenByID(ctx, nil, userID, &newRefreshToken); err != nil {
		return TokenResp{}, CodeTokenPersistFailed, err
	}
	if s.rdb != nil {
		if err := s.rdb.Set(ctx, jwt.UserTokenKey(userID), accessToken, jwt.AccessTTL()).Err(); err != nil {
			return TokenResp{}, CodeAccessTokenCacheFailed, err
		}
		if err := s.rdb.Set(ctx, jwt.UserRefreshKey(userID), newRefreshToken, jwt.RefreshTTL()).Err(); err != nil {
			return TokenResp{}, CodeRefreshTokenCacheFailed, err
		}
	}
	return TokenResp{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, 0, nil
}

func toListPageItems(liveUser []model.LiveUser) []ListPageItem {
	respList := make([]ListPageItem, 0, len(liveUser))
	now := time.Now().Unix()
	for _, v := range liveUser {
		item := ListPageItem{
			ID:              v.ID,
			UID:             v.UID,
			Uname:           v.Uname,
			Points:          v.Points,
			Stars:           v.Stars,
			TotalDanmuCount: v.TotalDanmuCount,
			TotalGiftAmount: v.TotalGiftAmount,
			VipType:         guardVipType(&v, now),
		}
		respList = append(respList, item)
	}
	return respList
}

// guardVipType 用户当前生效的大航海身份
func guardVipType(user *model.LiveUser, now int64) enum.BadgeType {
	switch {
	case user.GovernorExpireAt != nil && *user.GovernorExpireAt > now:
		return enum.BadgeTypeL3
	case user.AdmiralExpireAt != nil && *user.AdmiralExpireAt > now:
		return enum.BadgeTypeL2
	case user.CaptainExpireAt != nil && *user.CaptainExpireAt > now:
		return enum.BadgeTypeL1
	default:
		return enum.BadgeTypeL0
	}
}

func toUserAssetsPageItems(list []model.LiveUserCreditLogListItem) []UserAssetsPageItem {
	respList := make([]UserAssetsPageItem, 0, len(list))
	for _, v := range list {
		item := UserAssetsPageItem{
			ID:           v.ID,
			UserID:       v.UserID,
			UID:          v.UID,
			Uname:        v.Uname,
			Face:         v.Face,
			CreditType:   v.CreditType,
			ChangeAmount: v.ChangeAmount,
			ChangeType:   v.ChangeType,
			BeforeValue:  v.BeforeValue,
			AfterValue:   v.AfterValue,
			BizType:      v.BizType,
			Remark:       v.Remark,
			OperatorType: v.OperatorType,
			OperatorID:   v.OperatorID,
			CreatedAt:    timeutil.Format(v.CreatedAt),
		}
		respList = append(respList, item)
	}
	return respList
}
