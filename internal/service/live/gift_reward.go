package live

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zxc7563598/bilibili-live-assistant/internal/enum"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"github.com/zxc7563598/bilibili-live-assistant/internal/service/liveuser"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/ptr"
)

// grantGiftReward 按房间奖励配置给用户发放礼物奖励
//
// 直播链路与后台补录共用这一份金额口径，避免两处各算一遍导致对不上。
// 「机器人给自己刷礼物不发奖励」这类前置判断由调用方负责。
func grantGiftReward(ctx context.Context, configCache *robotconfig.Cache, liveUserSvc *liveuser.Service, userID int64, giftName string, giftType enum.GiftType, badgeType enum.BadgeType, price, num int64) error {
	var roomCfg robotconfig.RoomConfig
	if err := configCache.UnmarshalGroup("room", &roomCfg); err != nil {
		return fmt.Errorf("加载房间配置失败：%w", err)
	}
	switch ptr.ParseEnumInt[enum.RewardPolicy](roomCfg.ConsumeRewardEnabled) {
	case enum.RewardPolicyBatteryReward:
		magnification, _ := strconv.ParseInt(roomCfg.ConsumeBatteryRate, 10, 64)
		return grantBatteryReward(ctx, liveUserSvc, userID, giftName,
			ptr.ParseEnumInt[enum.RewardType](roomCfg.RewardType), price, num, magnification)
	case enum.RewardPolicyVipReward:
		if giftType != enum.GiftTypeGuard {
			return nil
		}
		var reward int64
		switch badgeType {
		case enum.BadgeTypeL1:
			reward, _ = strconv.ParseInt(roomCfg.CaptainRewardAmount, 10, 64)
		case enum.BadgeTypeL2:
			reward, _ = strconv.ParseInt(roomCfg.CommanderRewardAmount, 10, 64)
		case enum.BadgeTypeL3:
			reward, _ = strconv.ParseInt(roomCfg.GovernorRewardAmount, 10, 64)
		}
		return grantVipReward(ctx, liveUserSvc, userID,
			ptr.ParseEnumInt[enum.RewardType](roomCfg.RewardType), badgeType, reward)
	}
	return nil
}

// grantBatteryReward 按消费电池发放积分
func grantBatteryReward(ctx context.Context, liveUserSvc *liveuser.Service, userID int64, giftName string, rewardType enum.RewardType, price, num, magnification int64) error {
	battery := price / 10
	params := liveuser.AdjustCreditParams{
		UserID:       userID,
		ChangeType:   enum.ChangeTypeIncrease,
		ChangeAmount: (battery * num) * magnification,
		BizType:      "gift",
		Remark:       fmt.Sprintf("用户赠送 %d 个价值 %d 电池的 %s (积分倍率 %d)", num, battery, giftName, magnification),
		OperatorType: enum.OperatorTypeSystem,
		OperatorID:   0,
	}
	return adjustCredit(ctx, liveUserSvc, rewardType, params)
}

// grantVipReward 按航海类型发放积分
func grantVipReward(ctx context.Context, liveUserSvc *liveuser.Service, userID int64, rewardType enum.RewardType, level enum.BadgeType, reward int64) error {
	params := liveuser.AdjustCreditParams{
		UserID:       userID,
		ChangeType:   enum.ChangeTypeIncrease,
		ChangeAmount: reward,
		BizType:      "gift",
		Remark:       fmt.Sprintf("用户开通 %s 奖励 %d", level.Text("zh"), reward),
		OperatorType: enum.OperatorTypeSystem,
		OperatorID:   0,
	}
	return adjustCredit(ctx, liveUserSvc, rewardType, params)
}

// adjustCredit 按奖励类型写入积分或星光
func adjustCredit(ctx context.Context, liveUserSvc *liveuser.Service, rewardType enum.RewardType, params liveuser.AdjustCreditParams) error {
	switch rewardType {
	case enum.RewardTypePoints:
		return liveUserSvc.AdjustPoints(ctx, nil, params)
	case enum.RewardTypeStars:
		return liveUserSvc.AdjustStars(ctx, nil, params)
	}
	return nil
}
