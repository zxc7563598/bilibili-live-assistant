// Package live 提供 B站 直播间的实时交互能力：WebSocket 消息监听和弹幕发送队列
//
// 本包与 room 包的区别：
//   - room 包：HTTP API 调用（获取直播间信息、发送弹幕接口等）
//   - live 包：长连接实时操作（WebSocket 监听、弹幕队列发送）
package live

import "encoding/json"

// Cmd 是 B站 直播 WebSocket 协议的消息命令类型
type Cmd string

// B站 直播 WebSocket 常见命令字
const (
	CmdLiveStart    Cmd = "LIVE"              // 直播开始
	CmdLiveCutOff   Cmd = "CUT_OFF"           // 直播被超管切断
	CmdLiveRoomLock Cmd = "ROOM_LOCK"         // 直播间被封
	CmdLiveEnd      Cmd = "PREPARING"         // 直播结束（下播）
	CmdSendGift     Cmd = "SEND_GIFT"         // 送礼消息
	CmdSendGiftV2   Cmd = "SEND_GIFT_V2"      // 送礼消息V2
	CmdGuardBuy     Cmd = "GUARD_BUY"         // 大航海（舰长/提督/总督）购买
	CmdInteractWord Cmd = "INTERACT_WORD_V2"  // 用户互动（关注、分享等）
	CmdDanmuMsg     Cmd = "DANMU_MSG"         // 弹幕消息
	CmdPkStart      Cmd = "PK_BATTLE_PRE_NEW" // PK即将开始
)

// Message 表示一条从直播间 WebSocket 收到的消息
//
// Raw 保留原始 JSON 数据，调用方可根据 Cmd 类型自行反序列化到具体结构体
type Message struct {
	Cmd Cmd             `json:"cmd"` // 消息命令字
	Raw json.RawMessage `json:"-"`   // 原始 JSON（不参与序列化，仅用于数据透传）
}
