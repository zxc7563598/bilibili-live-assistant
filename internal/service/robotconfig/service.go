package robotconfig

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zxc7563598/bilibili-live-assistant/internal/repository/robot_config"
	"github.com/zxc7563598/bilibili-live-assistant/internal/robotconfig"
	"gorm.io/gorm"
)

type Service struct {
	robotconfigRepo robot_config.Repository
	configCache     *robotconfig.Cache
	db              *gorm.DB
}

func New(robotconfigRepo robot_config.Repository, configCache *robotconfig.Cache, db *gorm.DB) *Service {
	return &Service{
		robotconfigRepo: robotconfigRepo,
		configCache:     configCache,
		db:              db,
	}
}

// applyConfig 将 config_key -> config_value 映射写入数据库并刷新缓存
func (s *Service) applyConfig(ctx context.Context, groupName string, data map[string]string) (int, error) {
	// 查询该分组下所有配置记录（含 ID）
	records, err := s.robotconfigRepo.FindByField(ctx, nil, "group_name", groupName)
	if err != nil {
		return 60502, fmt.Errorf("查询 %s 配置失败: %w", groupName, err)
	}
	// 构建 config_key -> ID 索引
	keyToID := make(map[string]int64, len(records))
	for _, r := range records {
		keyToID[r.ConfigKey] = r.ID
	}
	// 逐条更新（事务保护：全部成功或全部回滚）
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for configKey, configValue := range data {
			id, ok := keyToID[configKey]
			if !ok {
				continue
			}
			if err := s.robotconfigRepo.UpdateByID(ctx, tx, id, configValue); err != nil {
				return fmt.Errorf("更新 %s.%s 失败: %w", groupName, configKey, err)
			}
		}
		return nil
	})
	if err != nil {
		return 60502, err
	}
	// 刷新缓存
	if err := s.configCache.Reload(ctx); err != nil {
		return 60503, fmt.Errorf("刷新配置缓存失败: %w", err)
	}
	return 0, nil
}

// ======================== room ========================

// GetRoomConfig 用于获取房间模块配置
func (s *Service) GetRoomConfig(ctx context.Context) (RoomConfigResp, int, error) {
	group, ok := s.configCache.GetGroup("room")
	if !ok {
		return RoomConfigResp{}, 0, nil
	}
	return toRoomConfigResp(group), 0, nil
}

// ApplyRoomConfig 用于存储房间模块配置
func (s *Service) ApplyRoomConfig(ctx context.Context, data RoomConfigReq) (int, error) {
	return s.applyConfig(ctx, "room", roomConfigReqToMap(data))
}

// GetRoomID 从缓存读取默认直播间房间号（内部使用，不暴露给前端配置页面）
func (s *Service) GetRoomID() int64 {
	val, ok := s.configCache.Get("room", "room_id")
	if !ok || val == "0" {
		return 0
	}
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// SetRoomID 将直播间房间号持久化到数据库并刷新缓存（内部使用，由 live service 调用）
func (s *Service) SetRoomID(ctx context.Context, roomID int64) (int, error) {
	return s.applyConfig(ctx, "room", map[string]string{
		"room_id": strconv.FormatInt(roomID, 10),
	})
}

// ======================== sign ========================

// GetSignConfig 用于获取签到模块配置
func (s *Service) GetSignConfig(ctx context.Context) (SignConfigResp, int, error) {
	group, ok := s.configCache.GetGroup("sign")
	if !ok {
		return SignConfigResp{}, 0, nil
	}
	resp, err := toSignConfigResp(group)
	if err != nil {
		return SignConfigResp{}, 60501, err
	}
	return resp, 0, nil
}

// ApplySignConfig 用于存储签到模块配置
func (s *Service) ApplySignConfig(ctx context.Context, data SignConfigReq) (int, error) {
	return s.applyConfig(ctx, "sign", signConfigReqToMap(data))
}

// ======================== ad ========================

// GetAdConfig 用于获取定时广告模块配置
func (s *Service) GetAdConfig(ctx context.Context) (AdConfigResp, int, error) {
	group, ok := s.configCache.GetGroup("ad")
	if !ok {
		return AdConfigResp{}, 0, nil
	}
	resp, err := toAdConfigResp(group)
	if err != nil {
		return AdConfigResp{}, 60501, err
	}
	return resp, 0, nil
}

// ApplyAdConfig 用于存储定时广告模块配置
func (s *Service) ApplyAdConfig(ctx context.Context, data AdConfigReq) (int, error) {
	return s.applyConfig(ctx, "ad", adConfigReqToMap(data))
}

// ======================== gift ========================

// GetGiftConfig 用于获取礼物答谢模块配置
func (s *Service) GetGiftConfig(ctx context.Context) (GiftConfigResp, int, error) {
	group, ok := s.configCache.GetGroup("gift")
	if !ok {
		return GiftConfigResp{}, 0, nil
	}
	resp, err := toGiftConfigResp(group)
	if err != nil {
		return GiftConfigResp{}, 60501, err
	}
	return resp, 0, nil
}

// ApplyGiftConfig 用于存储礼物答谢模块配置
func (s *Service) ApplyGiftConfig(ctx context.Context, data GiftConfigReq) (int, error) {
	return s.applyConfig(ctx, "gift", giftConfigReqToMap(data))
}

// ======================== pk ========================

// GetPkConfig 用于获取PK播报模块配置
func (s *Service) GetPkConfig(ctx context.Context) (PkConfigResp, int, error) {
	group, ok := s.configCache.GetGroup("pk")
	if !ok {
		return PkConfigResp{}, 0, nil
	}
	resp, err := toPkConfigResp(group)
	if err != nil {
		return PkConfigResp{}, 60501, err
	}
	return resp, 0, nil
}

// ApplyPkConfig 用于存储PK播报模块配置
func (s *Service) ApplyPkConfig(ctx context.Context, data PkConfigReq) (int, error) {
	return s.applyConfig(ctx, "pk", pkConfigReqToMap(data))
}

// ======================== welcome ========================

// GetWelcomeConfig 用于获取进房欢迎模块配置
func (s *Service) GetWelcomeConfig(ctx context.Context) (WelcomeConfigResp, int, error) {
	group, ok := s.configCache.GetGroup("welcome")
	if !ok {
		return WelcomeConfigResp{}, 0, nil
	}
	resp, err := toWelcomeConfigResp(group)
	if err != nil {
		return WelcomeConfigResp{}, 60501, err
	}
	return resp, 0, nil
}

// ApplyWelcomeConfig 用于存储进房欢迎模块配置
func (s *Service) ApplyWelcomeConfig(ctx context.Context, data WelcomeConfigReq) (int, error) {
	return s.applyConfig(ctx, "welcome", welcomeConfigReqToMap(data))
}

// ======================== follow ========================

// GetFollowConfig 用于获取感谢关注模块配置
func (s *Service) GetFollowConfig(ctx context.Context) (FollowConfigResp, int, error) {
	group, ok := s.configCache.GetGroup("follow")
	if !ok {
		return FollowConfigResp{}, 0, nil
	}
	resp, err := toFollowConfigResp(group)
	if err != nil {
		return FollowConfigResp{}, 60501, err
	}
	return resp, 0, nil
}

// ApplyFollowConfig 用于存储感谢关注模块配置
func (s *Service) ApplyFollowConfig(ctx context.Context, data FollowConfigReq) (int, error) {
	return s.applyConfig(ctx, "follow", followConfigReqToMap(data))
}

// ======================== share ========================

// GetShareConfig 用于获取感谢分享模块配置
func (s *Service) GetShareConfig(ctx context.Context) (ShareConfigResp, int, error) {
	group, ok := s.configCache.GetGroup("share")
	if !ok {
		return ShareConfigResp{}, 0, nil
	}
	resp, err := toShareConfigResp(group)
	if err != nil {
		return ShareConfigResp{}, 60501, err
	}
	return resp, 0, nil
}

// ApplyShareConfig 用于存储感谢分享模块配置
func (s *Service) ApplyShareConfig(ctx context.Context, data ShareConfigReq) (int, error) {
	return s.applyConfig(ctx, "share", shareConfigReqToMap(data))
}

// ======================== reply ========================

// GetReplyConfig 用于获取自动回复模块配置
func (s *Service) GetReplyConfig(ctx context.Context) (ReplyConfigResp, int, error) {
	group, ok := s.configCache.GetGroup("reply")
	if !ok {
		return ReplyConfigResp{}, 0, nil
	}
	resp, err := toReplyConfigResp(group)
	if err != nil {
		return ReplyConfigResp{}, 60501, err
	}
	return resp, 0, nil
}

// ApplyReplyConfig 用于存储自动回复模块配置
func (s *Service) ApplyReplyConfig(ctx context.Context, data ReplyConfigReq) (int, error) {
	return s.applyConfig(ctx, "reply", replyConfigReqToMap(data))
}
