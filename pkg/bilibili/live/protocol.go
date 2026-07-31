package live

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
)

// =========================================================================
// B站 直播间 WebSocket 二进制协议
// =========================================================================
//
// 帧格式（大端字节序）：
//   - 4 bytes: packet_len    总长度（包含帧头）
//   - 2 bytes: header_len    帧头长度（固定 16）
//   - 2 bytes: protocol_ver  协议版本
//   - 4 bytes: opcode        操作码
//   - 4 bytes: sequence      序列号（恒为 1）
//   - N bytes: payload       消息体
//
// 协议版本（protocol_ver）：
//   0 — 普通包（正文不使用压缩）
//   1 — 心跳及认证包（正文不使用压缩）
//   2 — 普通包（正文使用 zlib 压缩）
//   3 — 普通包（正文使用 brotli 压缩，可能含多个子包）
//
// 操作码（opcode）：
//   2 — 心跳包（客户端→服务端）
//   3 — 心跳包回复（服务端→客户端，payload 为人气值）
//   5 — 普通包（服务端→客户端，弹幕/礼物/互动等）
//   7 — 认证包（客户端→服务端）
//   8 — 认证包回复（服务端→客户端）
//
// 参考：
//   https://hejunjie.life/blog/e1ccd148

const (
	// headerLen B站协议帧头固定长度
	headerLen = 16
	// 操作码
	opHeartbeat      = 2 // 心跳包
	opHeartbeatReply = 3 // 心跳包回复（人气值）
	opMessage        = 5 // 普通包（命令/数据）
	opAuth           = 7 // 认证包
	opAuthReply      = 8 // 认证包回复
	// 协议版本
	protoVerJSON   = 0 // 普通包（无压缩）
	protoVerHB     = 1 // 心跳及认证包（无压缩）
	protoVerZlib   = 2 // zlib 压缩
	protoVerBrotli = 3 // brotli 压缩
)

// rawPacket 是解包后的原始帧数据
type rawPacket struct {
	PacketLen   int    // 总长度（含帧头）
	HeaderLen   int    // 帧头长度（固定 16）
	ProtocolVer int    // 协议版本
	Opcode      int    // 操作码
	Sequence    int    // 序列号
	Payload     []byte // 消息体
}

// =========================================================================
// 帧打包（客户端→服务端）
// =========================================================================

// buildFrame 按 B站 协议打包二进制帧
// version: 协议版本（认证/心跳用 1，普通消息用 0）
// opcode: 操作码（认证=7, 心跳=2）
// payload: 消息体
func buildFrame(version int, opcode int, payload []byte) []byte {
	packetLen := headerLen + len(payload)
	buf := make([]byte, packetLen)

	binary.BigEndian.PutUint32(buf[0:4], uint32(packetLen)) // 总长度
	binary.BigEndian.PutUint16(buf[4:6], headerLen)         // 帧头长度
	binary.BigEndian.PutUint16(buf[6:8], uint16(version))   // 协议版本
	binary.BigEndian.PutUint32(buf[8:12], uint32(opcode))   // 操作码
	binary.BigEndian.PutUint32(buf[12:16], 1)               // 序列号
	copy(buf[16:], payload)

	return buf
}

// authPayload 是认证帧 JSON 结构
type authPayload struct {
	UID      int64  `json:"uid"`
	RoomID   int64  `json:"roomid"`
	ProtoVer int    `json:"protover"`
	Buvid    string `json:"buvid"`
	Platform string `json:"platform"`
	Type     int    `json:"type"`
	Key      string `json:"key"`
}

// BuildAuthFrame 构建认证帧
func BuildAuthFrame(roomID int64, uid int64, buvid string, token string) ([]byte, error) {
	payload, err := json.Marshal(authPayload{
		UID:      uid,
		RoomID:   roomID,
		ProtoVer: protoVerBrotli,
		Buvid:    buvid,
		Platform: "web",
		Type:     2,
		Key:      token,
	})
	if err != nil {
		return nil, fmt.Errorf("llive: 将认证参数转换为 JSON 时出错: %w", err)
	}
	return buildFrame(protoVerHB, opAuth, payload), nil
}

// BuildHeartbeatFrame 构建心跳帧
func BuildHeartbeatFrame() []byte {
	payload := []byte("[object Object]")
	return buildFrame(protoVerHB, opHeartbeat, payload)
}

// =========================================================================
// 帧解包（服务端→客户端）
// =========================================================================

// unpackFrame 从原始字节中解出帧头与负载
func unpackFrame(data []byte) (*rawPacket, error) {
	if len(data) < headerLen {
		return nil, fmt.Errorf("live: packet too short: %d bytes", len(data))
	}

	p := &rawPacket{
		PacketLen:   int(binary.BigEndian.Uint32(data[0:4])),
		HeaderLen:   int(binary.BigEndian.Uint16(data[4:6])),
		ProtocolVer: int(binary.BigEndian.Uint16(data[6:8])),
		Opcode:      int(binary.BigEndian.Uint32(data[8:12])),
		Sequence:    int(binary.BigEndian.Uint32(data[12:16])),
	}

	if len(data) >= p.PacketLen {
		p.Payload = data[p.HeaderLen:p.PacketLen]
	}

	return p, nil
}

// ParseResponse 解析服务端响应的原始帧数据，提取业务消息列表
//
// 处理逻辑：
//   - 解包帧头，获取协议版本和操作码
//   - 操作码 3（心跳回复）：忽略，返回空列表
//   - 操作码 8（认证回复）：忽略，返回空列表
//   - 操作码 5（服务端消息）：根据协议版本解压并解析子包
//
// 返回的 Message 列表中，每条消息的 Raw 字段保留原始 payload JSON
func ParseResponse(raw []byte) ([]*Message, error) {
	packet, err := unpackFrame(raw)
	if err != nil {
		return nil, err
	}
	switch packet.Opcode {
	case opHeartbeatReply:
		// 心跳回复（payload 为人气值），不需要投递给调用方
		return nil, nil
	case opAuthReply:
		// 认证回复，不需要投递给调用方
		return nil, nil
	case opMessage:
		return parseMessagePayload(packet.ProtocolVer, packet.Payload)
	default:
		// 未知操作码，跳过
		return nil, nil
	}
}

// =========================================================================
// 消息解压与解析
// =========================================================================

// parseMessagePayload 根据协议版本解压并解析消息体
//
//   - protover 0: 直接 JSON 解析
//   - protover 1: 心跳/认证（不会走到这里，已在 ParseResponse 处理）
//   - protover 2: zlib 解压 + 递归解包子包
//   - protover 3: brotli 解压 + 递归解包子包
func parseMessagePayload(protoVer int, payload []byte) ([]*Message, error) {
	switch protoVer {
	case protoVerJSON:
		// 无压缩，直接解析 JSON
		return parseJSONPayload(payload)
	case protoVerZlib:
		// zlib 解压后再解析
		decompressed, err := zlibDecompress(payload)
		if err != nil {
			return nil, fmt.Errorf("live: 解压 zlib 压缩数据时发生错误: %w", err)
		}
		return unpackNestedPackets(decompressed)
	case protoVerBrotli:
		// brotli 解压后再解析
		decompressed, err := brotliDecompress(payload)
		if err != nil {
			return nil, fmt.Errorf("live: 解压 brotli 压缩数据时发生错误: %w", err)
		}
		return unpackNestedPackets(decompressed)

	default:
		return nil, fmt.Errorf("live: 协议版本 %d 无法识别", protoVer)
	}
}

// unpackNestedPackets 解压后的数据可能包含多个子包，逐个解包并 JSON 解析
func unpackNestedPackets(data []byte) ([]*Message, error) {
	var messages []*Message
	offset := 0

	for offset < len(data) {
		if offset+headerLen > len(data) {
			break
		}

		sub, err := unpackFrame(data[offset:])
		if err != nil {
			return messages, err
		}
		if sub.PacketLen <= 0 {
			break
		}

		// 递归解析子包的 payload（子包可能是 protover 0 JSON）
		subMsgs, _ := parseMessagePayload(sub.ProtocolVer, sub.Payload)
		messages = append(messages, subMsgs...)

		offset += sub.PacketLen
	}

	return messages, nil
}

// parseJSONPayload 直接 JSON 解析消息体，提取 cmd 字段
func parseJSONPayload(payload []byte) ([]*Message, error) {
	var msg Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, err
	}
	msg.Raw = payload
	return []*Message{&msg}, nil
}

// =========================================================================
// 压缩/解压工具
// =========================================================================

// zlibDecompress zlib 解压
func zlibDecompress(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

// brotliDecompress brotli 解压
func brotliDecompress(data []byte) ([]byte, error) {
	r := brotli.NewReader(bytes.NewReader(data))
	return io.ReadAll(r)
}
