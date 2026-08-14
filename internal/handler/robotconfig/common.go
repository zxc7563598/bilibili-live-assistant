package robotconfig

import (
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/input"
	"github.com/zxc7563598/bilibili-live-assistant/internal/dto/resp"
	robotconfigsvc "github.com/zxc7563598/bilibili-live-assistant/internal/service/robotconfig"
)

func toRoomConfigResp(svcResp robotconfigsvc.RoomConfigResp) resp.RoomConfigResp {
	return resp.RoomConfigResp{
		IsListening:   svcResp.IsListening,
		MaxNameLength: svcResp.MaxNameLength,
		NameTrimMode:  svcResp.NameTrimMode,
	}
}

func toSignConfigResp(svcResp robotconfigsvc.SignConfigResp) resp.SignConfigResp {
	return resp.SignConfigResp{
		Enabled:      svcResp.Enabled,
		Scene:        svcResp.Scene,
		Requirement:  svcResp.Requirement,
		RewardType:   svcResp.RewardType,
		RewardAmount: svcResp.RewardAmount,
		Keyword:      svcResp.Keyword,
		QueryKeyword: svcResp.QueryKeyword,
		SuccessReply: svcResp.SuccessReply,
		FailReply:    svcResp.FailReply,
		RepeatReply:  svcResp.RepeatReply,
		QueryReply:   svcResp.QueryReply,
	}
}

func toAdConfigResp(svcResp robotconfigsvc.AdConfigResp) resp.AdConfigResp {
	return resp.AdConfigResp{
		Enabled:  svcResp.Enabled,
		Scene:    svcResp.Scene,
		Interval: svcResp.Interval,
		SendMode: svcResp.SendMode,
		Content:  svcResp.Content,
	}
}

func toGiftConfigResp(svcResp robotconfigsvc.GiftConfigResp) resp.GiftConfigResp {
	return resp.GiftConfigResp{
		Enabled:         svcResp.Enabled,
		Scene:           svcResp.Scene,
		Requirement:     svcResp.Requirement,
		ShowCount:       svcResp.ShowCount,
		MergeGift:       svcResp.MergeGift,
		IncludeBlindbox: svcResp.IncludeBlindbox,
		MinBattery:      svcResp.MinBattery,
		Content:         svcResp.Content,
	}
}

func toPkConfigResp(svcResp robotconfigsvc.PkConfigResp) resp.PkConfigResp {
	return resp.PkConfigResp{
		Enabled: svcResp.Enabled,
		Content: svcResp.Content,
	}
}

func toWelcomeConfigResp(svcResp robotconfigsvc.WelcomeConfigResp) resp.WelcomeConfigResp {
	return resp.WelcomeConfigResp{
		Enabled:     svcResp.Enabled,
		Scene:       svcResp.Scene,
		Requirement: svcResp.Requirement,
		Content:     svcResp.Content,
	}
}

func toFollowConfigResp(svcResp robotconfigsvc.FollowConfigResp) resp.FollowConfigResp {
	return resp.FollowConfigResp{
		Enabled:     svcResp.Enabled,
		Scene:       svcResp.Scene,
		Requirement: svcResp.Requirement,
		Content:     svcResp.Content,
	}
}

func toShareConfigResp(svcResp robotconfigsvc.ShareConfigResp) resp.ShareConfigResp {
	return resp.ShareConfigResp{
		Enabled:     svcResp.Enabled,
		Scene:       svcResp.Scene,
		Requirement: svcResp.Requirement,
		Content:     svcResp.Content,
	}
}

func toReplyConfigResp(svcResp robotconfigsvc.ReplyConfigResp) resp.ReplyConfigResp {
	return resp.ReplyConfigResp{
		Enabled:     svcResp.Enabled,
		Scene:       svcResp.Scene,
		Requirement: svcResp.Requirement,
		Content:     toReplyItems(svcResp.Content),
	}
}

func toReplyItems(svcItems []robotconfigsvc.ReplyItem) []resp.ReplyItem {
	items := make([]resp.ReplyItem, 0, len(svcItems))
	for _, v := range svcItems {
		items = append(items, resp.ReplyItem{
			Keyword:             v.Keyword,
			KeywordMatchPolicy:  v.KeywordMatchPolicy,
			SafeWord:            v.SafeWord,
			SafeWordMatchPolicy: v.SafeWordMatchPolicy,
			MuteSender:          v.MuteSender,
			MuteDuration:        v.MuteDuration,
			RansomAmount:        v.RansomAmount,
			ReplyContent:        v.ReplyContent,
		})
	}
	return items
}

func fromReplyItems(inputItems []input.ReplyItem) []robotconfigsvc.ReplyItem {
	items := make([]robotconfigsvc.ReplyItem, 0, len(inputItems))
	for _, v := range inputItems {
		items = append(items, robotconfigsvc.ReplyItem{
			Keyword:             v.Keyword,
			KeywordMatchPolicy:  v.KeywordMatchPolicy,
			SafeWord:            v.SafeWord,
			SafeWordMatchPolicy: v.SafeWordMatchPolicy,
			MuteSender:          v.MuteSender,
			MuteDuration:        v.MuteDuration,
			RansomAmount:        v.RansomAmount,
			ReplyContent:        v.ReplyContent,
		})
	}
	return items
}
