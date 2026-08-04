// Package live 提供 B站 直播间的实时交互能力：WebSocket 消息监听和弹幕发送队列
//
// 本包与 room 包的区别：
//   - room 包：HTTP API 调用（获取直播间信息、发送弹幕接口等）
//   - live 包：长连接实时操作（WebSocket 监听、弹幕队列发送）
package live

import (
	"encoding/json"
	"fmt"
)

// Cmd 是 B站 直播 WebSocket 协议的消息命令类型
type Cmd string

// B站 直播 WebSocket 常见命令字
const (
	CmdLiveStart     Cmd = "LIVE"               // 直播开始
	CmdLiveCutOff    Cmd = "CUT_OFF"            // 直播被超管切断
	CmdLiveRoomLock  Cmd = "ROOM_LOCK"          // 直播间被封
	CmdLiveEnd       Cmd = "PREPARING"          // 直播结束（下播）
	CmdSendGift      Cmd = "SEND_GIFT"          // 送礼消息
	CmdSendGiftV2    Cmd = "SEND_GIFT_V2"       // 送礼消息V2
	CmdGuardBuy      Cmd = "GUARD_BUY"          // 大航海（舰长/提督/总督）购买
	CmdInteractWord  Cmd = "INTERACT_WORD_V2"   // 用户互动（关注、分享等）
	CmdDanmuMsg      Cmd = "DANMU_MSG"          // 弹幕消息
	CmdPkStart       Cmd = "PK_BATTLE_PRE_NEW"  // PK即将开始
	CmdSuperDanmuMsg Cmd = "SUPER_CHAT_MESSAGE" // 醒目留言
)

// Message 表示一条从直播间 WebSocket 收到的消息
//
// Raw 保留原始 JSON 数据，调用方可根据 Cmd 类型自行反序列化到具体结构体
type Message struct {
	Cmd Cmd             `json:"cmd"` // 消息命令字
	Raw json.RawMessage `json:"-"`   // 原始 JSON（不参与序列化，仅用于数据透传）
}

// DanmuMsgInfo 是 DANMU_MSG 的 info 字段结构
// info 数组各元素含义：
//
//	info[0] — 弹幕展示信息 []any（字体、颜色、表情等）
//	info[1] — 弹幕文本内容 string
//	info[2] — 用户信息 [uid, uname, is_admin, is_vip, is_svip, ...]
//	info[3] — 勋章信息 [level, name, anchor_name, room_id, ...] 或空数组 []
//	info[4] — 用户等级信息 [level, ...]
//	info[5] — 用户头像框等信息
type DanmuMsgInfo struct {
	UID         json.Number `json:"-"` // info[2][0] 用户UID
	Uname       string      `json:"-"` // info[2][1] 用户名
	Msg         string      `json:"-"` // info[1] 弹幕内容
	BadgeUID    json.Number `json:"-"` // info[3][12] 勋章主播UID
	BadgeUname  string      `json:"-"` // info[3][2] 勋章主播名
	BadgeRoomID json.Number `json:"-"` // info[3][3] 勋章房间ID
	BadgeName   string      `json:"-"` // info[3][1] 勋章名称
	BadgeLevel  json.Number `json:"-"` // info[3][0] 勋章等级
	BadgeType   json.Number `json:"-"` // info[3][10] 勋章类型 0=普通用户，1=总督，2=提督，3=舰长
}

// LiveInfo 是 LIVE 消息中提取的关键字段
type LiveInfo struct {
	LiveKey      string `json:"live_key"`      // 直播流标识 key
	LivePlatform string `json:"live_platform"` // 开播平台（pc_link / android / ios 等）
}

// CutOffInfo 是 CUT_OFF 消息中提取的关键字段
type CutOffInfo struct {
	MsgID    string `json:"msg_id"`    // 消息ID
	RoomID   int64  `json:"room_id"`   // 直播间ID
	SendTime int64  `json:"send_time"` // 发生时间（毫秒级时间戳）
	Msg      string `json:"msg"`       // 消息内容
}

// RoomLockInfo 是 ROOM_LOCK 消息中提取的关键字段
type RoomLockInfo struct {
	MsgID    string `json:"msg_id"`    // 消息ID
	RoomID   int64  `json:"room_id"`   // 直播间ID
	SendTime int64  `json:"send_time"` // 发生时间（毫秒级时间戳）
}

// PreparingInfo 是 PREPARING 消息中提取的关键字段
type PreparingInfo struct {
	MsgID    string `json:"msg_id"`    // 消息ID
	RoomID   string `json:"roomid"`    // 直播间ID
	SendTime int64  `json:"send_time"` // 发生时间（毫秒级时间戳）
}

// BlindGiftInfo 是 SEND_GIFT 中的盲盒礼物信息
type BlindGiftInfo struct {
	GiftAction        string      `json:"gift_action"`        // 盲盒动作
	GiftTipPrice      int64       `json:"-"`                  // 爆出礼物价格(分)
	OriginalGiftID    json.Number `json:"original_gift_id"`   // 原始礼物ID
	OriginalGiftName  string      `json:"original_gift_name"` // 原始礼物名称
	OriginalGiftPrice int64       `json:"-"`                  // 原始礼物价格(分)
}

// SendGiftInfo 是 SEND_GIFT 消息中提取的关键字段
type SendGiftInfo struct {
	UID        json.Number    `json:"-"` // 送礼用户UID
	Uname      string         `json:"-"` // 送礼用户名
	GiftID     json.Number    `json:"-"` // 礼物ID
	GiftName   string         `json:"-"` // 礼物名称
	Price      int64          `json:"-"` // 礼物价格(分)
	Num        json.Number    `json:"-"` // 礼物数量
	AnchorID   json.Number    `json:"-"` // 主播UID
	BadgeUID   json.Number    `json:"-"` // 勋章主播UID
	BadgeName  string         `json:"-"` // 勋章名称
	BadgeLevel json.Number    `json:"-"` // 勋章等级
	BadgeType  json.Number    `json:"-"` // 勋章类型 0=普通用户，1=总督，2=提督，3=舰长
	BlindGift  *BlindGiftInfo `json:"-"` // 盲盒礼物信息，非盲盒时为 nil
}

// ExtractSendGift 从原始 JSON 中提取送礼信息
func ExtractSendGift(raw string) (*SendGiftInfo, error) {
	// 解析外层 JSON，拿到 data 子对象
	var outer struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &outer); err != nil {
		return nil, fmt.Errorf("解析 SEND_GIFT 外层消息失败: %w", err)
	}
	// 解析 data 对象，嵌套字段用匿名结构体
	var data struct {
		UID           json.Number `json:"uid"`
		Uname         string      `json:"uname"`
		GiftID        json.Number `json:"giftId"`
		GiftName      string      `json:"giftName"`
		Price         int64       `json:"price"`
		Num           json.Number `json:"num"`
		ReceiverUinfo struct {
			UID json.Number `json:"uid"`
		} `json:"receiver_uinfo"`
		SenderUinfo struct {
			Medal *struct {
				RUID       json.Number `json:"ruid"`
				Name       string      `json:"name"`
				GuardLevel json.Number `json:"guard_level"`
				Level      json.Number `json:"level"`
			} `json:"medal"`
		} `json:"sender_uinfo"`
		BlindGift json.RawMessage `json:"blind_gift"`
	}
	if err := json.Unmarshal(outer.Data, &data); err != nil {
		return nil, fmt.Errorf("解析 SEND_GIFT data 字段失败: %w", err)
	}
	result := &SendGiftInfo{
		UID:      data.UID,
		Uname:    data.Uname,
		GiftID:   data.GiftID,
		GiftName: data.GiftName,
		Price:    data.Price / 10,
		Num:      data.Num,
		AnchorID: data.ReceiverUinfo.UID,
	}
	// 勋章信息 — 用户未佩戴勋章时为 null
	if data.SenderUinfo.Medal != nil {
		result.BadgeUID = data.SenderUinfo.Medal.RUID
		result.BadgeName = data.SenderUinfo.Medal.Name
		result.BadgeLevel = data.SenderUinfo.Medal.Level
		result.BadgeType = data.SenderUinfo.Medal.GuardLevel
	}
	// 盲盒礼物 - 非盲盒礼物时为 null
	if len(data.BlindGift) > 0 && string(data.BlindGift) != "null" {
		var raw struct {
			GiftAction        string      `json:"gift_action"`
			GiftTipPrice      int64       `json:"gift_tip_price"`
			OriginalGiftID    json.Number `json:"original_gift_id"`
			OriginalGiftName  string      `json:"original_gift_name"`
			OriginalGiftPrice int64       `json:"original_gift_price"`
		}
		if err := json.Unmarshal(data.BlindGift, &raw); err != nil {
			return nil, fmt.Errorf("解析 SEND_GIFT blind_gift 字段失败: %w", err)
		}
		result.BlindGift = &BlindGiftInfo{
			GiftAction:        raw.GiftAction,
			GiftTipPrice:      raw.GiftTipPrice / 10,
			OriginalGiftID:    raw.OriginalGiftID,
			OriginalGiftName:  raw.OriginalGiftName,
			OriginalGiftPrice: raw.OriginalGiftPrice / 10,
		}
	}
	return result, nil
}

// ExtractPreparing 从原始 JSON 中提取直播结束信息
func ExtractPreparing(raw string) (*PreparingInfo, error) {
	var info PreparingInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		return nil, fmt.Errorf("解析 PREPARING 消息失败: %w", err)
	}
	return &info, nil
}

// ExtractRoomLock 从原始 JSON 中提取直播间被封信息
func ExtractRoomLock(raw string) (*RoomLockInfo, error) {
	var info RoomLockInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		return nil, fmt.Errorf("解析 ROOM_LOCK 消息失败: %w", err)
	}
	return &info, nil
}

// ExtractCutOff 从原始 JSON 中提取直播被超管切断信息
func ExtractCutOff(raw string) (*CutOffInfo, error) {
	var info CutOffInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		return nil, fmt.Errorf("解析 CUT_OFF 消息失败: %w", err)
	}
	return &info, nil
}

// ExtractLive 从原始 JSON 中提取直播开始信息
func ExtractLive(raw string) (*LiveInfo, error) {
	var info LiveInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		return nil, fmt.Errorf("解析 LIVE 消息失败: %w", err)
	}
	return &info, nil
}

// extractDanmuMsg 从原始 JSON 中提取弹幕信息
func ExtractDanmuMsg(raw string) (*DanmuMsgInfo, error) {
	var outer struct {
		Info []json.RawMessage `json:"info"`
	}
	if err := json.Unmarshal([]byte(raw), &outer); err != nil {
		return nil, fmt.Errorf("解析外层 JSON 失败: %w", err)
	}
	if len(outer.Info) < 4 {
		return nil, fmt.Errorf("info 数组长度不足（需要至少 4 个元素，实际 %d 个）", len(outer.Info))
	}

	result := &DanmuMsgInfo{}

	// info[1] — 弹幕文本
	if err := json.Unmarshal(outer.Info[1], &result.Msg); err != nil {
		return nil, fmt.Errorf("解析 info[1]（弹幕文本）失败: %w", err)
	}

	// info[2] — 用户信息数组 [uid, uname, ...]
	var userInfo []json.RawMessage
	if err := json.Unmarshal(outer.Info[2], &userInfo); err != nil {
		return nil, fmt.Errorf("解析 info[2]（用户信息）失败: %w", err)
	}
	if len(userInfo) >= 2 {
		_ = json.Unmarshal(userInfo[0], &result.UID)
		_ = json.Unmarshal(userInfo[1], &result.Uname)
	}

	// info[3] — 勋章信息数组（可能为空 []）
	var medalInfo []json.RawMessage
	if err := json.Unmarshal(outer.Info[3], &medalInfo); err != nil {
		return nil, fmt.Errorf("解析 info[3]（勋章信息）失败: %w", err)
	}
	if len(medalInfo) > 0 {
		if len(medalInfo) >= 1 {
			_ = json.Unmarshal(medalInfo[0], &result.BadgeLevel)
		}
		if len(medalInfo) >= 2 {
			_ = json.Unmarshal(medalInfo[1], &result.BadgeName)
		}
		if len(medalInfo) >= 3 {
			_ = json.Unmarshal(medalInfo[2], &result.BadgeUname)
		}
		if len(medalInfo) >= 4 {
			_ = json.Unmarshal(medalInfo[3], &result.BadgeRoomID)
		}
		if len(medalInfo) >= 11 {
			_ = json.Unmarshal(medalInfo[10], &result.BadgeType)
		}
		if len(medalInfo) >= 13 {
			_ = json.Unmarshal(medalInfo[12], &result.BadgeUID)
		}
	}

	return result, nil
}
