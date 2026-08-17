package robotconfig

import (
	"encoding/json"
	"fmt"
)

// getString 从缓存 map 中获取字符串值，不存在则返回空字符串
func getString(group map[string]string, key string) string {
	if v, ok := group[key]; ok {
		return v
	}
	return ""
}

// getStringSlice 从缓存 map 中获取 JSON 数组字符串并反序列化为 []string
// 若值为空字符串则返回空切片，若反序列化失败则返回错误
func getStringSlice(group map[string]string, key string) ([]string, error) {
	raw := getString(group, key)
	if raw == "" {
		return []string{}, nil
	}
	var result []string
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("解析 %s 失败: %w", key, err)
	}
	return result, nil
}

// getReplyItems 从缓存 map 中获取 JSON 数组并反序列化为 []ReplyItem
func getReplyItems(group map[string]string, key string) ([]ReplyItem, error) {
	raw := getString(group, key)
	if raw == "" {
		return []ReplyItem{}, nil
	}
	var result []ReplyItem
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("解析 %s 失败: %w", key, err)
	}
	return result, nil
}

// marshalSlice 将 []string 序列化为 JSON 字符串，空切片返回 "[]"
func marshalSlice(v []string) string {
	if v == nil {
		v = []string{}
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// marshalReplyItems 将 []ReplyItem 序列化为 JSON 字符串
func marshalReplyItems(v []ReplyItem) string {
	if v == nil {
		v = []ReplyItem{}
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// ======================== room ========================

func toRoomConfigResp(group map[string]string) RoomConfigResp {
	return RoomConfigResp{
		IsListening:           getString(group, "is_listening"),
		MaxNameLength:         getString(group, "max_name_length"),
		NameTrimMode:          getString(group, "name_trim_mode"),
		ConsumeRewardEnabled:  getString(group, "consume_reward_enabled"),
		RewardType:            getString(group, "reward_type"),
		ConsumeBatteryRate:    getString(group, "consume_battery_rate"),
		CaptainRewardAmount:   getString(group, "captain_reward_amount"),
		CommanderRewardAmount: getString(group, "commander_reward_amount"),
		GovernorRewardAmount:  getString(group, "governor_reward_amount"),
	}
}

func roomConfigReqToMap(req RoomConfigReq) map[string]string {
	return map[string]string{
		"is_listening":            req.IsListening,
		"max_name_length":         req.MaxNameLength,
		"name_trim_mode":          req.NameTrimMode,
		"consume_reward_enabled":  req.ConsumeRewardEnabled,
		"reward_type":             req.RewardType,
		"consume_battery_rate":    req.ConsumeBatteryRate,
		"captain_reward_amount":   req.CaptainRewardAmount,
		"commander_reward_amount": req.CommanderRewardAmount,
		"governor_reward_amount":  req.GovernorRewardAmount,
	}
}

// ======================== sign ========================

func toSignConfigResp(group map[string]string) (SignConfigResp, error) {
	successReply, err := getStringSlice(group, "success_reply")
	if err != nil {
		return SignConfigResp{}, err
	}
	failReply, err := getStringSlice(group, "fail_reply")
	if err != nil {
		return SignConfigResp{}, err
	}
	repeatReply, err := getStringSlice(group, "repeat_reply")
	if err != nil {
		return SignConfigResp{}, err
	}
	queryReply, err := getStringSlice(group, "query_reply")
	if err != nil {
		return SignConfigResp{}, err
	}
	return SignConfigResp{
		Enabled:      getString(group, "enabled"),
		Scene:        getString(group, "scene"),
		Requirement:  getString(group, "requirement"),
		RewardType:   getString(group, "reward_type"),
		RewardAmount: getString(group, "reward_amount"),
		Keyword:      getString(group, "keyword"),
		QueryKeyword: getString(group, "query_keyword"),
		SuccessReply: successReply,
		FailReply:    failReply,
		RepeatReply:  repeatReply,
		QueryReply:   queryReply,
	}, nil
}

func signConfigReqToMap(req SignConfigReq) map[string]string {
	return map[string]string{
		"enabled":       req.Enabled,
		"scene":         req.Scene,
		"requirement":   req.Requirement,
		"reward_type":   req.RewardType,
		"reward_amount": req.RewardAmount,
		"keyword":       req.Keyword,
		"query_keyword": req.QueryKeyword,
		"success_reply": marshalSlice(req.SuccessReply),
		"fail_reply":    marshalSlice(req.FailReply),
		"repeat_reply":  marshalSlice(req.RepeatReply),
		"query_reply":   marshalSlice(req.QueryReply),
	}
}

// ======================== ad ========================

func toAdConfigResp(group map[string]string) (AdConfigResp, error) {
	content, err := getStringSlice(group, "content")
	if err != nil {
		return AdConfigResp{}, err
	}
	return AdConfigResp{
		Enabled:  getString(group, "enabled"),
		Scene:    getString(group, "scene"),
		Interval: getString(group, "interval"),
		SendMode: getString(group, "send_mode"),
		Content:  content,
	}, nil
}

func adConfigReqToMap(req AdConfigReq) map[string]string {
	return map[string]string{
		"enabled":   req.Enabled,
		"scene":     req.Scene,
		"interval":  req.Interval,
		"send_mode": req.SendMode,
		"content":   marshalSlice(req.Content),
	}
}

// ======================== gift ========================

func toGiftConfigResp(group map[string]string) (GiftConfigResp, error) {
	content, err := getStringSlice(group, "content")
	if err != nil {
		return GiftConfigResp{}, err
	}
	return GiftConfigResp{
		Enabled:         getString(group, "enabled"),
		Scene:           getString(group, "scene"),
		Requirement:     getString(group, "requirement"),
		ShowCount:       getString(group, "show_count"),
		MergeGift:       getString(group, "merge_gift"),
		IncludeBlindbox: getString(group, "include_blindbox"),
		MinBattery:      getString(group, "min_battery"),
		Content:         content,
	}, nil
}

func giftConfigReqToMap(req GiftConfigReq) map[string]string {
	return map[string]string{
		"enabled":          req.Enabled,
		"scene":            req.Scene,
		"requirement":      req.Requirement,
		"show_count":       req.ShowCount,
		"merge_gift":       req.MergeGift,
		"include_blindbox": req.IncludeBlindbox,
		"min_battery":      req.MinBattery,
		"content":          marshalSlice(req.Content),
	}
}

// ======================== pk ========================

func toPkConfigResp(group map[string]string) (PkConfigResp, error) {
	content, err := getStringSlice(group, "content")
	if err != nil {
		return PkConfigResp{}, err
	}
	return PkConfigResp{
		Enabled: getString(group, "enabled"),
		Content: content,
	}, nil
}

func pkConfigReqToMap(req PkConfigReq) map[string]string {
	return map[string]string{
		"enabled": req.Enabled,
		"content": marshalSlice(req.Content),
	}
}

// ======================== welcome ========================

func toWelcomeConfigResp(group map[string]string) (WelcomeConfigResp, error) {
	content, err := getStringSlice(group, "content")
	if err != nil {
		return WelcomeConfigResp{}, err
	}
	return WelcomeConfigResp{
		Enabled:     getString(group, "enabled"),
		Scene:       getString(group, "scene"),
		Requirement: getString(group, "requirement"),
		Content:     content,
	}, nil
}

func welcomeConfigReqToMap(req WelcomeConfigReq) map[string]string {
	return map[string]string{
		"enabled":     req.Enabled,
		"scene":       req.Scene,
		"requirement": req.Requirement,
		"content":     marshalSlice(req.Content),
	}
}

// ======================== follow ========================

func toFollowConfigResp(group map[string]string) (FollowConfigResp, error) {
	content, err := getStringSlice(group, "content")
	if err != nil {
		return FollowConfigResp{}, err
	}
	return FollowConfigResp{
		Enabled:     getString(group, "enabled"),
		Scene:       getString(group, "scene"),
		Requirement: getString(group, "requirement"),
		Content:     content,
	}, nil
}

func followConfigReqToMap(req FollowConfigReq) map[string]string {
	return map[string]string{
		"enabled":     req.Enabled,
		"scene":       req.Scene,
		"requirement": req.Requirement,
		"content":     marshalSlice(req.Content),
	}
}

// ======================== share ========================

func toShareConfigResp(group map[string]string) (ShareConfigResp, error) {
	content, err := getStringSlice(group, "content")
	if err != nil {
		return ShareConfigResp{}, err
	}
	return ShareConfigResp{
		Enabled:     getString(group, "enabled"),
		Scene:       getString(group, "scene"),
		Requirement: getString(group, "requirement"),
		Content:     content,
	}, nil
}

func shareConfigReqToMap(req ShareConfigReq) map[string]string {
	return map[string]string{
		"enabled":     req.Enabled,
		"scene":       req.Scene,
		"requirement": req.Requirement,
		"content":     marshalSlice(req.Content),
	}
}

// ======================== reply ========================

func toReplyConfigResp(group map[string]string) (ReplyConfigResp, error) {
	content, err := getReplyItems(group, "content")
	if err != nil {
		return ReplyConfigResp{}, err
	}
	return ReplyConfigResp{
		Enabled:     getString(group, "enabled"),
		Scene:       getString(group, "scene"),
		Requirement: getString(group, "requirement"),
		Content:     content,
	}, nil
}

func replyConfigReqToMap(req ReplyConfigReq) map[string]string {
	return map[string]string{
		"enabled":     req.Enabled,
		"scene":       req.Scene,
		"requirement": req.Requirement,
		"content":     marshalReplyItems(req.Content),
	}
}
