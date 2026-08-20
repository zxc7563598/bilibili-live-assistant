package liveuser

import (
	"context"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/internal/model"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_danmu"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_gift"
	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/live_user"
)

type Service struct {
	liveUserRepo  live_user.Repository
	liveDanmuRepo live_danmu.Repository
	liveGiftRepo  live_gift.Repository
}

func New(liveUserRepo live_user.Repository, liveDanmuRepo live_danmu.Repository, liveGiftRepo live_gift.Repository) *Service {
	return &Service{
		liveUserRepo:  liveUserRepo,
		liveDanmuRepo: liveDanmuRepo,
		liveGiftRepo:  liveGiftRepo,
	}
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
	giftTotalAmount, err := s.liveGiftRepo.GetUserTotalGiftAmount(ctx, nil, uid)
	if err != nil {
		return 0, fmt.Errorf("获取用户消费金额失败：%w", err)
	}
	user, err = s.liveUserRepo.CreateIfNotExist(ctx, nil, &model.LiveUser{
		Uid:             uid,
		Uname:           uname,
		TotalDanmuCount: danmuCount,
		TotalGiftAmount: giftTotalAmount,
	})
	if err != nil {
		return 0, fmt.Errorf("用户注册失败：%w", err)
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
