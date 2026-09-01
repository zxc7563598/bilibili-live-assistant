package liveuser

import (
	"context"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/jwt"
)

// updateToken 用于更新用户token
//
// 写库顺序说明（有意为之）：先写 DB 再写 Redis。
// DB 是 refresh_token 的唯一权威来源，先写 DB 保证刷新校验永远基于最新值；
// Redis 仅用于单点登录校验，若写入失败最坏情况是新 access_token 无法通过校验、用户重新登录，不会产生安全泄漏。
func (s *Service) updateToken(ctx context.Context, userID int64) (TokenResp, int, error) {
	accessToken, err := jwt.GenerateAccessToken(userID, "user", 0, "")
	if err != nil {
		return TokenResp{}, 60802, err
	}
	newRefreshToken, err := jwt.GenerateRefreshToken(userID, "user", 0, "")
	if err != nil {
		return TokenResp{}, 60803, err
	}
	if err := s.liveUserRepo.UpdateTokenByID(ctx, nil, userID, &newRefreshToken); err != nil {
		return TokenResp{}, 60804, err
	}
	if s.rdb != nil {
		if err := s.rdb.Set(ctx, jwt.UserTokenKey(userID), accessToken, jwt.AccessTTL()).Err(); err != nil {
			return TokenResp{}, 60805, err
		}
		if err := s.rdb.Set(ctx, jwt.UserRefreshKey(userID), newRefreshToken, jwt.RefreshTTL()).Err(); err != nil {
			return TokenResp{}, 60806, err
		}
	}
	return TokenResp{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, 0, nil
}

// configValue 读取应用配置值，配置项缺失时返回空字符串
func (s *Service) configValue(key string) string {
	val, _ := s.appConfigCache.Get(key)
	return val
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
