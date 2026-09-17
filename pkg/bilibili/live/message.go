// Package live 提供 B站 直播间的实时交互能力：WebSocket 消息监听和弹幕发送队列
//
// 本包与 room 包的区别：
//   - room 包：HTTP API 调用（获取直播间信息、发送弹幕接口等）
//   - live 包：长连接实时操作（WebSocket 监听、弹幕队列发送）
package live

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/internal/protobuf/interactwordv2"
	"github.com/zxc7563598/bilibili-live-assistant/pkg/bilibili/internal/protobuf/sendgiftv2"
	"google.golang.org/protobuf/proto"
)

// Cmd 是 B站 直播 WebSocket 协议的消息命令类型
type Cmd string

// B站 直播 WebSocket 常见命令字
const (
	CmdLiveStart     Cmd = "LIVE"                 // 直播开始
	CmdLiveCutOff    Cmd = "CUT_OFF"              // 直播被超管切断
	CmdLiveRoomLock  Cmd = "ROOM_LOCK"            // 直播间被封
	CmdLiveEnd       Cmd = "PREPARING"            // 直播结束（下播）
	CmdSendGift      Cmd = "SEND_GIFT"            // 送礼消息
	CmdSendGiftV2    Cmd = "SEND_GIFT_V2"         // 送礼消息V2
	CmdGuardBuy      Cmd = "GUARD_BUY"            // 大航海（舰长/提督/总督）购买
	CmdInteractWord  Cmd = "INTERACT_WORD_V2"     // 用户互动（关注、分享等）
	CmdDanmuMsg      Cmd = "DANMU_MSG"            // 弹幕消息
	CmdPkStart       Cmd = "PK_BATTLE_PRE_NEW"    // PK即将开始
	CmdPkBattleEnd   Cmd = "PK_BATTLE_END"        // PK结束（battle_type=2 经典PK）
	CmdPkSettleNew   Cmd = "PK_BATTLE_SETTLE_NEW" // PK结算（battle_type=6 大乱斗）
	CmdSuperDanmuMsg Cmd = "SUPER_CHAT_MESSAGE"   // 醒目留言
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
	UID         int64  `json:"uid"`           // info[2][0] 用户UID
	Uname       string `json:"username"`      // info[2][1] 用户名
	Msg         string `json:"content"`       // info[1] 弹幕内容
	BadgeUID    int64  `json:"badge_uid"`     // info[3][12] 勋章主播UID
	BadgeUname  string `json:"badge_uname"`   // info[3][2] 勋章主播名
	BadgeRoomID int64  `json:"badge_room_id"` // info[3][3] 勋章房间ID
	BadgeName   string `json:"badge_name"`    // info[3][1] 勋章名称
	BadgeLevel  int64  `json:"badge_level"`   // info[3][0] 勋章等级
	BadgeType   int64  `json:"badge_type"`    // info[3][10] 勋章类型 0=普通用户，1=总督，2=提督，3=舰长
}

// LiveInfo 是 LIVE 消息中提取的关键字段
type LiveInfo struct {
	LiveKey      string `json:"live_key"`      // 直播流标识 key
	LivePlatform string `json:"live_platform"` // 开播平台（pc_link / android / ios 等）
	LiveTime     int64  `json:"live_time"`     // 开播时间（Unix 秒级时间戳）
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
	GiftAction        string `json:"gift_action"`         // 盲盒动作
	GiftTipPrice      int64  `json:"gift_tip_price"`      // 爆出礼物价格(分)
	OriginalGiftID    int64  `json:"original_gift_id"`    // 原始礼物ID
	OriginalGiftName  string `json:"original_gift_name"`  // 原始礼物名称
	OriginalGiftPrice int64  `json:"original_gift_price"` // 原始礼物价格(分)
}

// GiftItemInfo 是单份礼物的信息
//
// 一次送礼广播可能携带多份礼物：普通连击只有 1 份，盲盒爆出时可以有多份
// （例如 100 个盲盒炸成虚天秘境/黄枫谷/燕家堡/血色禁地 4 种），因此用列表承载。
type GiftItemInfo struct {
	GiftID    int64  `json:"gift_id"`    // 礼物ID
	GiftName  string `json:"gift_name"`  // 礼物名称
	Num       int64  `json:"num"`        // 礼物数量
	Price     int64  `json:"price"`      // 礼物标价单价(分)
	TotalCoin int64  `json:"total_coin"` // 本项实际消耗(分)，盲盒爆出时按盲盒单价计
}

// SendGiftInfo 是 SEND_GIFT 与 SEND_GIFT_V2 消息中提取的关键字段
type SendGiftInfo struct {
	UID        int64          `json:"uid"`                  // 送礼用户UID
	Uname      string         `json:"username"`             // 送礼用户名
	AnchorID   int64          `json:"anchor_id"`            // 主播UID
	BadgeUID   int64          `json:"badge_uid"`            // 勋章主播UID
	BadgeName  string         `json:"badge_name"`           // 勋章名称
	BadgeLevel int64          `json:"badge_level"`          // 勋章等级
	BadgeType  int64          `json:"badge_type"`           // 勋章类型 0=普通用户，1=总督，2=提督，3=舰长
	BlindGift  *BlindGiftInfo `json:"blind_gift,omitempty"` // 盲盒礼物信息，非盲盒时为 nil
	Gifts      []GiftItemInfo `json:"gifts"`                // 礼物明细，普通送礼 1 份；盲盒爆出时多份，gift_list 缺失时为空
}

// GuardBuyInfo 是 GUARD_BUY 消息中提取的关键字段
type GuardBuyInfo struct {
	UID        int64  `json:"uid"`         // 送礼用户UID
	Uname      string `json:"username"`    // 送礼用户名
	GiftID     int64  `json:"gift_id"`     // 礼物ID
	GiftName   string `json:"gift_name"`   // 礼物名称(舰长\提督\总督)
	GuardLevel int64  `json:"guard_level"` // 航海类型 0=普通用户，1=总督，2=提督，3=舰长
	Num        int64  `json:"num"`         // 数量
	Price      int64  `json:"price"`       // 价格(分)
	SendTime   int64  `json:"start_time"`  // 发生时间（秒级时间戳）
}

// InteractWordV2Info 是 INTERACT_WORD_V2 消息中提取的关键字段
type InteractWordV2Info struct {
	UID        int64  `json:"uid"`         // 用户UID
	Uname      string `json:"username"`    // 用户名
	MsgType    int64  `json:"msg_type"`    // 消息类型 1=进入直播间 2=关注 3=分享
	RoomID     int64  `json:"room_id"`     // 直播间ID
	Timestamp  int64  `json:"timestamp"`   // 时间戳
	BadgeUID   int64  `json:"badge_uid"`   // 勋章主播UID
	BadgeName  string `json:"badge_name"`  // 勋章名称
	BadgeLevel int64  `json:"badge_level"` // 勋章等级
	BadgeType  int64  `json:"badge_type"`  // 勋章类型 0=普通用户 1=总督 2=提督 3=舰长
}

// PkBattlePreNewInfo 是 PK_BATTLE_PRE_NEW 消息中提取的关键字段
type PkBattlePreNewInfo struct {
	PkID       int64  `json:"pk_id"`       // PK ID
	PkStatus   int64  `json:"pk_status"`   // PK 状态
	Timestamp  int64  `json:"timestamp"`   // 时间戳
	Uname      string `json:"uname"`       // 对方用户名
	UID        int64  `json:"uid"`         // 对方用户UID
	RoomID     int64  `json:"room_id"`     // 对方房间ID
	BattleType int64  `json:"battle_type"` // 对战类型
	MatchType  int64  `json:"match_type"`  // 匹配类型
}

// PkSideInfo 是 PK 对战其中一方的结算数据
//
// 对应 PK_BATTLE_END / PK_BATTLE_SETTLE_NEW 中的 init_info（发起方）与 match_info（被匹配方）。
type PkSideInfo struct {
	RoomID int64 `json:"room_id"` // 房间ID
	Votes  int64 `json:"votes"`   // 最终PK值
	// 胜负字段：两个事件对该字段的命名不同，这里按各自的原始名称原样接收，
	// 未出现的一方恒为 0，具体怎么判定交给业务层
	WinnerType int64 `json:"winner_type"`
	ResultType int64 `json:"result_type"`
}

// PkBattleEndInfo 是 PK_BATTLE_END 消息中提取的关键字段（battle_type=2 经典PK）
type PkBattleEndInfo struct {
	PkID       int64       `json:"pk_id"`       // PK ID
	PkStatus   int64       `json:"pk_status"`   // PK 状态
	Timestamp  int64       `json:"timestamp"`   // 秒级时间戳
	SendTime   int64       `json:"send_time"`   // 毫秒级时间戳
	BattleType int64       `json:"battle_type"` // 对战类型
	InitInfo   *PkSideInfo `json:"init_info"`   // 发起方
	MatchInfo  *PkSideInfo `json:"match_info"`  // 被匹配方
	Timer      int64       `json:"timer"`       // 冻结倒计时
	ShowStreak bool        `json:"show_streak"` // 是否展示连胜
}

// PkBattleSettleNewInfo 是 PK_BATTLE_SETTLE_NEW 消息中提取的关键字段（battle_type=6 大乱斗）
type PkBattleSettleNewInfo struct {
	PkID          int64       `json:"pk_id"`           // PK ID
	PkStatus      int64       `json:"pk_status"`       // PK 状态
	Timestamp     int64       `json:"timestamp"`       // 秒级时间戳
	SendTime      int64       `json:"send_time"`       // 毫秒级时间戳
	BattleType    int64       `json:"battle_type"`     // 对战类型
	InitInfo      *PkSideInfo `json:"init_info"`       // 发起方
	MatchInfo     *PkSideInfo `json:"match_info"`      // 被匹配方
	PunishName    string      `json:"punish_name"`     // 惩罚名称
	PunishEndTime int64       `json:"punish_end_time"` // 惩罚结束时间（秒级）
	SettleStatus  int64       `json:"settle_status"`   // 结算状态
	DmScore       int64       `json:"dmscore"`         // 弹幕分数
}

// SuperChatMessage 是 SUPER_CHAT_MESSAGE 消息中提取的关键字段
type SuperChatMessage struct {
	GiftID           int64  `json:"gift_id"`            // 礼物ID
	GiftName         string `json:"gift_name"`          // 礼物名称
	GiftNum          int64  `json:"gift_num"`           // 礼物数量
	UID              int64  `json:"uid"`                // 用户UID
	Uname            string `json:"username"`           // 用户名
	Message          string `json:"message"`            // 留言内容
	Price            int64  `json:"price"`              // 价格
	StartTime        int64  `json:"start_time"`         // 开始时间（秒级时间戳）
	BadgeLevel       int64  `json:"badge_level"`        // 勋章等级
	BadgeName        string `json:"badge_name"`         // 勋章名称
	BadgeRoomID      int64  `json:"badge_room_id"`      // 勋章主播房间ID
	BadgeAnchorUname string `json:"badge_anchor_uname"` // 勋章主播名
	BadgeType        int64  `json:"badge_type"`         // 勋章类型 0=普通用户 1=总督 2=提督 3=舰长
}

// ExtractSuperChatMessage 从原始 JSON 中提取醒目留言信息
func ExtractSuperChatMessage(raw string) (*SuperChatMessage, error) {
	var outer struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &outer); err != nil {
		return nil, fmt.Errorf("解析 SUPER_CHAT_MESSAGE 外层消息失败: %w", err)
	}
	var data struct {
		Gift struct {
			GiftID   int64  `json:"gift_id"`
			GiftName string `json:"gift_name"`
			Num      int64  `json:"num"`
		} `json:"gift"`
		UID      int64 `json:"uid"`
		UserInfo struct {
			Uname string `json:"uname"`
		} `json:"user_info"`
		Message   string `json:"message"`
		Price     int64  `json:"price"`
		StartTime int64  `json:"start_time"`
		MedalInfo *struct {
			MedalLevel   int64  `json:"medal_level"`
			MedalName    string `json:"medal_name"`
			AnchorRoomID int64  `json:"anchor_roomid"`
			AnchorUname  string `json:"anchor_uname"`
			GuardLevel   int64  `json:"guard_level"`
		} `json:"medal_info"`
	}
	if err := json.Unmarshal(outer.Data, &data); err != nil {
		return nil, fmt.Errorf("解析 SUPER_CHAT_MESSAGE data 字段失败: %w", err)
	}
	result := &SuperChatMessage{
		GiftID:    data.Gift.GiftID,
		GiftName:  data.Gift.GiftName,
		GiftNum:   data.Gift.Num,
		UID:       data.UID,
		Uname:     data.UserInfo.Uname,
		Message:   data.Message,
		Price:     data.Price * 100,
		StartTime: data.StartTime,
	}
	if data.MedalInfo != nil {
		m := data.MedalInfo
		result.BadgeLevel = m.MedalLevel
		result.BadgeName = m.MedalName
		result.BadgeRoomID = m.AnchorRoomID
		result.BadgeAnchorUname = m.AnchorUname
		result.BadgeType = m.GuardLevel
	}
	return result, nil
}

// ExtractPkBattlePreNew 从原始 JSON 中提取 PK 准备信息
func ExtractPkBattlePreNew(raw string) (*PkBattlePreNewInfo, error) {
	var outer struct {
		PkID      int64           `json:"pk_id"`
		PkStatus  int64           `json:"pk_status"`
		Timestamp int64           `json:"timestamp"`
		Data      json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &outer); err != nil {
		return nil, fmt.Errorf("解析 PK_BATTLE_PRE_NEW 外层消息失败: %w", err)
	}
	var data struct {
		Uname      string `json:"uname"`
		UID        int64  `json:"uid"`
		RoomID     int64  `json:"room_id"`
		BattleType int64  `json:"battle_type"`
		MatchType  int64  `json:"match_type"`
	}
	if err := json.Unmarshal(outer.Data, &data); err != nil {
		return nil, fmt.Errorf("解析 PK_BATTLE_PRE_NEW data 字段失败: %w", err)
	}
	return &PkBattlePreNewInfo{
		PkID:       outer.PkID,
		PkStatus:   outer.PkStatus,
		Timestamp:  outer.Timestamp,
		Uname:      data.Uname,
		UID:        data.UID,
		RoomID:     data.RoomID,
		BattleType: data.BattleType,
		MatchType:  data.MatchType,
	}, nil
}

// ExtractPkBattleEnd 从原始 JSON 中提取 PK 结束信息（PK_BATTLE_END，battle_type=2）
func ExtractPkBattleEnd(raw string) (*PkBattleEndInfo, error) {
	var outer struct {
		PkID      int64           `json:"pk_id"`
		PkStatus  int64           `json:"pk_status"`
		Timestamp int64           `json:"timestamp"`
		SendTime  int64           `json:"send_time"`
		Data      json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &outer); err != nil {
		return nil, fmt.Errorf("解析 PK_BATTLE_END 外层消息失败: %w", err)
	}
	var data struct {
		BattleType int64       `json:"battle_type"`
		InitInfo   *PkSideInfo `json:"init_info"`
		MatchInfo  *PkSideInfo `json:"match_info"`
		Timer      int64       `json:"timer"`
		ShowStreak bool        `json:"show_streak"`
	}
	if err := json.Unmarshal(outer.Data, &data); err != nil {
		return nil, fmt.Errorf("解析 PK_BATTLE_END data 字段失败: %w", err)
	}
	return &PkBattleEndInfo{
		PkID:       outer.PkID,
		PkStatus:   outer.PkStatus,
		Timestamp:  outer.Timestamp,
		SendTime:   outer.SendTime,
		BattleType: data.BattleType,
		InitInfo:   data.InitInfo,
		MatchInfo:  data.MatchInfo,
		Timer:      data.Timer,
		ShowStreak: data.ShowStreak,
	}, nil
}

// ExtractPkBattleSettleNew 从原始 JSON 中提取 PK 结算信息（PK_BATTLE_SETTLE_NEW，battle_type=6）
func ExtractPkBattleSettleNew(raw string) (*PkBattleSettleNewInfo, error) {
	var outer struct {
		PkID      int64           `json:"pk_id"`
		PkStatus  int64           `json:"pk_status"`
		Timestamp int64           `json:"timestamp"`
		SendTime  int64           `json:"send_time"`
		Data      json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &outer); err != nil {
		return nil, fmt.Errorf("解析 PK_BATTLE_SETTLE_NEW 外层消息失败: %w", err)
	}
	var data struct {
		BattleType    int64       `json:"battle_type"`
		InitInfo      *PkSideInfo `json:"init_info"`
		MatchInfo     *PkSideInfo `json:"match_info"`
		PunishName    string      `json:"punish_name"`
		PunishEndTime int64       `json:"punish_end_time"`
		SettleStatus  int64       `json:"settle_status"`
		DmScore       int64       `json:"dmscore"`
	}
	if err := json.Unmarshal(outer.Data, &data); err != nil {
		return nil, fmt.Errorf("解析 PK_BATTLE_SETTLE_NEW data 字段失败: %w", err)
	}
	return &PkBattleSettleNewInfo{
		PkID:          outer.PkID,
		PkStatus:      outer.PkStatus,
		Timestamp:     outer.Timestamp,
		SendTime:      outer.SendTime,
		BattleType:    data.BattleType,
		InitInfo:      data.InitInfo,
		MatchInfo:     data.MatchInfo,
		PunishName:    data.PunishName,
		PunishEndTime: data.PunishEndTime,
		SettleStatus:  data.SettleStatus,
		DmScore:       data.DmScore,
	}, nil
}

// ExtractInteractWordV2 从原始 JSON 中提取 protobuf 编码的用户互动信息(INTERACT_WORD_V2)
func ExtractInteractWordV2(raw string) (*InteractWordV2Info, error) {
	// JSON 解析外层，提取 data.pb 字段
	var outer struct {
		Data struct {
			Pb string `json:"pb"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &outer); err != nil {
		return nil, fmt.Errorf("解析 INTERACT_WORD_V2 外层消息失败: %w", err)
	}
	if outer.Data.Pb == "" {
		return nil, fmt.Errorf("INTERACT_WORD_V2 data.pb 字段为空")
	}
	// Base64 解码 protobuf 二进制
	pbBytes, err := base64.StdEncoding.DecodeString(outer.Data.Pb)
	if err != nil {
		return nil, fmt.Errorf("INTERACT_WORD_V2 base64 解码失败: %w", err)
	}
	// Protobuf 反序列化
	var pb interactwordv2.InteractWordV2
	if err := proto.Unmarshal(pbBytes, &pb); err != nil {
		return nil, fmt.Errorf("INTERACT_WORD_V2 protobuf 反序列化失败: %w", err)
	}
	// 映射到 InteractWordV2Info
	result := &InteractWordV2Info{
		UID:       int64(pb.Uid),
		Uname:     pb.Uname,
		MsgType:   int64(pb.MsgType),
		RoomID:    int64(pb.Roomid),
		Timestamp: int64(pb.Timestamp),
	}
	if pb.FansMedal != nil {
		result.BadgeUID = pb.FansMedal.TargetId
		result.BadgeName = pb.FansMedal.MedalName
		result.BadgeLevel = pb.FansMedal.MedalLevel
		result.BadgeType = pb.FansMedal.GuardLevel
	}
	return result, nil
}

// ExtractGuardBuy 从原始 JSON 中提取大航海（舰长/提督/总督）购买信息
func ExtractGuardBuy(raw string) (*GuardBuyInfo, error) {
	// 解析外层 JSON，拿到 data 子对象
	var outer struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &outer); err != nil {
		return nil, fmt.Errorf("解析 GUARD_BUY 外层消息失败: %w", err)
	}
	var info GuardBuyInfo
	if err := json.Unmarshal(outer.Data, &info); err != nil {
		return nil, fmt.Errorf("解析 GUARD_BUY data 字段失败: %w", err)
	}
	info.Price = info.Price / 10
	return &info, nil
}

// ExtractSendGiftV2 从原始 JSON 中提取 protobuf 编码的送礼信息(SEND_GIFT_V2)
//
// SEND_GIFT_V2 的 data.pb 字段是 base64 编码的 protobuf 二进制数据，并将结果映射到与 SEND_GIFT 相同的 SendGiftInfo 结构体。
func ExtractSendGiftV2(raw string) (*SendGiftInfo, error) {
	// JSON 解析外层，提取 data.pb 字段
	var outer struct {
		Data struct {
			Pb string `json:"pb"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &outer); err != nil {
		return nil, fmt.Errorf("解析 SEND_GIFT_V2 外层消息失败: %w", err)
	}
	if outer.Data.Pb == "" {
		return nil, fmt.Errorf("SEND_GIFT_V2 data.pb 字段为空")
	}
	// Base64 解码 protobuf 二进制
	pbBytes, err := base64.StdEncoding.DecodeString(outer.Data.Pb)
	if err != nil {
		return nil, fmt.Errorf("SEND_GIFT_V2 base64 解码失败: %w", err)
	}
	// Protobuf 反序列化
	var pb sendgiftv2.SendGiftV2
	if err := proto.Unmarshal(pbBytes, &pb); err != nil {
		return nil, fmt.Errorf("SEND_GIFT_V2 protobuf 反序列化失败: %w", err)
	}
	// 映射到 SendGiftInfo
	result := &SendGiftInfo{
		UID:   pb.Uid,
		Uname: pb.Uname,
	}
	// 礼物明细 — 盲盒爆出时一次广播会携带多份礼物，必须全部保留
	for _, item := range pb.GiftList {
		// 实际消耗以 total_coin 为准；该字段缺失时退化为标价总额
		totalCoin := item.TotalCoin
		if totalCoin == 0 {
			totalCoin = item.Price * item.Num
		}
		result.Gifts = append(result.Gifts, GiftItemInfo{
			GiftID:    item.GiftId,
			GiftName:  item.GiftName,
			Num:       item.Num,
			Price:     item.Price / 10,
			TotalCoin: totalCoin / 10,
		})
		// 接受礼物的主播ID
		if result.AnchorID == 0 && item.ReceiveUserInfo != nil {
			result.AnchorID = item.ReceiveUserInfo.Uid
		}
	}
	// 勋章信息 — 用户未佩戴勋章时为 null
	if pb.SenderUinfo != nil && pb.SenderUinfo.Medal != nil {
		m := pb.SenderUinfo.Medal
		result.BadgeUID = m.Ruid
		result.BadgeName = m.Name
		result.BadgeLevel = m.Level
		result.BadgeType = m.GuardLevel
	}
	// 盲盒礼物 - 非盲盒礼物时为 null
	if pb.BlindGift != nil {
		result.BlindGift = &BlindGiftInfo{
			GiftAction:        pb.BlindGift.GiftAction,
			GiftTipPrice:      pb.BlindGift.GiftTipPrice / 10,
			OriginalGiftID:    pb.BlindGift.OriginalGiftId,
			OriginalGiftName:  pb.BlindGift.OriginalGiftName,
			OriginalGiftPrice: pb.BlindGift.OriginalGiftPrice / 10,
		}
	}
	return result, nil
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
		UID           int64  `json:"uid"`
		Uname         string `json:"uname"`
		GiftID        int64  `json:"giftId"`
		GiftName      string `json:"giftName"`
		Price         int64  `json:"price"`
		Num           int64  `json:"num"`
		TotalCoin     int64  `json:"total_coin"`
		ReceiverUinfo struct {
			UID int64 `json:"uid"`
		} `json:"receiver_uinfo"`
		SenderUinfo struct {
			Medal *struct {
				RUID       int64  `json:"ruid"`
				Name       string `json:"name"`
				GuardLevel int64  `json:"guard_level"`
				Level      int64  `json:"level"`
			} `json:"medal"`
		} `json:"sender_uinfo"`
		BlindGift json.RawMessage `json:"blind_gift"`
	}
	if err := json.Unmarshal(outer.Data, &data); err != nil {
		return nil, fmt.Errorf("解析 SEND_GIFT data 字段失败: %w", err)
	}
	// 实际消耗以 total_coin 为准；该字段缺失时退化为标价总额
	totalCoin := data.TotalCoin
	if totalCoin == 0 {
		totalCoin = data.Price * data.Num
	}
	result := &SendGiftInfo{
		UID:      data.UID,
		Uname:    data.Uname,
		AnchorID: data.ReceiverUinfo.UID,
		Gifts: []GiftItemInfo{{
			GiftID:    data.GiftID,
			GiftName:  data.GiftName,
			Num:       data.Num,
			Price:     data.Price / 10,
			TotalCoin: totalCoin / 10,
		}},
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
			GiftAction        string `json:"gift_action"`
			GiftTipPrice      int64  `json:"gift_tip_price"`
			OriginalGiftID    int64  `json:"original_gift_id"`
			OriginalGiftName  string `json:"original_gift_name"`
			OriginalGiftPrice int64  `json:"original_gift_price"`
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
	if info.LiveTime == 0 {
		return nil, fmt.Errorf("LIVE 消息缺少 live_time 字段 (B站重复推送)")
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
