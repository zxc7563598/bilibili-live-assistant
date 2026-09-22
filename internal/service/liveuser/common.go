package liveuser

import (
	"context"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/jwt"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/timeutil"
)

// guardValidDays 一次大航海购买的有效天数
const guardValidDays = 30

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
	for _, v := range liveUser {
		item := ListPageItem{
			ID:              v.ID,
			UID:             v.UID,
			Uname:           v.Uname,
			Points:          v.Points,
			Stars:           v.Stars,
			TotalDanmuCount: v.TotalDanmuCount,
			TotalGiftAmount: v.TotalGiftAmount,
		}
		respList = append(respList, item)
	}
	return respList
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
